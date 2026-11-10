package handler

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"whisper/backend/internal/config"
	"whisper/backend/internal/middleware"
	"whisper/backend/internal/model"
	"whisper/backend/internal/realtime"
	"whisper/backend/internal/repository"
	"whisper/backend/internal/service"
)

const sessionCookieName = "whisper_session"

type API struct {
	auth           *service.AuthService
	conversations  *service.ConversationService
	messages       *service.MessageService
	frontendOrigin string
	secureCookies  bool
	websocket      *realtime.Handler
	loginLimiter   *middleware.FixedWindowLimiter
	createLimiter  *middleware.FixedWindowLimiter
	messageLimiter *middleware.FixedWindowLimiter
}

func NewAPI(auth *service.AuthService, conversations *service.ConversationService, messages *service.MessageService, websocket *realtime.Handler, cfg config.Config) *API {
	return &API{auth: auth, conversations: conversations, messages: messages, websocket: websocket, frontendOrigin: cfg.FrontendOrigin, secureCookies: cfg.SecureCookies, loginLimiter: middleware.NewFixedWindowLimiter(8, time.Minute), createLimiter: middleware.NewFixedWindowLimiter(12, time.Minute), messageLimiter: middleware.NewFixedWindowLimiter(60, time.Minute)}
}

func (a *API) RegisterRoutes(router *gin.Engine) {
	api := router.Group("/api/v1")
	api.POST("/admin/auth/login", a.requireFrontendOrigin(), a.loginLimiter.Middleware(), a.login)
	api.POST("/admin/auth/logout", a.requireFrontendOrigin(), a.requireAdmin(), a.logout)
	api.GET("/admin/me", a.requireAdmin(), a.me)
	api.POST("/public/:slug/conversations", a.createLimiter.Middleware(), a.createConversation)
	api.GET("/conversations/:id", a.getConversation)
	api.POST("/conversations/:id/messages", a.messageLimiter.Middleware(), a.createVisitorMessage)
	api.GET("/conversations/:id/messages", a.listVisitorMessages)
	api.GET("/ws", gin.WrapH(a.websocket))
	api.POST("/conversations/:id/messages/read", a.markVisitorRead)
	admin := api.Group("/admin", a.requireAdmin())
	admin.GET("/conversations", a.listAdminConversations)
	admin.POST("/conversations/:id/messages", a.createAdminMessage)
	admin.GET("/conversations/:id/messages", a.listAdminMessages)
	admin.POST("/conversations/:id/messages/read", a.markAdminRead)
}

func (a *API) listAdminConversations(c *gin.Context) {
	admin, _ := c.Get("admin")
	items, err := a.conversations.ListForAdmin(c.Request.Context(), admin.(model.AdminUser).ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not load conversations"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"conversations": items})
}

func (a *API) CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if origin == a.frontendOrigin {
			c.Header("Access-Control-Allow-Origin", a.frontendOrigin)
			c.Header("Access-Control-Allow-Credentials", "true")
			c.Header("Access-Control-Allow-Headers", "Authorization, Content-Type")
			c.Header("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			c.Header("Vary", "Origin")
		}
		if c.Request.Method == http.MethodOptions {
			if origin != a.frontendOrigin {
				c.AbortWithStatus(http.StatusForbidden)
				return
			}
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}

func (a *API) createVisitorMessage(c *gin.Context) {
	token, ok := bearerToken(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "conversation access token required"})
		return
	}
	var req messageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid message request"})
		return
	}
	m, err := a.messages.Create(c.Request.Context(), c.Param("id"), "anonymous", req.Content, req.ClientMessageID, "", token)
	a.writeMessageResult(c, m, err)
}

func (a *API) createAdminMessage(c *gin.Context) {
	var req messageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid message request"})
		return
	}
	admin, _ := c.Get("admin")
	user := admin.(model.AdminUser)
	m, err := a.messages.Create(c.Request.Context(), c.Param("id"), "admin", req.Content, req.ClientMessageID, user.ID, "")
	a.writeMessageResult(c, m, err)
}

type messageRequest struct {
	Content         string  `json:"content"`
	ClientMessageID *string `json:"client_message_id"`
}

func (a *API) writeMessageResult(c *gin.Context, m model.Message, err error) {
	if errors.Is(err, service.ErrInvalidMessage) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "message must contain 1-5000 characters"})
		return
	}
	if errors.Is(err, repository.ErrNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": "conversation not found or unavailable"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not save message"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": m})
}

