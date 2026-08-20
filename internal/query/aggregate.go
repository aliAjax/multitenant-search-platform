package query

import (
	"fmt"
	"github.com/example/multitenant-search/internal/platform"
	"sort"
	"time"
)

type Bucket struct {
	Key   string `json:"key"`
	Count int    `json:"count"`
}

func KeywordBuckets(docs []platform.Document, field string) []Bucket {
	m := map[string]int{}
	for _, d := range docs {
		if v, ok := d.Data[field].(string); ok {
			m[v]++
		}
	}
	o := make([]Bucket, 0, len(m))
	for k, n := range m {
		o = append(o, Bucket{Key: k, Count: n})
	}
	sort.Slice(o, func(i, j int) bool {
		if o[i].Count == o[j].Count {
			return o[i].Key < o[j].Key
		}
		return o[i].Count > o[j].Count
	})
	return o
}
func TimeBuckets(docs []platform.Document, field string, layout string) []Bucket {
	m := map[string]int{}
	for _, d := range docs {
		if s, ok := d.Data[field].(string); ok {
			if t, e := time.Parse(time.RFC3339, s); e == nil {
				m[t.UTC().Format(layout)]++
			}
		}
	}
	o := []Bucket{}
	for k, n := range m {
		o = append(o, Bucket{Key: k, Count: n})
	}
	sort.Slice(o, func(i, j int) bool { return o[i].Key < o[j].Key })
	return o
}
func ValidateAggregation(field string) error {
	if field == "" {
		return fmt.Errorf("aggregation field empty")
	}
	return nil
}
