package collection

import (
	"fmt"
	"github.com/example/multitenant-search/internal/platform"
	"strings"
)

func ValidateMapping(m map[string]platform.FieldMapping) error {
	valid := map[platform.FieldType]bool{platform.Text: true, platform.Keyword: true, platform.Integer: true, platform.Float: true, platform.Boolean: true, platform.Timestamp: true, platform.GeoPoint: true}
	for n, f := range m {
		if strings.TrimSpace(n) == "" || !valid[f.Type] {
			return fmt.Errorf("field %s has unsupported type %s: %w", n, f.Type, platform.ErrInvalid)
		}
	}
	return nil
}
func Compatible(old, next map[string]platform.FieldMapping) error {
	for n, f := range old {
		if x, ok := next[n]; ok && x.Type != f.Type {
			return fmt.Errorf("field %s changes %s to %s: %w", n, f.Type, x.Type, platform.ErrConflict)
		}
	}
	return nil
}

func ValidateAndExplain(old, next map[string]platform.FieldMapping) error {
	if e := Compatible(old, next); e != nil {
		return fmt.Errorf("mapping compatibility: %v", e)
	}
	if e := ValidateMapping(next); e != nil {
		return fmt.Errorf("mapping validation: %w", e)
	}
	return nil
}
