package middleware

import (
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type bucket struct {
	started time.Time
	count   int
}
type FixedWindowLimiter struct {
	mu      sync.Mutex
	entries map[string]bucket
	limit   int
	window  time.Duration
}

func NewFixedWindowLimiter(limit int, window time.Duration) *FixedWindowLimiter {
	return &FixedWindowLimiter{entries: make(map[string]bucket), limit: limit, window: window}
}

func (l *FixedWindowLimiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		key := c.RemoteIP()
		allowed, retry := l.Allow(key, time.Now())
		if !allowed {
			seconds := int(retry.Round(time.Second).Seconds())
			if seconds < 1 {
				seconds = 1
			}
			c.Header("Retry-After", strconv.Itoa(seconds))
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "too many requests; try again later"})
			return
		}
		c.Next()
	}
}

func (l *FixedWindowLimiter) Allow(key string, now time.Time) (bool, time.Duration) {
	l.mu.Lock()
	defer l.mu.Unlock()
	current, ok := l.entries[key]
	if !ok || now.Sub(current.started) >= l.window {
		l.entries[key] = bucket{started: now, count: 1}
		return true, 0
	}
	if current.count >= l.limit {
		return false, l.window - now.Sub(current.started)
	}
	current.count++
	l.entries[key] = current
	if len(l.entries) > 10000 {
		for k, v := range l.entries {
			if now.Sub(v.started) >= l.window {
				delete(l.entries, k)
			}
		}
	}
	return true, 0
}
