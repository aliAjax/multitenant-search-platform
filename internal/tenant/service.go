package tenant

import (
	"context"
	"fmt"
	"github.com/example/multitenant-search/internal/platform"
	"sync"
)

type Repository interface {
	Create(context.Context, *platform.Tenant) error
	GetTenant(context.Context, string) (*platform.Tenant, error)
}
type Service struct {
	repo  Repository
	clock platform.Clock
	mu    sync.RWMutex
}

func NewService(r Repository, c platform.Clock) *Service { return &Service{repo: r, clock: c} }
func (s *Service) Create(ctx context.Context, name string, quota int) (*platform.Tenant, error) {
	if name == "" {
		return nil, fmt.Errorf("tenant name: %w", platform.ErrInvalid)
	}
	t := &platform.Tenant{ID: platform.NewID("t"), Name: name, Quota: quota, CreatedAt: s.clock.Now()}
	if e := s.repo.Create(ctx, t); e != nil {
		return nil, fmt.Errorf("create tenant: %w", e)
	}
	return t, nil
}
func (s *Service) Get(ctx context.Context, id string) (*platform.Tenant, error) {
	t, e := s.repo.GetTenant(ctx, id)
	if e != nil {
		return nil, fmt.Errorf("get tenant: %w", e)
	}
	return t, nil
}
