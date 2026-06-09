package cmdutil

import (
	"fmt"

	"github.com/atlanticbt/magecli/pkg/magento"
)

// ValidateLimit checks a --limit flag value against the supported page-size
// range (1 to magento.MaxPageSize).
func ValidateLimit(limit int) error {
	if limit < 1 || limit > magento.MaxPageSize {
		return fmt.Errorf("--limit must be between 1 and %d (got %d)", magento.MaxPageSize, limit)
	}
	return nil
}

// Truncate shortens s to at most max runes for table display, appending "..."
// when content is dropped. For max <= 3 there is no room for an ellipsis, so
// it returns the first max runes.
func Truncate(s string, max int) string {
	if len(s) <= max { // fast path: byte length bounds rune length
		return s
	}
	r := []rune(s)
	if len(r) <= max {
		return s
	}
	if max <= 3 {
		return string(r[:max])
	}
	return string(r[:max-3]) + "..."
}
