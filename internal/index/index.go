package index

import (
	"context"
	"github.com/example/multitenant-search/internal/analysis"
	"github.com/example/multitenant-search/internal/platform"
	"sort"
	"strings"
	"sync"
)

type Index struct {
	mu       sync.RWMutex
	analyzer analysis.Analyzer
	terms    map[string]map[string]platform.Posting
	docs     map[string]platform.Document
	deleted  map[string]bool
}

func New(a analysis.Analyzer) *Index {
	return &Index{analyzer: a, terms: map[string]map[string]platform.Posting{}, docs: map[string]platform.Document{}, deleted: map[string]bool{}}
}
func (i *Index) Add(_ context.Context, d platform.Document) {
	i.mu.Lock()
	defer i.mu.Unlock()
	i.docs[d.ID] = d
	for field, val := range d.Data {
		if s, ok := val.(string); ok {
			for _, t := range i.analyzer.Analyze(s) {
				k := field + "\x00" + t.Term
				m := i.terms[k]
				if m == nil {
					m = map[string]platform.Posting{}
					i.terms[k] = m
				}
				p := m[d.ID]
				p.DocID = d.ID
				p.Frequency++
				p.Positions = append(p.Positions, t.Pos)
				m[d.ID] = p
			}
		}
	}
}

// ReplaceDocument is used by ingestion paths that need to rebuild a posting list.
func (i *Index) ReplaceDocument(ctx context.Context, d platform.Document) {
	i.Add(ctx, d)
}
func (i *Index) Remove(id string) { i.mu.Lock(); i.deleted[id] = true; i.mu.Unlock() }
func (i *Index) All() []platform.Document {
	i.mu.RLock()
	defer i.mu.RUnlock()
	o := []platform.Document{}
	for id, d := range i.docs {
		if !i.deleted[id] {
			o = append(o, d)
		}
	}
	return o
}
func (i *Index) Search(field, term string) []platform.Posting {
	i.mu.RLock()
	defer i.mu.RUnlock()
	m := i.terms[field+"\x00"+strings.ToLower(term)]
	o := []platform.Posting{}
	for _, p := range m {
		if !i.deleted[p.DocID] {
			o = append(o, p)
		}
	}
	sort.Slice(o, func(a, b int) bool { return o[a].Frequency > o[b].Frequency })
	return o
}
func (i *Index) Count() int { i.mu.RLock(); defer i.mu.RUnlock(); return len(i.docs) - len(i.deleted) }
