package quota

import (
	"context"
	"testing"

	"github.com/example/multitenant-search/internal/platform"
)

func TestZeroManagerConfigure(t *testing.T) {
	m := NewZeroManager(platform.RealClock{})
	m.Configure("tenant-a", 3, 0)
	if len(m.Snapshot()) != 1 {
		t.Fatal("zero manager did not retain configuration")
	}
}

func TestOptionalLimiterDisabled(t *testing.T) {
	if err := UseLimiter(context.Background(), OptionalLimiter(false), "tenant-a", 1); err != nil {
		t.Fatalf("disabled limiter rejected request: %v", err)
	}
}

func TestOptionalLimiterMethodsDisabled(t *testing.T) {
	var l *Limiter
	l.Set("tenant-a", 1)
	l.Release("tenant-a", 1)
}
