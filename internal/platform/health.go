package platform

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type HealthCheck interface {
	Name() string
	Check(context.Context) error
}
type HealthRegistry struct {
	mu     sync.RWMutex
	checks []HealthCheck
}

func NewHealthRegistry() *HealthRegistry { return &HealthRegistry{checks: []HealthCheck{}} }
func (r *HealthRegistry) Register(c HealthCheck) {
	r.mu.Lock()
	r.checks = append(r.checks, c)
	r.mu.Unlock()
}
func (r *HealthRegistry) Ready(ctx context.Context) error {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, c := range r.checks {
		if e := c.Check(ctx); e != nil {
			return fmt.Errorf("%s: %w", c.Name(), e)
		}
	}
	return nil
}

type MemoryCheck struct {
	Label string
	Fn    func(context.Context) error
}

func (c MemoryCheck) Name() string { return c.Label }
func (c MemoryCheck) Check(ctx context.Context) error {
	if c.Fn == nil {
		return nil
	}
	return c.Fn(ctx)
}

type Lease struct {
	Owner   string
	Expires time.Time
}

func (l Lease) Valid(now time.Time) bool { return l.Owner != "" && now.Before(l.Expires) }
