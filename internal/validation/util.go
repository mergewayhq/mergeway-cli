package validation

import (
	"fmt"
	"path/filepath"

	"github.com/mergewayhq/mergeway-cli/internal/scalar"
)

func appendFiltered(dst []Error, errs []Error, phases map[Phase]bool, p Phase) []Error {
	if phases[p] {
		dst = append(dst, errs...)
	}
	return dst
}

func normalizePhases(phases []Phase) map[Phase]bool {
	if len(phases) == 0 {
		return map[Phase]bool{
			PhaseFormat:     true,
			PhaseSchema:     true,
			PhaseReferences: true,
		}
	}

	set := make(map[Phase]bool, len(phases))
	for _, p := range phases {
		set[p] = true
	}
	return set
}

func relPath(root, path string) string {
	rel, err := filepath.Rel(root, path)
	if err != nil {
		return path
	}
	return filepath.ToSlash(rel)
}

func getString(m map[string]any, key string) (string, bool) {
	if m == nil {
		return "", false
	}
	value, ok := m[key]
	if !ok {
		return "", false
	}
	return scalar.AsString(value)
}

func toSliceMap(value any) ([]map[string]any, error) {
	slice, ok := value.([]any)
	if !ok {
		return nil, fmt.Errorf("items must be an array, got %T", value)
	}

	result := make([]map[string]any, len(slice))
	for i, item := range slice {
		m, ok := item.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("items[%d] must be an object, got %T", i, item)
		}
		result[i] = m
	}
	return result, nil
}

// osReadFile is a test seam for injecting failures.
