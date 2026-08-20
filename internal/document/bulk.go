package document

import (
	"context"
	"fmt"
	"github.com/example/multitenant-search/internal/platform"
)

type BulkItem struct {
	ID       string            `json:"id"`
	Action   string            `json:"action"`
	Document platform.Document `json:"document"`
	Error    string            `json:"error,omitempty"`
}
type BulkResponse struct {
	Items     []BulkItem `json:"items"`
	Succeeded int        `json:"succeeded"`
	Failed    int        `json:"failed"`
}

func (s *Service) BulkResponse(ctx context.Context, docs []platform.Document) BulkResponse {
	r := BulkResponse{Items: make([]BulkItem, len(docs))}
	for i := range docs {
		r.Items[i] = BulkItem{ID: docs[i].ID, Action: "index", Document: docs[i]}
		if e := s.Put(ctx, &docs[i]); e != nil {
			r.Items[i].Error = e.Error()
			r.Failed++
		} else {
			r.Succeeded++
		}
	}
	return r
}
func ValidateDocument(c platform.Collection, d platform.Document) error {
	for n, f := range c.Mappings {
		if f.Required {
			if _, ok := d.Data[n]; !ok {
				return fmt.Errorf("missing field %s: %w", n, platform.ErrInvalid)
			}
		}
	}
	return nil
}
