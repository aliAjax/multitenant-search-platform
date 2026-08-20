package platform

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"
)

type Store struct {
	mu          sync.RWMutex
	tenants     map[string]Tenant
	collections map[string]Collection
	docs        map[string]Document
	wal         *EventLog
}

func NewStore(dir string) (*Store, error) {
	w, e := NewEventLog(dir)
	if e != nil {
		return nil, e
	}
	s := &Store{tenants: map[string]Tenant{}, collections: map[string]Collection{}, docs: map[string]Document{}, wal: w}
	s.replay()
	return s, nil
}
func (s *Store) replay() {
	events, e := s.wal.ReadAll()
	if e != nil {
		return
	}
	for _, raw := range events {
		var ev struct {
			Type       string      `json:"type"`
			Tenant     *Tenant     `json:"tenant"`
			Collection *Collection `json:"collection"`
			Doc        *Document   `json:"doc"`
			ID         string      `json:"id"`
		}
		if json.Unmarshal(raw, &ev) != nil {
			continue
		}
		switch ev.Type {
		case "tenant":
			if ev.Tenant != nil {
				s.tenants[ev.Tenant.ID] = *ev.Tenant
			}
		case "collection":
			if ev.Collection != nil {
				s.collections[ev.Collection.ID] = *ev.Collection
			}
		case "doc":
			if ev.Doc != nil {
				s.docs[ev.Doc.ID] = *ev.Doc
			}
		case "delete":
			delete(s.docs, ev.ID)
		}
	}
}
func (s *Store) CreateTenant(ctx context.Context, t *Tenant) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.tenants[t.ID]; ok {
		return ErrConflict
	}
	s.tenants[t.ID] = *t
	return s.wal.Append(ctx, map[string]any{"type": "tenant", "tenant": t})
}
func (s *Store) Create(ctx context.Context, t *Tenant) error { return s.CreateTenant(ctx, t) }
func (s *Store) GetTenant(ctx context.Context, id string) (*Tenant, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	t, ok := s.tenants[id]
	if !ok {
		return nil, fmt.Errorf("tenant lookup: %w", ErrNotFound)
	}
	return &t, nil
}
func (s *Store) ListTenants(context.Context) []Tenant {
	s.mu.RLock()
	defer s.mu.RUnlock()
	o := make([]Tenant, 0, len(s.tenants))
	for _, t := range s.tenants {
		o = append(o, t)
	}
	return o
}
func (s *Store) CreateCollection(ctx context.Context, c *Collection) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.collections[c.ID]; ok {
		return ErrConflict
	}
	s.collections[c.ID] = *c
	return s.wal.Append(ctx, map[string]any{"type": "collection", "collection": c})
}
func (s *Store) GetCollection(ctx context.Context, id string) (*Collection, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	c, ok := s.collections[id]
	if !ok {
		return nil, fmt.Errorf("collection lookup: %w", ErrNotFound)
	}
	return &c, nil
}
func (s *Store) UpdateCollection(ctx context.Context, c *Collection) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.collections[c.ID] = *c
	return s.wal.Append(ctx, map[string]any{"type": "collection", "collection": c})
}
func (s *Store) ListCollections(_ context.Context, tenant string) []Collection {
	s.mu.RLock()
	defer s.mu.RUnlock()
	o := []Collection{}
	for _, c := range s.collections {
		if c.TenantID == tenant {
			o = append(o, c)
		}
	}
	return o
}
func (s *Store) PutDoc(ctx context.Context, d *Document) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, old := range s.docs {
		if old.TenantID == d.TenantID && old.CollectionID == d.CollectionID && old.ID == d.ID && old.Version >= d.Version {
			return fmt.Errorf("version conflict: %w", ErrConflict)
		}
	}
	s.docs[d.ID] = *d
	return s.wal.Append(ctx, map[string]any{"type": "doc", "doc": d})
}
func (s *Store) GetDoc(_ context.Context, id string) (*Document, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	d, ok := s.docs[id]
	if !ok {
		return nil, ErrNotFound
	}
	return &d, nil
}
func (s *Store) ListDocs(_ context.Context, tenant, collection string) []Document {
	s.mu.RLock()
	defer s.mu.RUnlock()
	o := []Document{}
	for _, d := range s.docs {
		if d.TenantID == tenant && d.CollectionID == collection && !d.Deleted {
			o = append(o, d)
		}
	}
	return o
}
func (s *Store) DeleteDoc(ctx context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	d, ok := s.docs[id]
	if !ok {
		return ErrNotFound
	}
	d.Deleted = true
	d.Version++
	s.docs[id] = d
	return s.wal.Append(ctx, map[string]any{"type": "doc", "doc": &d})
}
func (s *Store) Snapshot(path string) error {
	s.mu.RLock()
	defer s.mu.RUnlock()
	v := map[string]any{"tenants": s.tenants, "collections": s.collections, "docs": s.docs}
	b, e := json.MarshalIndent(v, "", "  ")
	if e != nil {
		return e
	}
	return os.WriteFile(path, b, 0644)
}
