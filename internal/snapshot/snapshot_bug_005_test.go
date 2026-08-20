package snapshot

import (
	"context"
	"testing"

	"github.com/example/multitenant-search/internal/platform"
)

func TestSnapshotCreateUsesCallContext(t *testing.T) {
	store, err := platform.NewStore(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	s := New(store, t.TempDir())
	old, cancel := context.WithCancel(context.Background())
	cancel()
	s.SetContext(old)
	if _, err := s.CreateWithContext(context.Background()); err != nil {
		t.Fatalf("fresh context was replaced by stale one: %v", err)
	}
	if _, err := s.CreateWithContext(nil); err != nil {
		t.Fatalf("nil context should use a request-independent context: %v", err)
	}
}
