package repository

import (
	"context"
	"errors"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"whisper/backend/internal/model"
)

type MessageRepository struct{ pool *pgxpool.Pool }

func NewMessageRepository(pool *pgxpool.Pool) *MessageRepository {
	return &MessageRepository{pool: pool}
}

func (r *MessageRepository) Create(ctx context.Context, conversationID, sender, content string, clientID *string, adminID, visitorHash string) (model.Message, error) {
	query := `INSERT INTO messages (conversation_id, sender_type, content, client_message_id)
		SELECT c.id, $2, $3, $4 FROM conversations c WHERE c.id = $1 AND c.status = 'active' AND `
	args := []any{conversationID, sender, content, clientID}
	if sender == "admin" {
		query += `c.admin_id = $5`
		args = append(args, adminID)
	} else {
		query += `c.visitor_token_hash = $5`
		args = append(args, visitorHash)
	}
	query += ` ON CONFLICT (conversation_id, sender_type, client_message_id) WHERE client_message_id IS NOT NULL DO UPDATE SET client_message_id=EXCLUDED.client_message_id RETURNING id::text, conversation_id::text, sender_type, content, client_message_id::text, created_at, read_at`
	var m model.Message
	err := r.pool.QueryRow(ctx, query, args...).Scan(&m.ID, &m.ConversationID, &m.SenderType, &m.Content, &m.ClientMessageID, &m.CreatedAt, &m.ReadAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return model.Message{}, ErrNotFound
	}
	if err == nil {
		_, err = r.pool.Exec(ctx, `UPDATE conversations SET updated_at=NOW() WHERE id=$1`, conversationID)
	}
	return m, err
}

func (r *MessageRepository) List(ctx context.Context, conversationID string, limit int, before string, adminID, visitorHash string, admin bool) ([]model.Message, error) {
	ownerClause, ownerArg := `c.visitor_token_hash = $3`, visitorHash
	if admin {
		ownerClause, ownerArg = `c.admin_id = $3`, adminID
	}
	var rows pgx.Rows
	var err error
	if before == "" {
		rows, err = r.pool.Query(ctx, `SELECT m.id::text,m.conversation_id::text,m.sender_type,m.content,m.client_message_id::text,m.created_at,m.read_at FROM messages m JOIN conversations c ON c.id=m.conversation_id WHERE m.conversation_id=$1 AND `+ownerClause+` ORDER BY m.created_at DESC,m.id DESC LIMIT $2`, conversationID, limit, ownerArg)
	} else {
		ownerClause = strings.ReplaceAll(ownerClause, "$3", "$4")
		rows, err = r.pool.Query(ctx, `SELECT m.id::text,m.conversation_id::text,m.sender_type,m.content,m.client_message_id::text,m.created_at,m.read_at FROM messages m JOIN conversations c ON c.id=m.conversation_id WHERE m.conversation_id=$1 AND `+ownerClause+` AND (m.created_at,m.id) < (SELECT created_at,id FROM messages WHERE id=$3) ORDER BY m.created_at DESC,m.id DESC LIMIT $2`, conversationID, limit, before, ownerArg)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]model.Message, 0)
	for rows.Next() {
		var m model.Message
		if err := rows.Scan(&m.ID, &m.ConversationID, &m.SenderType, &m.Content, &m.ClientMessageID, &m.CreatedAt, &m.ReadAt); err != nil {
			return nil, err
		}
		items = append(items, m)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if len(items) == 0 {
		var exists bool
		ownerClause = strings.ReplaceAll(ownerClause, "$4", "$2")
		if err := r.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM conversations c WHERE c.id=$1 AND `+ownerClause+`)`, conversationID, ownerArg).Scan(&exists); err != nil {
			return nil, err
		}
		if !exists {
			return nil, ErrNotFound
		}
	}
	for i, j := 0, len(items)-1; i < j; i, j = i+1, j-1 {
		items[i], items[j] = items[j], items[i]
	}
	return items, nil
}

func (r *MessageRepository) MarkRead(ctx context.Context, conversationID, reader, adminID, visitorHash string) (int64, error) {
	condition, arg := `c.visitor_token_hash=$2 AND m.sender_type='admin'`, visitorHash
	if reader == "admin" {
		condition, arg = `c.admin_id=$2 AND m.sender_type='anonymous'`, adminID
	}
	result, err := r.pool.Exec(ctx, `UPDATE messages m SET read_at=NOW() FROM conversations c WHERE m.conversation_id=c.id AND m.conversation_id=$1 AND m.read_at IS NULL AND `+condition, conversationID, arg)
	return result.RowsAffected(), err
}
