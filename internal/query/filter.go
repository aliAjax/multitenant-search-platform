package query

import (
	"fmt"
	"github.com/example/multitenant-search/internal/platform"
	"strings"
)

type Filter struct {
	Field  string
	Equals any
	Prefix string
	Min    *float64
	Max    *float64
}

func ApplyFilters(d platform.Document, filters []Filter) bool {
	for _, f := range filters {
		v, ok := d.Data[f.Field]
		if !ok {
			return false
		}
		if f.Equals != nil && fmt.Sprint(v) != fmt.Sprint(f.Equals) {
			return false
		}
		if f.Prefix != "" {
			s, ok := v.(string)
			if !ok || !strings.HasPrefix(strings.ToLower(s), strings.ToLower(f.Prefix)) {
				return false
			}
		}
		if n, ok := toFloat(v); ok {
			if f.Min != nil && n < *f.Min {
				return false
			}
			if f.Max != nil && n > *f.Max {
				return false
			}
		}
	}
	return true
}
func ExplainQuery(r Request) map[string]any {
	return map[string]any{"match_fields": len(r.Match), "term_fields": len(r.Term), "phrase_fields": len(r.Phrase), "prefix_fields": len(r.Prefix), "from": r.From, "size": r.Size}
}
