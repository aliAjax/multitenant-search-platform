package document

import (
	"context"
	"fmt"
	"github.com/example/multitenant-search/internal/index"
	"github.com/example/multitenant-search/internal/platform"
	"sync"
	"time"
)

type Service struct {
	store   *platform.Store
	mu      sync.RWMutex
	indexes map[string]*index.Index
	clock   platform.Clock
}

func NewService(s *platform.Store, c platform.Clock) *Service {
	return &Service{store: s, indexes: map[string]*index.Index{}, clock: c}
}
func (s *Service) EnsureIndex(id string) *index.Index {
	s.mu.RLock()
	if i := s.indexes[id]; i != nil {
		s.mu.RUnlock()
		return i
	}
	s.mu.RUnlock()
	s.mu.Lock()
	defer s.mu.Unlock()
	// Re-check under the write lock: another goroutine may have created the
	// index between the read unlock and this acquisition.
	if i := s.indexes[id]; i != nil {
		return i
	}
	i := index.New(nilAnalyzer{})
	s.indexes[id] = i
	return i
}

type nilAnalyzer struct{}

func (nilAnalyzer) Analyze(x string) []platform.Token { return simpleTokens(x) }
func simpleTokens(x string) []platform.Token {
	out := []platform.Token{}
	for p, w := range splitWords(x) {
		out = append(out, platform.Token{Term: w, Pos: p})
	}
	return out
}
func splitWords(x string) []string {
	r := []string{}
	cur := ""
	for _, c := range x {
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') {
			cur += string(c)
		} else if cur != "" {
			r = append(r, cur)
			cur = ""
		}
	}
	if cur != "" {
		r = append(r, cur)
	}
	return r
}
func (s *Service) Put(ctx context.Context, d *platform.Document) error {
	if d.ID == "" {
		d.ID = platform.NewID("doc")
	}
	if d.Version == 0 {
		d.Version = 1
	}
	d.UpdatedAt = s.clock.Now()
	if e := s.store.PutDoc(ctx, d); e != nil {
		return fmt.Errorf("put document: %w", e)
	}
	// Replace rather than accumulate postings so re-indexing the same ID drops
	// terms that no longer appear in the document.
	s.EnsureIndex(d.CollectionID).ReplaceDocument(ctx, *d)
	return nil
}
func (s *Service) Delete(ctx context.Context, id string) error {
	d, e := s.store.GetDoc(ctx, id)
	if e != nil {
		return e
	}
	if e = s.store.DeleteDoc(ctx, id); e != nil {
		return e
	}
	if i := s.EnsureIndex(d.CollectionID); i != nil {
		i.Remove(id)
	}
	return nil
}
func (s *Service) Bulk(ctx context.Context, docs []platform.Document) []error {
	errs := make([]error, len(docs))
	for i := range docs {
		errs[i] = s.Put(ctx, &docs[i])
	}
	return errs
}
func (s *Service) List(ctx context.Context, tenant, col string) []platform.Document {
	return s.store.ListDocs(ctx, tenant, col)
}

var _ = time.Second
