package repository

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"whisper/backend/internal/model"
)

var ErrNotFound = errors.New("record not found")

type Admin struct {
	ID           string
	Username     string
	PasswordHash string
}

type AdminRepository struct {
	pool *pgxpool.Pool
}

func NewAdminRepository(pool *pgxpool.Pool) *AdminRepository {
	return &AdminRepository{pool: pool}
}

func (r *AdminRepository) FindByUsername(ctx context.Context, username string) (Admin, error) {
	var admin Admin
	err := r.pool.QueryRow(ctx, `
		SELECT id::text, username, password_hash
		FROM admin_users WHERE username = $1
	`, username).Scan(&admin.ID, &admin.Username, &admin.PasswordHash)
	if errors.Is(err, pgx.ErrNoRows) {
		return Admin{}, ErrNotFound
	}
	return admin, err
}

func (r *AdminRepository) FindByID(ctx context.Context, id string) (model.AdminUser, error) {
	var admin model.AdminUser
	err := r.pool.QueryRow(ctx, `
		SELECT id::text, username FROM admin_users WHERE id = $1
	`, id).Scan(&admin.ID, &admin.Username)
	if errors.Is(err, pgx.ErrNoRows) {
		return model.AdminUser{}, ErrNotFound
	}
	return admin, err
}

func (r *AdminRepository) Create(ctx context.Context, username, passwordHash string) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO admin_users (username, password_hash) VALUES ($1, $2)
	`, username, passwordHash)
	return err
}

func (r *AdminRepository) CreateSession(ctx context.Context, adminID, tokenHash string, expiresAt time.Time) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO admin_sessions (admin_id, refresh_token_hash, expires_at)
		VALUES ($1, $2, $3)
	`, adminID, tokenHash, expiresAt)
	return err
}

func (r *AdminRepository) FindSession(ctx context.Context, tokenHash string) (model.AdminUser, error) {
	var admin model.AdminUser
	err := r.pool.QueryRow(ctx, `
		SELECT u.id::text, u.username
		FROM admin_sessions s
		JOIN admin_users u ON u.id = s.admin_id
		WHERE s.refresh_token_hash = $1 AND s.expires_at > NOW()
	`, tokenHash).Scan(&admin.ID, &admin.Username)
	if errors.Is(err, pgx.ErrNoRows) {
		return model.AdminUser{}, ErrNotFound
	}
	return admin, err
}

func (r *AdminRepository) DeleteSession(ctx context.Context, tokenHash string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM admin_sessions WHERE refresh_token_hash = $1`, tokenHash)
	return err
}
