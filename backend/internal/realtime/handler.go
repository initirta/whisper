package realtime

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
	"whisper/backend/internal/middleware"
	"whisper/backend/internal/model"
	"whisper/backend/internal/service"
)

type event struct {
	Type            string          `json:"type"`
	ConversationID  string          `json:"conversation_id,omitempty"`
	Message         *model.Message  `json:"message,omitempty"`
	Messages        []model.Message `json:"messages,omitempty"`
	Token           string          `json:"token,omitempty"`
	After           string          `json:"after,omitempty"`
	Content         string          `json:"content,omitempty"`
	ClientMessageID *string         `json:"client_message_id,omitempty"`
	Error           string          `json:"error,omitempty"`
	Updated         int64           `json:"updated,omitempty"`
}
type peer struct {
	conn                       *websocket.Conn
	send                       chan event
	room, role, adminID, token string
}
type Handler struct {
	auth           *service.AuthService
	conversations  *service.ConversationService
	messages       *service.MessageService
	origin         string
	mu             sync.RWMutex
	rooms          map[string]map[*peer]bool
	messageLimiter *middleware.FixedWindowLimiter
}

func NewHandler(a *service.AuthService, c *service.ConversationService, m *service.MessageService, origin string) *Handler {
	return &Handler{auth: a, conversations: c, messages: m, origin: origin, rooms: make(map[string]map[*peer]bool), messageLimiter: middleware.NewFixedWindowLimiter(60, time.Minute)}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Origin") != h.origin {
		http.Error(w, "origin not allowed", http.StatusForbidden)
		return
	}
	c, err := websocket.Accept(w, r, &websocket.AcceptOptions{OriginPatterns: []string{h.origin}, InsecureSkipVerify: false})
	if err != nil {
		return
	}
	defer c.Close(websocket.StatusNormalClosure, "")
	// 5,000 Unicode characters can require up to 20 KB before JSON framing.
	c.SetReadLimit(32 << 10)
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	var first event
	err = wsjson.Read(ctx, c, &first)
	cancel()
	if err != nil || first.Type != "authenticate" {
		_ = wsjson.Write(r.Context(), c, event{Type: "auth.failed", Error: "authenticate event required"})
		return
	}
	room := r.URL.Query().Get("conversation_id")
	if room == "" {
		_ = wsjson.Write(r.Context(), c, event{Type: "auth.failed", Error: "conversation_id required"})
		return
	}
	p := &peer{conn: c, send: make(chan event, 32), room: room}
	// Prefer an explicitly supplied visitor token. Browsers may also carry the
	// owner's admin cookie when they open a visitor link in the same profile.
	if first.Token != "" {
		if _, e := h.conversations.GetAuthorized(r.Context(), room, first.Token); e != nil {
			_ = wsjson.Write(r.Context(), c, event{Type: "auth.failed", Error: "conversation authorization failed"})
			return
		}
		p.role = "anonymous"
		p.token = first.Token
	} else if cookie, e := r.Cookie("whisper_session"); e == nil {
		if admin, e := h.auth.ValidateSession(r.Context(), cookie.Value); e == nil {
			if _, e = h.conversations.GetAdminAuthorized(r.Context(), room, admin.ID); e == nil {
				p.role = "admin"
				p.adminID = admin.ID
			}
		}
	}
	if p.role == "" {
		_ = wsjson.Write(r.Context(), c, event{Type: "auth.failed", Error: "valid session or visitor token required"})
		return
	}
	h.add(p)
	defer func() {
		h.remove(p)
		h.broadcast(room, event{Type: "presence.update", ConversationID: room, Content: "offline"}, nil)
	}()
	writeCtx, stop := context.WithCancel(r.Context())
	defer stop()
	go h.writeLoop(writeCtx, p)
	p.send <- event{Type: "auth.success", ConversationID: room}
	listCtx, done := context.WithTimeout(r.Context(), 5*time.Second)
	items, e := h.messages.List(listCtx, room, 100, "", p.adminID, p.token, p.role == "admin")
	done()
	if e == nil {
		p.send <- event{Type: "message.sync", Messages: items}
	}
	h.broadcast(room, event{Type: "presence.update", ConversationID: room, Content: "online"}, p)
	go h.heartbeat(writeCtx, p)
	for {
		var in event
		if err := wsjson.Read(r.Context(), c, &in); err != nil {
			return
		}
		switch in.Type {
		case "message.send":
			if ok, _ := h.messageLimiter.Allow(room, time.Now()); !ok {
				p.send <- event{Type: "error", Error: "message rate limit reached; try again shortly"}
				continue
			}
			m, e := h.messages.Create(r.Context(), room, p.role, in.Content, in.ClientMessageID, p.adminID, p.token)
			if e != nil {
				p.send <- event{Type: "error", Error: "message could not be saved"}
				continue
			}
			p.send <- event{Type: "message.ack", Message: &m}
			h.broadcast(room, event{Type: "message.new", Message: &m}, nil)
		case "message.read":
			n, e := h.messages.MarkRead(r.Context(), room, p.role, p.adminID, p.token)
			if e == nil {
				h.broadcast(room, event{Type: "message.read", Updated: n}, nil)
			}
		case "typing.start", "typing.stop":
			h.broadcast(room, event{Type: in.Type, ConversationID: room}, p)
		case "ping":
			p.send <- event{Type: "pong"}
		default:
			p.send <- event{Type: "error", Error: "unknown event"}
		}
	}
}

func (h *Handler) add(p *peer) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.rooms[p.room] == nil {
		h.rooms[p.room] = make(map[*peer]bool)
	}
	h.rooms[p.room][p] = true
}
func (h *Handler) remove(p *peer) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.rooms[p.room], p)
	if len(h.rooms[p.room]) == 0 {
		delete(h.rooms, p.room)
	}
	close(p.send)
}
func (h *Handler) broadcast(room string, e event, exclude *peer) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for p := range h.rooms[room] {
		if p != exclude {
			select {
			case p.send <- e:
			default:
			}
		}
	}
}
func (h *Handler) writeLoop(ctx context.Context, p *peer) {
	for {
		select {
		case <-ctx.Done():
			return
		case e, ok := <-p.send:
			if !ok {
				return
			}
			wctx, cancel := context.WithTimeout(ctx, 5*time.Second)
			err := wsjson.Write(wctx, p.conn, e)
			cancel()
			if err != nil {
				return
			}
		}
	}
}
func (h *Handler) heartbeat(ctx context.Context, p *peer) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
			err := p.conn.Ping(pingCtx)
			cancel()
			if err != nil {
				return
			}
		}
	}
}
