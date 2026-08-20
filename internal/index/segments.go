package index

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/example/multitenant-search/internal/platform"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"
)

type SegmentMeta struct {
	ID        string    `json:"id"`
	Docs      int       `json:"docs"`
	Terms     int       `json:"terms"`
	Checksum  string    `json:"checksum"`
	CreatedAt time.Time `json:"created_at"`
	Deleted   int       `json:"deleted"`
}
type SegmentStore interface {
	Write(context.Context, SegmentMeta, []byte) error
	Read(context.Context, string) ([]byte, error)
	List(context.Context) []SegmentMeta
}
type FileSegments struct {
	dir  string
	mu   sync.Mutex
	meta map[string]SegmentMeta
}

func NewFileSegments(dir string) (*FileSegments, error) {
	if e := os.MkdirAll(dir, 0755); e != nil {
		return nil, e
	}
	return &FileSegments{dir: dir, meta: map[string]SegmentMeta{}}, nil
}
func (f *FileSegments) Write(ctx context.Context, m SegmentMeta, b []byte) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	p := filepath.Join(f.dir, m.ID+".seg")
	if e := os.WriteFile(p, b, 0644); e != nil {
		return fmt.Errorf("write segment: %w", e)
	}
	f.meta[m.ID] = m
	return nil
}
func (f *FileSegments) Read(_ context.Context, id string) ([]byte, error) {
	return os.ReadFile(filepath.Join(f.dir, id+".seg"))
}
func (f *FileSegments) List(_ context.Context) []SegmentMeta {
	f.mu.Lock()
	defer f.mu.Unlock()
	o := []SegmentMeta{}
	for _, m := range f.meta {
		o = append(o, m)
	}
	sort.Slice(o, func(i, j int) bool { return o[i].CreatedAt.Before(o[j].CreatedAt) })
	return o
}

func FilterSegments(in []SegmentMeta, keep func(SegmentMeta) bool) []SegmentMeta {
	out := in[:0]
	for _, m := range in {
		if keep == nil || keep(m) {
			out = append(out, m)
		}
	}
	return out
}

func BuildMeta(docs []platform.Document, terms int) (SegmentMeta, []byte, error) {
	b, e := json.Marshal(docs)
	if e != nil {
		return SegmentMeta{}, nil, e
	}
	h := sha256.Sum256(b)
	return SegmentMeta{ID: platform.NewID("seg"), Docs: len(docs), Terms: terms, Checksum: hex.EncodeToString(h[:]), CreatedAt: time.Now().UTC()}, b, nil
}
