package platform

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"
)

type Reloadable struct {
	mu   sync.RWMutex
	cfg  Config
	path string
	mod  time.Time
}

func NewReloadable(path string, c Config) *Reloadable { return &Reloadable{cfg: c, path: path} }
func (r *Reloadable) Get() Config                     { r.mu.RLock(); defer r.mu.RUnlock(); return r.cfg }
func (r *Reloadable) Reload(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	st, e := os.Stat(r.path)
	if e != nil {
		return fmt.Errorf("config stat: %w", e)
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if !st.ModTime().After(r.mod) {
		return nil
	}
	c, e := LoadConfig(r.path)
	if e != nil {
		return e
	}
	r.cfg = c
	r.mod = st.ModTime()
	return nil
}
func (r *Reloadable) Watch(ctx context.Context, interval time.Duration, fn func(Config)) {
	t := time.NewTicker(interval)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			if r.Reload(ctx) == nil && fn != nil {
				fn(r.Get())
			}
		}
	}
}
