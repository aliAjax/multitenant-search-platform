package snapshot

import (
	"context"
	"fmt"
	"github.com/example/multitenant-search/internal/platform"
	"os"
	"path/filepath"
	"time"
)

type Service struct {
	store *platform.Store
	dir   string
	ctx   context.Context
}

func New(s *platform.Store, dir string) *Service  { return &Service{store: s, dir: dir} }
func (s *Service) SetContext(ctx context.Context) { s.ctx = ctx }
func (s *Service) CreateWithContext(ctx context.Context) (string, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	return s.Create(ctx)
}
func (s *Service) Create(ctx context.Context) (string, error) {
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
	}
	if e := os.MkdirAll(s.dir, 0755); e != nil {
		return "", e
	}
	name := fmt.Sprintf("snapshot-%s.json", time.Now().UTC().Format("20060102T150405.000000000Z"))
	p := filepath.Join(s.dir, name)
	if e := s.store.Snapshot(p); e != nil {
		return "", fmt.Errorf("snapshot: %w", e)
	}
	return p, nil
}
