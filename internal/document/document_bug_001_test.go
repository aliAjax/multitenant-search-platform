package document

import (
	"context"
	"sync"
	"testing"

	"github.com/example/multitenant-search/internal/analysis"
	"github.com/example/multitenant-search/internal/index"
	"github.com/example/multitenant-search/internal/platform"
)

func TestConcurrentEnsureIndex(t *testing.T) {
	store, err := platform.NewStore(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	s := NewService(store, platform.RealClock{})
	start := make(chan struct{})
	var wg sync.WaitGroup
	for n := 0; n < 16; n++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			if s.EnsureIndex("shared") == nil {
				t.Error("missing index")
			}
		}()
	}
	close(start)
	wg.Wait()
}

func TestIndexReplacesTerms(t *testing.T) {
	idx := index.New(analysis.NewPipeline(nil, nil, false))
	ctx := context.Background()
	idx.Add(ctx, platform.Document{ID: "d1", Data: map[string]any{"body": "alpha"}})
	idx.ReplaceDocument(ctx, platform.Document{ID: "d1", Data: map[string]any{"body": "beta"}})
	if got := idx.Search("body", "alpha"); len(got) != 0 {
		t.Fatalf("old posting remains: %#v", got)
	}
	if got := idx.Search("body", "beta"); len(got) != 1 {
		t.Fatalf("new posting missing: %#v", got)
	}
}