func (a *API) listVisitorMessages(c *gin.Context) {
	token, ok := bearerToken(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "conversation access token required"})
		return
	}
	a.listMessages(c, "", token, false)
}
func (a *API) listAdminMessages(c *gin.Context) {
	admin, _ := c.Get("admin")
	a.listMessages(c, admin.(model.AdminUser).ID, "", true)
}
func (a *API) listMessages(c *gin.Context, adminID, token string, isAdmin bool) {
	limit := 50
	if raw := c.Query("limit"); raw != "" {
		n, err := strconv.Atoi(raw)
		if err != nil || n < 1 || n > 100 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "limit must be between 1 and 100"})
			return
		}
		limit = n
	}
	items, err := a.messages.List(c.Request.Context(), c.Param("id"), limit, c.Query("before"), adminID, token, isAdmin)
	if errors.Is(err, repository.ErrNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": "conversation not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not load messages"})
		return
	}
	var next any
	if len(items) == limit {
		next = items[0].ID
	}
	c.JSON(http.StatusOK, gin.H{"messages": items, "next_before": next})
}

func (a *API) markVisitorRead(c *gin.Context) {
	token, ok := bearerToken(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "conversation access token required"})
		return
	}
	a.markRead(c, "visitor", "", token)
}
func (a *API) markAdminRead(c *gin.Context) {
	admin, _ := c.Get("admin")
	a.markRead(c, "admin", admin.(model.AdminUser).ID, "")
}
func (a *API) markRead(c *gin.Context, reader, adminID, token string) {
	n, err := a.messages.MarkRead(c.Request.Context(), c.Param("id"), reader, adminID, token)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not mark messages read"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"updated": n})
}

func bearerToken(c *gin.Context) (string, bool) {
	value := c.GetHeader("Authorization")
	token, ok := strings.CutPrefix(value, "Bearer ")
	return token, ok && token != ""
}

func (a *API) requireFrontendOrigin() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.GetHeader("Origin") != a.frontendOrigin {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "request origin is not allowed"})
			return
		}
		c.Next()
	}
}

func (a *API) requireAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		cookie, err := c.Cookie(sessionCookieName)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
			return
		}
		admin, err := a.auth.ValidateSession(c.Request.Context(), cookie)
		if errors.Is(err, repository.ErrNotFound) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
			return
		}
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "could not validate session"})
			return
		}
		c.Set("admin", admin)
		c.Set("session_token", cookie)
		c.Next()
	}
}

func (a *API) login(c *gin.Context) {
	var request struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&request); err != nil || len(request.Username) < 3 || len(request.Username) > 50 || len(request.Password) < 1 || len(request.Password) > 72 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "username or password is invalid"})
		return
	}
	admin, token, expiresAt, err := a.auth.Login(c.Request.Context(), request.Username, request.Password)
	if errors.Is(err, service.ErrInvalidCredentials) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid username or password"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not create session"})
		return
	}
	maxAge := int(time.Until(expiresAt).Seconds())
	http.SetCookie(c.Writer, &http.Cookie{
		Name: sessionCookieName, Value: token, Path: "/api/v1", HttpOnly: true,
		Secure: a.secureCookies, SameSite: a.sessionSameSite(), MaxAge: maxAge, Expires: expiresAt,
	})
	c.JSON(http.StatusOK, gin.H{"admin": admin})
}

func (a *API) logout(c *gin.Context) {
	token, _ := c.Get("session_token")
	if raw, ok := token.(string); ok {
		if err := a.auth.Logout(c.Request.Context(), raw); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "could not end session"})
			return
		}
	}
	http.SetCookie(c.Writer, &http.Cookie{Name: sessionCookieName, Value: "", Path: "/api/v1", HttpOnly: true, Secure: a.secureCookies, SameSite: a.sessionSameSite(), MaxAge: -1, Expires: time.Unix(1, 0)})
	c.Status(http.StatusNoContent)
}

func (a *API) sessionSameSite() http.SameSite {
	if a.secureCookies {
		return http.SameSiteNoneMode
	}
	return http.SameSiteLaxMode
}

func (a *API) me(c *gin.Context) {
	admin, _ := c.Get("admin")
	c.JSON(http.StatusOK, gin.H{"admin": admin})
}

func (a *API) createConversation(c *gin.Context) {
	conversation, token, err := a.conversations.Create(c.Request.Context(), c.Param("slug"))
	if errors.Is(err, service.ErrInvalidSlug) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "profile slug is invalid"})
		return
	}
	if errors.Is(err, repository.ErrNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": "profile not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not create conversation"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{
		"conversation":     conversation,
		"access_token":     token,
		"conversation_url": a.frontendOrigin + "/c/" + conversation.ID + "#k=" + token,
	})
}

func (a *API) getConversation(c *gin.Context) {
	value := c.GetHeader("Authorization")
	if !strings.HasPrefix(value, "Bearer ") {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "conversation access token required"})
		return
	}
	conversation, err := a.conversations.GetAuthorized(c.Request.Context(), c.Param("id"), strings.TrimPrefix(value, "Bearer "))
	if errors.Is(err, repository.ErrNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": "conversation not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not load conversation"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"conversation": conversation})
}
