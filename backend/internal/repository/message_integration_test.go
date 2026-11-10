package repository

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func TestMessageCreateIsIdempotent(t *testing.T) {
	databaseURL := os.Getenv("WHISPER_TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("set WHISPER_TEST_DATABASE_URL to run PostgreSQL integration tests")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		t.Fatal("connect test database:", err)
	}
	defer pool.Close()
	if err := pool.Ping(ctx); err != nil {
		t.Fatal("ping test database:", err)
	}
	var nonce [8]byte
	if _, err := rand.Read(nonce[:]); err != nil {
		t.Fatal(err)
	}
	username := fmt.Sprintf("phase7_%s", hex.EncodeToString(nonce[:]))
	visitorHash := hex.EncodeToString(nonce[:]) + hex.EncodeToString(nonce[:])
	var adminID, conversationID string
	if err := pool.QueryRow(ctx, `INSERT INTO admin_users(username,password_hash) VALUES($1,'integration-test') RETURNING id::text`, username).Scan(&adminID); err != nil {
		t.Fatal("create test admin:", err)
	}
	defer func() { _, _ = pool.Exec(context.Background(), `DELETE FROM admin_users WHERE id=$1`, adminID) }()
	if err := pool.QueryRow(ctx, `INSERT INTO conversations(admin_id,visitor_token_hash) VALUES($1,$2) RETURNING id::text`, adminID, visitorHash).Scan(&conversationID); err != nil {
		t.Fatal("create test conversation:", err)
	}
	clientID := "31ee7d14-8521-4e83-8e2c-96d19a02f6ac"
	repository := NewMessageRepository(pool)
	first, err := repository.Create(ctx, conversationID, "anonymous", "hello", &clientID, "", visitorHash)
	if err != nil {
		t.Fatal("create first message:", err)
	}
	second, err := repository.Create(ctx, conversationID, "anonymous", "hello retry", &clientID, "", visitorHash)
	if err != nil {
		t.Fatal("retry message:", err)
	}
	if first.ID != second.ID {
		t.Fatalf("retry created a different message: first=%s second=%s", first.ID, second.ID)
	}
	if second.Content != "hello" {
		t.Fatalf("idempotent retry changed stored content: got %q", second.Content)
	}
	var count int
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM messages WHERE conversation_id=$1 AND sender_type='anonymous' AND client_message_id=$2`, conversationID, clientID).Scan(&count); err != nil {
		t.Fatal("count test messages:", err)
	}
	if count != 1 {
		t.Fatalf("expected one stored message after retry, got %d", count)
	}
}
