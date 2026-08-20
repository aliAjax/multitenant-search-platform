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
	// docTerms maps a document ID to the set of term keys it contributes to,
	// so a document's stale postings can be removed without scanning every term.
	docTerms map[string]map[string]struct{}
}

func New(a analysis.Analyzer) *Index {
	return &Index{
		analyzer: a,
		terms:    map[string]map[string]platform.Posting{},
		docs:     map[string]platform.Document{},
		deleted:  map[string]bool{},
		docTerms: map[string]map[string]struct{}{},
	}
}
func (i *Index) Add(_ context.Context, d platform.Document) {
	i.mu.Lock()
	defer i.mu.Unlock()
	i.addLocked(d)
}

// addLocked indexes a document's fields. The caller must hold i.mu.
func (i *Index) addLocked(d platform.Document) {
	i.docs[d.ID] = d
	keys := i.docTerms[d.ID]
	if keys == nil {
		keys = map[string]struct{}{}
		i.docTerms[d.ID] = keys
	}
	for field, val := range d.Data {
		if s, ok := val.(string); ok {
			for _, t := range i.analyzer.Analyze(s) {
				k := field + "\x00" + t.Term
				keys[k] = struct{}{}
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

// removeLocked drops a document and every posting it contributed. It is the
// inverse of addLocked and is the reason old terms stop matching on reindex.
// The caller must hold i.mu.
func (i *Index) removeLocked(id string) {
	delete(i.docs, id)
	delete(i.deleted, id)
	for k := range i.docTerms[id] {
		if m := i.terms[k]; m != nil {
			delete(m, id)
			if len(m) == 0 {
				delete(i.terms, k)
			}
		}
	}
	delete(i.docTerms, id)
}

// ReplaceDocument reindexes an existing document by removing its stale postings
// before adding the new ones, so terms that no longer appear stop matching.
func (i *Index) ReplaceDocument(_ context.Context, d platform.Document) {
	i.mu.Lock()
	defer i.mu.Unlock()
	i.removeLocked(d.ID)
	i.addLocked(d)
}
func (i *Index) Remove(id string) { i.mu.Lock(); defer i.mu.Unlock(); i.removeLocked(id) }
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
