package model

import "time"

type AdminUser struct {
	ID       string `json:"id"`
	Username string `json:"username"`
}

type Conversation struct {
	ID        string    `json:"id"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at,omitempty"`
	LastMessage string `json:"last_message,omitempty"`
	UnreadCount int `json:"unread_count,omitempty"`
}

type Message struct {
	ID              string     `json:"id"`
	ConversationID  string     `json:"conversation_id"`
	SenderType      string     `json:"sender_type"`
	Content         string     `json:"content"`
	ClientMessageID *string    `json:"client_message_id,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	ReadAt          *time.Time `json:"read_at"`
}
