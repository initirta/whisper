package service

import (
	"context"
	"strings"
	"testing"
)

func TestMessageValidation(t *testing.T) {
	s := NewMessageService(nil)
	for _, content := range []string{"", "   ", strings.Repeat("a", 5001)} {
		if _, err := s.Create(context.Background(), "id", "anonymous", content, nil, "", "token"); err != ErrInvalidMessage {
			t.Fatalf("content length %d: expected ErrInvalidMessage, got %v", len(content), err)
		}
	}
	badID := "not-a-uuid"
	if _, err := s.Create(context.Background(), "id", "anonymous", "hello", &badID, "", "token"); err != ErrInvalidMessage {
		t.Fatalf("expected invalid client ID, got %v", err)
	}
}
