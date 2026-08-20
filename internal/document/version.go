package document

import (
	"fmt"
	"github.com/example/multitenant-search/internal/platform"
)

func CheckVersion(existing *platform.Document, incoming platform.Document) error {
	if existing == nil {
		return nil
	}
	if incoming.Version > 0 && incoming.Version <= existing.Version {
		return fmt.Errorf("incoming version %d <= current %d: %w", incoming.Version, existing.Version, platform.ErrConflict)
	}
	if incoming.ExternalTS > 0 && incoming.ExternalTS < existing.ExternalTS {
		return fmt.Errorf("external timestamp regressed: %w", platform.ErrConflict)
	}
	return nil
}
func MergePartial(existing platform.Document, patch map[string]any) platform.Document {
	if existing.Data == nil {
		existing.Data = map[string]any{}
	}
	for k, v := range patch {
		existing.Data[k] = v
	}
	existing.Version++
	return existing
}
