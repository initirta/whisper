package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"whisper/backend/internal/model"
)

type ConversationRepository struct {
	pool *pgxpool.Pool
}

func NewConversationRepository(pool *pgxpool.Pool) *ConversationRepository {
	return &ConversationRepository{pool: pool}
}

func (r *ConversationRepository) Create(ctx context.Context, adminID, tokenHash string) (model.Conversation, error) {
	var conversation model.Conversation
	err := r.pool.QueryRow(ctx, `
		INSERT INTO conversations (admin_id, visitor_token_hash)
		VALUES ($1, $2)
		RETURNING id::text, status, created_at
	`, adminID, tokenHash).Scan(&conversation.ID, &conversation.Status, &conversation.CreatedAt)
	return conversation, err
}

func (r *ConversationRepository) FindAuthorized(ctx context.Context, id, tokenHash string) (model.Conversation, error) {
	var conversation model.Conversation
	err := r.pool.QueryRow(ctx, `
		SELECT id::text, status, created_at
		FROM conversations
		WHERE id = $1 AND visitor_token_hash = $2
	`, id, tokenHash).Scan(&conversation.ID, &conversation.Status, &conversation.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return model.Conversation{}, ErrNotFound
	}
	return conversation, err
}

func (r *ConversationRepository) FindOwned(ctx context.Context, id, adminID string) (model.Conversation, error) {
	var conversation model.Conversation
	err := r.pool.QueryRow(ctx, `SELECT id::text,status,created_at FROM conversations WHERE id=$1 AND admin_id=$2`, id, adminID).Scan(&conversation.ID, &conversation.Status, &conversation.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return model.Conversation{}, ErrNotFound
	}
	return conversation, err
}

func (r *ConversationRepository) ListForAdmin(ctx context.Context, adminID string) ([]model.Conversation, error) {
	rows, err := r.pool.Query(ctx, `SELECT c.id::text,c.status,c.created_at,c.updated_at,COALESCE(last.content,''),COALESCE(unread.total,0)
	FROM conversations c
	LEFT JOIN LATERAL (SELECT content FROM messages WHERE conversation_id=c.id ORDER BY created_at DESC,id DESC LIMIT 1) last ON true
	LEFT JOIN LATERAL (SELECT count(*)::int AS total FROM messages WHERE conversation_id=c.id AND sender_type='anonymous' AND read_at IS NULL) unread ON true
	WHERE c.admin_id=$1 ORDER BY c.updated_at DESC LIMIT 100`, adminID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]model.Conversation, 0)
	for rows.Next() {
		var c model.Conversation
		if err := rows.Scan(&c.ID, &c.Status, &c.CreatedAt, &c.UpdatedAt, &c.LastMessage, &c.UnreadCount); err != nil {
			return nil, err
		}
		items = append(items, c)
	}
	return items, rows.Err()
}
func (r *ConversationRepository) AdminIDBySlug(ctx context.Context, slug string) (string, error) {
	var id string
	err := r.pool.QueryRow(ctx, `SELECT id::text FROM admin_users WHERE username = $1`, slug).Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", ErrNotFound
	}
	return id, err
}
