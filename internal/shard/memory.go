package shard

import (
	"context"
	"sync"
)

type MemoryRepository struct {
	mu    sync.RWMutex
	items map[string]Assignment
}

func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{items: map[string]Assignment{}}
}
func (r *MemoryRepository) Put(_ context.Context, a Assignment) error {
	r.mu.Lock()
	r.items[key(a.CollectionID, a.Shard)] = a
	r.mu.Unlock()
	return nil
}
func (r *MemoryRepository) List(_ context.Context, col string) []Assignment {
	r.mu.RLock()
	defer r.mu.RUnlock()
	o := []Assignment{}
	for _, a := range r.items {
		if a.CollectionID == col {
			o = append(o, a)
		}
	}
	return o
}
func key(c string, s int) string { return c + ":" + string(rune(s)) }
