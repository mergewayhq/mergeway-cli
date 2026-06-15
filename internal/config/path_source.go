package config

import (
	"fmt"
	"path/filepath"
	"strings"
)

// DerivePathSourceValue resolves a declared field source from a normalized relative file path.
func DerivePathSourceValue(source *FieldSourceDefinition, path string) (string, error) {
	if source == nil || !source.IsPathDerived() {
		return "", nil
	}

	path = strings.TrimSpace(path)
	if path == "" {
		return "", fmt.Errorf("field source requires a file-backed object")
	}

	cleaned := filepath.ToSlash(filepath.Clean(filepath.FromSlash(path)))
	if cleaned == "." || cleaned == "" {
		return "", fmt.Errorf("field source requires a file-backed object")
	}
	if source.Path {
		return cleaned, nil
	}

	segments := strings.Split(cleaned, "/")
	if source.PathSegment != nil {
		index := *source.PathSegment
		if index >= len(segments) {
			return "", fmt.Errorf("field source path_segment %d is out of range for %q", index, cleaned)
		}
		return segments[index], nil
	}
	if source.PathSegmentRev != nil {
		index := *source.PathSegmentRev
		if index >= len(segments) {
			return "", fmt.Errorf("field source path_segment_rev %d is out of range for %q", index, cleaned)
		}
		return segments[len(segments)-1-index], nil
	}

	return "", nil
}
