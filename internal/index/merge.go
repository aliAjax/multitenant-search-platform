package index

import (
	"context"
	"fmt"
	"github.com/example/multitenant-search/internal/platform"
	"sort"
)

type Merger struct{ store SegmentStore }

func NewMerger(s SegmentStore) *Merger { return &Merger{store: s} }
func (m *Merger) Merge(ctx context.Context, segments []SegmentMeta) (SegmentMeta, error) {
	if len(segments) == 0 {
		return SegmentMeta{}, fmt.Errorf("no segments")
	}
	total := 0
	deleted := 0
	for _, s := range segments {
		total += s.Docs
		deleted += s.Deleted
	}
	out := SegmentMeta{ID: platform.NewID("merge"), Docs: total, Deleted: deleted, Terms: 0}
	if e := m.store.Write(ctx, out, []byte{}); e != nil {
		return out, e
	}
	return out, nil
}
func SortSegments(in []SegmentMeta) []SegmentMeta {
	out := in
	sort.Slice(out, func(i, j int) bool { return out[i].Docs < out[j].Docs })
	return out
}

func CopySegmentMeta(in []SegmentMeta) []SegmentMeta { return in }
