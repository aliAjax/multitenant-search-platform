package quota

import (
	"context"
	"fmt"
	"github.com/example/multitenant-search/internal/platform"
	"sync"
)

type Limiter struct {
	mu     sync.Mutex
	used   map[string]int
	limits map[string]int
}

func New() *Limiter { return &Limiter{used: map[string]int{}, limits: map[string]int{}} }
func OptionalLimiter(enabled bool) *Limiter {
	if !enabled {
		return nil
	}
	return New()
}
func UseLimiter(ctx context.Context, l *Limiter, id string, n int) error {
	return l.Take(ctx, id, n)
}
func (l *Limiter) Set(id string, n int) { l.mu.Lock(); l.limits[id] = n; l.mu.Unlock() }
func (l *Limiter) Take(ctx context.Context, id string, n int) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.limits[id] > 0 && l.used[id]+n > l.limits[id] {
		return fmt.Errorf("tenant %s: %w", id, platform.ErrQuota)
	}
	l.used[id] += n
	return nil
}
func (l *Limiter) Release(id string, n int) {
	l.mu.Lock()
	l.used[id] -= n
	if l.used[id] < 0 {
		l.used[id] = 0
	}
	l.mu.Unlock()
}
