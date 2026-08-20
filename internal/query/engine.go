package query

import (
	"context"
	"fmt"
	"github.com/example/multitenant-search/internal/index"
	"github.com/example/multitenant-search/internal/platform"
	"math"
	"sort"
	"strings"
	"time"
)

type Engine struct {
	idx    *index.Index
	budget int
}

func New(i *index.Index, budget int) *Engine { return &Engine{idx: i, budget: budget} }

type Request struct {
	Match     map[string]string  `json:"match"`
	Term      map[string]any     `json:"term"`
	Phrase    map[string]string  `json:"phrase"`
	Prefix    map[string]string  `json:"prefix"`
	From      int                `json:"from"`
	Size      int                `json:"size"`
	Sort      string             `json:"sort"`
	Highlight bool               `json:"highlight"`
	AggField  string             `json:"agg_field"`
	Range     map[string]float64 `json:"range"`
}

func (e *Engine) Search(ctx context.Context, req Request) (platform.SearchResult, error) {
	start := time.Now()
	if req.Size <= 0 {
		req.Size = 10
	}
	if req.Size > 100 || req.From > 10000 {
		return platform.SearchResult{}, fmt.Errorf("pagination: %w", platform.ErrQuota)
	}
	if req.Size+req.From > e.budget {
		return platform.SearchResult{}, fmt.Errorf("query budget: %w", platform.ErrQuota)
	}
	docs := e.idx.All()
	scores := map[string]float64{}
	matched := map[string]bool{}
	for f, t := range req.Match {
		for _, p := range e.idx.Search(f, t) {
			scores[p.DocID] += 1 + math.Log(float64(p.Frequency))
			matched[p.DocID] = true
		}
	}
	for f, t := range req.Phrase {
		for _, p := range e.idx.Search(f, t) {
			scores[p.DocID] += 2
			matched[p.DocID] = true
		}
	}
	for f, prefix := range req.Prefix {
		for _, d := range docs {
			if v, ok := d.Data[f].(string); ok && strings.HasPrefix(strings.ToLower(v), strings.ToLower(prefix)) {
				scores[d.ID] += 0.5
				matched[d.ID] = true
			}
		}
	}
	for _, d := range docs {
		if len(req.Match)+len(req.Phrase)+len(req.Prefix) > 0 && !matched[d.ID] {
			continue
		}
		for f, v := range req.Term {
			if d.Data[f] != v {
				delete(scores, d.ID)
				matched[d.ID] = false
			}
		}
		for f, min := range req.Range {
			if n, ok := toFloat(d.Data[f]); !ok || n < min {
				delete(scores, d.ID)
				matched[d.ID] = false
			}
		}
	}
	sort.Slice(docs, func(a, b int) bool { return scores[docs[a].ID] > scores[docs[b].ID] })
	res := platform.SearchResult{}
	for _, d := range docs {
		if !matched[d.ID] && len(req.Match)+len(req.Phrase)+len(req.Prefix) > 0 {
			continue
		}
		res.Total++
		if res.Total <= req.From {
			continue
		}
		if len(res.Hits) >= req.Size {
			continue
		}
		h := platform.SearchHit{ID: d.ID, Score: scores[d.ID], Source: d.Data}
		if req.Highlight {
			h.Highlight = map[string][]string{}
			for f, v := range d.Data {
				if s, ok := v.(string); ok {
					h.Highlight[f] = []string{s}
				}
			}
		}
		res.Hits = append(res.Hits, h)
	}
	res.TookMS = time.Since(start).Milliseconds()
	select {
	case <-ctx.Done():
		return res, ctx.Err()
	default:
		return res, nil
	}
}
func toFloat(v any) (float64, bool) {
	switch n := v.(type) {
	case float64:
		return n, true
	case int:
		return float64(n), true
	case int64:
		return float64(n), true
	}
	return 0, false
}
