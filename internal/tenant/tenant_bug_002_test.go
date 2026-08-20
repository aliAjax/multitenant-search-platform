package tenant

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/example/multitenant-search/internal/platform"
)

type countingTenantRepo struct{ calls int }

func (r *countingTenantRepo) Create(context.Context, *platform.Tenant) error { return nil }
func (r *countingTenantRepo) GetTenant(context.Context, string) (*platform.Tenant, error) {
	r.calls++
	return nil, fmt.Errorf("backend: %w", platform.ErrNotFound)
}

type transientTenantRepo struct{}

func (transientTenantRepo) Create(context.Context, *platform.Tenant) error { return nil }
func (transientTenantRepo) GetTenant(context.Context, string) (*platform.Tenant, error) {
	return nil, errors.New("temporary backend")
}

func TestTenantMissingPreservesChain(t *testing.T) {
	store, err := platform.NewStore(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	s := NewService(store, platform.RealClock{})
	_, err = s.Get(context.Background(), "missing")
	if !errors.Is(err, platform.ErrNotFound) {
		t.Fatalf("missing sentinel lost: %v", err)
	}
}

func TestTenantRetryStopsOnNotFound(t *testing.T) {
	repo := &countingTenantRepo{}
	s := NewService(repo, platform.RealClock{})
	_, err := s.GetWithRetry(context.Background(), "missing")
	if !errors.Is(err, platform.ErrNotFound) {
		t.Fatalf("retry changed missing error: %v", err)
	}
	if repo.calls != 1 {
		t.Fatalf("not-found was retried %d times", repo.calls)
	}
}

func TestTenantRetryHonorsCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	s := NewService(transientTenantRepo{}, platform.RealClock{})
	_, err := s.GetWithRetry(ctx, "tenant")
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("canceled retry continued: %v", err)
	}
}

func TestCollectionMissingPreservesChain(t *testing.T) {
	store, err := platform.NewStore(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	_, err = store.GetCollection(context.Background(), "missing")
	if !errors.Is(err, platform.ErrNotFound) {
		t.Fatalf("collection sentinel lost: %v", err)
	}
}
