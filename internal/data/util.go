package data

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mergewayhq/mergeway-cli/internal/scalar"
)

func detectFormat(path string) fileFormat {
	ext := strings.ToLower(filepath.Ext(path))
	if ext == ".json" {
		return formatJSON
	}
	return formatYAML
}

func ensureDir(path string) error {
	if err := os.MkdirAll(path, 0o755); err != nil {
		return fmt.Errorf("data: mkdir %s: %w", path, err)
	}
	return nil
}

func requiredString(fields map[string]any, key string) (string, error) {
	if fields == nil {
		return "", fmt.Errorf("missing field %q", key)
	}
	value, ok := fields[key]
	if !ok {
		return "", fmt.Errorf("missing field %q", key)
	}
	str, ok := scalar.AsString(value)
	if !ok {
		return "", fmt.Errorf("field %q must be a non-empty string or number", key)
	}
	return str, nil
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
		return nil, fmt.Errorf("expected array, got %T", value)
	}

	result := make([]map[string]any, len(slice))
	for i, item := range slice {
		m, ok := item.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("expected object at index %d, got %T", i, item)
		}
		result[i] = cloneMap(m)
	}
	return result, nil
}

func cloneMap(m map[string]any) map[string]any {
	if m == nil {
		return nil
	}
	copy := make(map[string]any, len(m))
	for k, v := range m {
		copy[k] = cloneValue(v)
	}
	return copy
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

func mergeMaps(dst map[string]any, src map[string]any) {
	if src == nil {
		return
	}
	for k, v := range src {
		existing, exists := dst[k]
		if !exists {
			dst[k] = cloneValue(v)
			continue
		}

		mapDst, okDst := existing.(map[string]any)
		mapSrc, okSrc := v.(map[string]any)
		if okDst && okSrc {
			mergeMaps(mapDst, mapSrc)
			dst[k] = mapDst
			continue
		}

		dst[k] = cloneValue(v)
	}
}

func cleanFields(fields map[string]any) map[string]any {
	data := cloneMap(fields)
	removeTypeKeys(data)
	return data
}

func removeTypeKeys(m map[string]any) {
	if m == nil {
		return
	}
	delete(m, "type")
	delete(m, "Type")
}

func sanitizeFilename(value string) string {
	if value == "" {
		return "object"
	}
	var builder strings.Builder
	for _, r := range value {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9', r == '-', r == '_', r == '.':
			builder.WriteRune(r)
		default:
			builder.WriteByte('-')
		}
	}
	return builder.String()
}

func replaceGlob(pattern, value string) string {
	if strings.Contains(pattern, "*") {
		pattern = strings.ReplaceAll(pattern, "*", value)
	}
	if strings.Contains(pattern, "?") {
		pattern = strings.ReplaceAll(pattern, "?", value)
	}
	return pattern
}
