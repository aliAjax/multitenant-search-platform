package collection

import (
	"errors"
	"fmt"

	"github.com/example/multitenant-search/internal/platform"
)

func classifyMappingError(e error) error {
	if e == nil {
		return nil
	}
	if errors.Is(e, platform.ErrConflict) {
		return fmt.Errorf("mapping conflict: %w", e)
	}
	return fmt.Errorf("mapping rejected: %w", e)
}
