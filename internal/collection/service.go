package collection

import (
	"context"
	"fmt"
	"github.com/example/multitenant-search/internal/platform"
)

type Repository interface {
	CreateCollection(context.Context, *platform.Collection) error
	GetCollection(context.Context, string) (*platform.Collection, error)
	UpdateCollection(context.Context, *platform.Collection) error
}
type Service struct {
	repo  Repository
	clock platform.Clock
}

func NewService(r Repository, c platform.Clock) *Service { return &Service{repo: r, clock: c} }
func (s *Service) Create(ctx context.Context, tenantID, name string, m map[string]platform.FieldMapping, shards, replicas int) (*platform.Collection, error) {
	if tenantID == "" || name == "" || len(m) == 0 {
		return nil, fmt.Errorf("collection arguments: %w", platform.ErrInvalid)
	}
	if shards < 1 {
		shards = 1
	}
	c := &platform.Collection{ID: platform.NewID("c"), TenantID: tenantID, Name: name, Mappings: m, MappingVersion: 1, Shards: shards, Replicas: replicas, CreatedAt: s.clock.Now()}
	if e := s.repo.CreateCollection(ctx, c); e != nil {
		return nil, fmt.Errorf("create collection: %w", e)
	}
	return c, nil
}
func (s *Service) PublishMapping(ctx context.Context, id string, next map[string]platform.FieldMapping) (*platform.Collection, error) {
	c, e := s.repo.GetCollection(ctx, id)
	if e != nil {
		return nil, e
	}
	for n, old := range c.Mappings {
		if nw, ok := next[n]; ok && old.Type != nw.Type {
			return nil, fmt.Errorf("mapping field %s type change: %w", n, platform.ErrConflict)
		}
	}
	c.Mappings = next
	c.MappingVersion++
	if e = s.repo.UpdateCollection(ctx, c); e != nil {
		return nil, fmt.Errorf("publish mapping: %w", e)
	}
	return c, nil
}
