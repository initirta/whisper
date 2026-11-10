package service

import (
	"context"
	"errors"
	"time"

	"golang.org/x/crypto/bcrypt"
	"whisper/backend/internal/model"
	"whisper/backend/internal/repository"
)

var ErrInvalidCredentials = errors.New("invalid username or password")

const sessionLifetime = 7 * 24 * time.Hour

type AuthService struct {
	admins *repository.AdminRepository
}

func NewAuthService(admins *repository.AdminRepository) *AuthService {
	return &AuthService{admins: admins}
}

func (s *AuthService) Login(ctx context.Context, username, password string) (model.AdminUser, string, time.Time, error) {
	admin, err := s.admins.FindByUsername(ctx, username)
	if errors.Is(err, repository.ErrNotFound) || (err == nil && bcrypt.CompareHashAndPassword([]byte(admin.PasswordHash), []byte(password)) != nil) {
		return model.AdminUser{}, "", time.Time{}, ErrInvalidCredentials
	}
	if err != nil {
		return model.AdminUser{}, "", time.Time{}, err
	}

	token, err := newAccessToken()
	if err != nil {
		return model.AdminUser{}, "", time.Time{}, err
	}
	expiresAt := time.Now().UTC().Add(sessionLifetime)
	if err := s.admins.CreateSession(ctx, admin.ID, hashToken(token), expiresAt); err != nil {
		return model.AdminUser{}, "", time.Time{}, err
	}
	return model.AdminUser{ID: admin.ID, Username: admin.Username}, token, expiresAt, nil
}

func (s *AuthService) ValidateSession(ctx context.Context, token string) (model.AdminUser, error) {
	if token == "" {
		return model.AdminUser{}, repository.ErrNotFound
	}
	return s.admins.FindSession(ctx, hashToken(token))
}

func (s *AuthService) Logout(ctx context.Context, token string) error {
	if token == "" {
		return nil
	}
	return s.admins.DeleteSession(ctx, hashToken(token))
}
func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(hash), err
}
