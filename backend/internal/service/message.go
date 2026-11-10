package service

import (
	"context"
	"errors"
	"regexp"
	"strings"

	"whisper/backend/internal/model"
	"whisper/backend/internal/repository"
)

var ErrInvalidMessage = errors.New("message is invalid")
var validMessageID = regexp.MustCompile(`(?i)^[0-9a-f]{8}-[0-9a-f]{4}-[1-8][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)

type MessageService struct{ messages *repository.MessageRepository }

func NewMessageService(messages *repository.MessageRepository) *MessageService {
	return &MessageService{messages: messages}
}
func (s *MessageService) Create(ctx context.Context, id, sender, content string, clientID *string, adminID, token string) (model.Message, error) {
	content = strings.TrimSpace(content)
	if len([]rune(content)) < 1 || len([]rune(content)) > 5000 {
		return model.Message{}, ErrInvalidMessage
	}
	if clientID != nil && !validMessageID.MatchString(*clientID) {
		return model.Message{}, ErrInvalidMessage
	}
	return s.messages.Create(ctx, id, sender, content, clientID, adminID, hashToken(token))
}
func (s *MessageService) List(ctx context.Context, id string, limit int, before, adminID, token string, admin bool) ([]model.Message, error) {
	if limit < 1 || limit > 100 {
		limit = 50
	}
	return s.messages.List(ctx, id, limit, before, adminID, hashToken(token), admin)
}
func (s *MessageService) MarkRead(ctx context.Context, id, reader, adminID, token string) (int64, error) {
	return s.messages.MarkRead(ctx, id, reader, adminID, hashToken(token))
}
