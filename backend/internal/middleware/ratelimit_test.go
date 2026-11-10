package middleware

import (
	"testing"
	"time"
)

func TestFixedWindowLimiter(t *testing.T) {
	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	l := NewFixedWindowLimiter(2, time.Minute)
	if ok, _ := l.Allow("ip-a", start); !ok {
		t.Fatal("first request should pass")
	}
	if ok, _ := l.Allow("ip-a", start.Add(time.Second)); !ok {
		t.Fatal("second request should pass")
	}
	if ok, retry := l.Allow("ip-a", start.Add(2*time.Second)); ok || retry <= 0 {
		t.Fatal("third request should be limited")
	}
	if ok, _ := l.Allow("ip-b", start.Add(2*time.Second)); !ok {
		t.Fatal("different IP should have a separate bucket")
	}
	if ok, _ := l.Allow("ip-a", start.Add(time.Minute)); !ok {
		t.Fatal("window should reset")
	}
}
