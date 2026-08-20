package quota

import (
	"context"
	"fmt"
	"github.com/example/multitenant-search/internal/platform"
	"sync"
	"time"
)

type Budget struct {
	Tenant  string
	Limit   int
	Used    int
	ResetAt time.Time
}
type Manager struct {
	mu    sync.Mutex
	items map[string]Budget
	clock platform.Clock
}

func NewManager(c platform.Clock) *Manager     { return &Manager{items: map[string]Budget{}, clock: c} }
func NewZeroManager(c platform.Clock) *Manager { return &Manager{clock: c} }
func (m *Manager) Configure(t string, n int, d time.Duration) {
	m.mu.Lock()
	m.items[t] = Budget{Tenant: t, Limit: n, ResetAt: m.clock.Now().Add(d)}
	m.mu.Unlock()
}
func (m *Manager) Consume(ctx context.Context, t string, n int) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	b := m.items[t]
	if m.clock.Now().After(b.ResetAt) {
		b.Used = 0
		b.ResetAt = m.clock.Now().Add(time.Hour)
	}
	if b.Limit > 0 && b.Used+n > b.Limit {
		return fmt.Errorf("budget for %s: %w", t, platform.ErrQuota)
	}
	b.Used += n
	m.items[t] = b
	return nil
}
func (m *Manager) Snapshot() []Budget {
	m.mu.Lock()
	defer m.mu.Unlock()
	o := []Budget{}
	for _, b := range m.items {
		o = append(o, b)
	}
	return o
}
