package service

import (
	"context"
	"errors"
	"regexp"

	"whisper/backend/internal/model"
	"whisper/backend/internal/repository"
)

var validUsername = regexp.MustCompile(`^[A-Za-z0-9_]{3,50}$`)
var ErrInvalidSlug = errors.New("invalid profile slug")

type ConversationService struct {
	admins        *repository.AdminRepository
	conversations *repository.ConversationRepository
}

func NewConversationService(admins *repository.AdminRepository, conversations *repository.ConversationRepository) *ConversationService {
	return &ConversationService{admins: admins, conversations: conversations}
}

func (s *ConversationService) Create(ctx context.Context, slug string) (model.Conversation, string, error) {
	if !validUsername.MatchString(slug) {
		return model.Conversation{}, "", ErrInvalidSlug
	}
	adminID, err := s.conversations.AdminIDBySlug(ctx, slug)
	if err != nil {
		return model.Conversation{}, "", err
	}
	token, err := newAccessToken()
	if err != nil {
		return model.Conversation{}, "", err
	}
	conversation, err := s.conversations.Create(ctx, adminID, hashToken(token))
	if err != nil {
		return model.Conversation{}, "", err
	}
	return conversation, token, nil
}

func (s *ConversationService) GetAuthorized(ctx context.Context, id, token string) (model.Conversation, error) {
	if token == "" {
		return model.Conversation{}, repository.ErrNotFound
	}
	return s.conversations.FindAuthorized(ctx, id, hashToken(token))
}

func (s *ConversationService) GetAdminAuthorized(ctx context.Context, id, adminID string) (model.Conversation, error) {
	return s.conversations.FindOwned(ctx, id, adminID)
}

func (s *ConversationService) ListForAdmin(ctx context.Context, adminID string) ([]model.Conversation, error) {
	return s.conversations.ListForAdmin(ctx, adminID)
}
