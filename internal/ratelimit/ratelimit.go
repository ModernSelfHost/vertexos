package ratelimit

import (
	"sync"
	"time"
)

// Limiter is a simple per-key fixed window counter.
type Limiter struct {
	mu      sync.Mutex
	counts  map[string]int
	resetAt map[string]time.Time
	limit   int
	window  time.Duration
}

func New(limit int, window time.Duration) *Limiter {
	return &Limiter{counts: map[string]int{}, resetAt: map[string]time.Time{}, limit: limit, window: window}
}

// Allow returns true if the key is under the limit.
func (l *Limiter) Allow(key string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	now := time.Now()
	if t, ok := l.resetAt[key]; !ok || now.After(t) {
		l.counts[key] = 0
		l.resetAt[key] = now.Add(l.window)
	}
	if l.counts[key] >= l.limit {
		return false
	}
	l.counts[key]++
	return true
}
