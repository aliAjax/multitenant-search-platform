package analysis

import (
	"github.com/example/multitenant-search/internal/platform"
	"sync"
)

type Registry struct {
	mu    sync.RWMutex
	items map[string]map[int]Analyzer
}

func NewRegistry() *Registry { return &Registry{items: map[string]map[int]Analyzer{}} }
func (r *Registry) Register(name string, version int, a Analyzer) {
	r.mu.Lock()
	if r.items[name] == nil {
		r.items[name] = map[int]Analyzer{}
	}
	r.items[name][version] = a
	r.mu.Unlock()
}
func (r *Registry) Get(name string, version int) (Analyzer, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	m := r.items[name]
	a, ok := m[version]
	return a, ok
}
func (r *Registry) Analyze(name string, version int, text string) []platform.Token {
	a, ok := r.Get(name, version)
	if !ok {
		return nil
	}
	return a.Analyze(text)
}

func (r *Registry) AnalyzeOrDefault(name string, version int, text string) []platform.Token {
	a, ok := r.Get(name, version)
	if !ok {
		a = NewPipeline(nil, nil, false)
	}
	return a.Analyze(text)
}
