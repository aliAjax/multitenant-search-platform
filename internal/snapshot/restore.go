package snapshot

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/example/multitenant-search/internal/platform"
	"os"
)

type SnapshotData struct {
	Tenants     map[string]platform.Tenant     `json:"tenants"`
	Collections map[string]platform.Collection `json:"collections"`
	Docs        map[string]platform.Document   `json:"docs"`
}

func Load(path string) (SnapshotData, error) {
	b, e := os.ReadFile(path)
	if e != nil {
		return SnapshotData{}, fmt.Errorf("read snapshot: %w", e)
	}
	var d SnapshotData
	if e = json.Unmarshal(b, &d); e != nil {
		return d, fmt.Errorf("decode snapshot: %w", e)
	}
	return d, nil
}
func Restore(ctx context.Context, s *platform.Store, d SnapshotData) error {
	for _, t := range d.Tenants {
		if e := s.CreateTenant(ctx, &t); e != nil {
			return e
		}
	}
	for _, c := range d.Collections {
		if e := s.CreateCollection(ctx, &c); e != nil {
			return e
		}
	}
	for _, doc := range d.Docs {
		if e := s.PutDoc(ctx, &doc); e != nil {
			return e
		}
	}
	return nil
}
