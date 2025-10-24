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

func normalizeYAMLValue(value any) (any, error) {
	switch v := value.(type) {
	case map[string]any:
		result := make(map[string]any, len(v))
		for key, child := range v {
			normalized, err := normalizeYAMLValue(child)
			if err != nil {
				return nil, err
			}
			result[key] = normalized
		}
		return result, nil
	case map[any]any:
		result := make(map[string]any, len(v))
		for key, child := range v {
			strKey, ok := key.(string)
			if !ok {
				return nil, fmt.Errorf("expected string map key, got %T", key)
			}
			normalized, err := normalizeYAMLValue(child)
			if err != nil {
				return nil, err
			}
			result[strKey] = normalized
		}
		return result, nil
	case []any:
		result := make([]any, len(v))
		for i, item := range v {
			normalized, err := normalizeYAMLValue(item)
			if err != nil {
				return nil, err
			}
			result[i] = normalized
		}
		return result, nil
	default:
		return v, nil
	}
}

func normalizeObject(value any) (map[string]any, error) {
	normalized, err := normalizeYAMLValue(value)
	if err != nil {
		return nil, err
	}

	obj, ok := normalized.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("expected object, got %T", normalized)
	}
	return obj, nil
}

func removeTypeKeys(m map[string]any) {
	if m == nil {
		return
	}
	delete(m, "type")
	delete(m, "Type")
}

func cloneMap(m map[string]any) map[string]any {
	if m == nil {
		return nil
	}
	dup := make(map[string]any, len(m))
	for k, v := range m {
		dup[k] = cloneValue(v)
	}
	return dup
}

func cloneValue(v any) any {
	switch val := v.(type) {
	case map[string]any:
		return cloneMap(val)
	case []any:
		res := make([]any, len(val))
		for i, item := range val {
			res[i] = cloneValue(item)
		}
		return res
	default:
		return val
	}
}

// osReadFile is a test seam for injecting failures.
