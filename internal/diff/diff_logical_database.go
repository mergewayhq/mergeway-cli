package diff

import (
	"encoding/json"
	"fmt"
	"math"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	internalconfig "github.com/mergewayhq/mergeway-cli/internal/config"
	"github.com/mergewayhq/mergeway-cli/internal/scalar"
	"github.com/theory/jsonpath"
	"gopkg.in/yaml.v3"
)

// LogicalDatabase is the normalized repository-wide data model used by diff.
//
// Invariants:
//   - object identity is derived from Mergeway semantics, never file path
//   - source paths are preserved as metadata for rendering and future merge work
//   - object ordering is deterministic so downstream diff/render steps stay stable
type LogicalDatabase struct {
	Snapshot SnapshotRef
	Objects  []LogicalObject
}

type LogicalObject struct {
	Type      string
	ID        string
	Fields    map[string]any
	Canonical string
	Sources   []LogicalObjectSource
}

type LogicalObjectSource struct {
	Path     string
	Selector string
	ReadOnly bool
}

type LogicalDatabaseErrorKind string

const (
	LogicalDatabaseErrorParse             LogicalDatabaseErrorKind = "parse"
	LogicalDatabaseErrorInvalidObject     LogicalDatabaseErrorKind = "invalid_object"
	LogicalDatabaseErrorIdentityCollision LogicalDatabaseErrorKind = "identity_collision"
)

type LogicalDatabaseBuildError struct {
	Kind     LogicalDatabaseErrorKind
	Snapshot SnapshotRef
	TypeName string
	ObjectID string
	Path     string
	Selector string
	Err      error
}

func (e *LogicalDatabaseBuildError) Error() string {
	if e == nil {
		return ""
	}

	parts := []string{fmt.Sprintf("diff: build logical database for %s", e.Snapshot)}
	if e.TypeName != "" {
		parts = append(parts, "type "+e.TypeName)
	}
	if e.ObjectID != "" {
		parts = append(parts, fmt.Sprintf("id %q", e.ObjectID))
	}
	if e.Path != "" {
		parts = append(parts, "path "+e.Path)
	}
	if e.Selector != "" {
		parts = append(parts, "selector "+strconv.Quote(e.Selector))
	}

	message := strings.Join(parts, " ")
	if e.Err != nil {
		return message + ": " + e.Err.Error()
	}
	return message
}

func (e *LogicalDatabaseBuildError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Err
}

func buildLogicalDatabase(corpus SnapshotDataCorpus) (LogicalDatabase, error) {
	if corpus.Schema == nil {
		return LogicalDatabase{}, fmt.Errorf("diff: missing snapshot schema for %s", corpus.Snapshot)
	}

	objects := make(map[string]LogicalObject)
	for _, file := range corpus.Files {
		if !file.Exists {
			continue
		}

		matches := matchingSnapshotTypeIncludes(corpus.Schema, file.Path)
		for _, match := range matches {
			parsed, err := parseLogicalObjectsFromFile(corpus.Snapshot, match.Type, match.Include, file)
			if err != nil {
				return LogicalDatabase{}, err
			}
			for _, obj := range parsed {
				key := logicalObjectMapKey(obj.Type, obj.ID)
				if existing, ok := objects[key]; ok {
					return LogicalDatabase{}, &LogicalDatabaseBuildError{
						Kind:     LogicalDatabaseErrorIdentityCollision,
						Snapshot: corpus.Snapshot,
						TypeName: obj.Type,
						ObjectID: obj.ID,
						Path:     file.Path,
						Selector: match.Include.Selector,
						Err: fmt.Errorf(
							"object already defined at %s",
							existing.Sources[0].Path,
						),
					}
				}
				objects[key] = obj
			}
		}
	}

	result := LogicalDatabase{
		Snapshot: corpus.Snapshot,
		Objects:  make([]LogicalObject, 0, len(objects)),
	}
	for _, obj := range objects {
		result.Objects = append(result.Objects, obj)
	}
	sort.Slice(result.Objects, func(i, j int) bool {
		if result.Objects[i].Type != result.Objects[j].Type {
			return result.Objects[i].Type < result.Objects[j].Type
		}
		if result.Objects[i].ID != result.Objects[j].ID {
			return result.Objects[i].ID < result.Objects[j].ID
		}
		return result.Objects[i].Canonical < result.Objects[j].Canonical
	})

	return result, nil
}

type snapshotTypeIncludeMatch struct {
	Type    *diffSnapshotType
	Include diffSnapshotInclude
}

func matchingSnapshotTypeIncludes(schema *diffSnapshotSchema, path string) []snapshotTypeIncludeMatch {
	if schema == nil {
		return nil
	}

	var matches []snapshotTypeIncludeMatch
	for _, typeName := range sortedSchemaTypeNames(schema) {
		typeDef := schema.Types[typeName]
		for _, include := range typeDef.Includes {
			ok, err := filepathMatch(include.Path, path)
			if err != nil || !ok {
				continue
			}
			matches = append(matches, snapshotTypeIncludeMatch{
				Type:    typeDef,
				Include: include,
			})
		}
	}
	return matches
}

func sortedSchemaTypeNames(schema *diffSnapshotSchema) []string {
	names := make([]string, 0, len(schema.Types))
	for name := range schema.Types {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func parseLogicalObjectsFromFile(snapshot SnapshotRef, typeDef *diffSnapshotType, include diffSnapshotInclude, file SnapshotDataFile) ([]LogicalObject, error) {
	if typeDef == nil {
		return nil, &LogicalDatabaseBuildError{
			Kind:     LogicalDatabaseErrorInvalidObject,
			Snapshot: snapshot,
			Path:     file.Path,
			Err:      fmt.Errorf("missing type definition"),
		}
	}

	rawObjects, err := parseIncludedSnapshotObjects(typeDef, include, file)
	if err != nil {
		return nil, &LogicalDatabaseBuildError{
			Kind:     classifyLogicalDatabaseError(err),
			Snapshot: snapshot,
			TypeName: typeDef.Name,
			Path:     file.Path,
			Selector: include.Selector,
			Err:      err,
		}
	}

	objects := make([]LogicalObject, 0, len(rawObjects))
	for _, fields := range rawObjects {
		id, err := deriveLogicalObjectID(typeDef, fields, file.Path)
		if err != nil {
			return nil, &LogicalDatabaseBuildError{
				Kind:     LogicalDatabaseErrorInvalidObject,
				Snapshot: snapshot,
				TypeName: typeDef.Name,
				Path:     file.Path,
				Selector: include.Selector,
				Err:      err,
			}
		}

		canonical, err := canonicalizeLogicalFields(fields)
		if err != nil {
			return nil, &LogicalDatabaseBuildError{
				Kind:     LogicalDatabaseErrorInvalidObject,
				Snapshot: snapshot,
				TypeName: typeDef.Name,
				ObjectID: id,
				Path:     file.Path,
				Selector: include.Selector,
				Err:      err,
			}
		}

		objects = append(objects, LogicalObject{
			Type:      typeDef.Name,
			ID:        id,
			Fields:    cloneMap(fields),
			Canonical: canonical,
			Sources: []LogicalObjectSource{{
				Path:     file.Path,
				Selector: include.Selector,
				ReadOnly: include.Selector != "",
			}},
		})
	}

	return objects, nil
}

func parseIncludedSnapshotObjects(typeDef *diffSnapshotType, include diffSnapshotInclude, file SnapshotDataFile) ([]map[string]any, error) {
	if include.Selector == "" {
		var doc map[string]any
		if err := yaml.Unmarshal(file.Content, &doc); err != nil {
			return nil, fmt.Errorf("parse %s: %w", file.Path, err)
		}

		declaredType, hasType := getString(doc, "type")
		if hasType {
			delete(doc, "type")
		}
		if declaredType == "" {
			declaredType = typeDef.Name
		}
		if declaredType != typeDef.Name {
			return nil, fmt.Errorf("file %s declares type %s; expected %s", file.Path, declaredType, typeDef.Name)
		}

		itemsRaw, hasItems := doc["items"]
		if !hasItems {
			removeTypeKeys(doc)
			return []map[string]any{doc}, nil
		}

		if typeDef.identifierIsPath() {
			return nil, fmt.Errorf("type %s uses identifier %q, but file %s contains multiple objects", typeDef.Name, internalconfig.PathIdentifierField, file.Path)
		}

		items, err := toSliceMap(itemsRaw)
		if err != nil {
			return nil, fmt.Errorf("file %s items: %w", file.Path, err)
		}
		for _, item := range items {
			removeTypeKeys(item)
		}
		return items, nil
	}

	var root any
	if err := yaml.Unmarshal(file.Content, &root); err != nil {
		return nil, fmt.Errorf("parse %s: %w", file.Path, err)
	}

	normalizedRoot, err := normalizeYAMLValue(root)
	if err != nil {
		return nil, fmt.Errorf("normalize %s: %w", file.Path, err)
	}

	compiled, err := jsonpath.Parse(include.Selector)
	if err != nil {
		return nil, fmt.Errorf("parse selector %q in %s: %w", include.Selector, file.Path, err)
	}

	located := compiled.SelectLocated(normalizedRoot)
	if len(located) == 0 {
		return nil, fmt.Errorf("selector %q in %s matched no values", include.Selector, file.Path)
	}

	items := make([]map[string]any, 0, len(located))
	for _, node := range located {
		obj, err := normalizeObject(node.Node)
		if err != nil {
			return nil, fmt.Errorf("selector %q in %s at %s: %w", include.Selector, file.Path, node.Path.String(), err)
		}

		if declaredType, ok := getString(obj, "type"); ok && declaredType != "" && declaredType != typeDef.Name {
			return nil, fmt.Errorf("selector %q in %s at %s declares type %s; expected %s", include.Selector, file.Path, node.Path.String(), declaredType, typeDef.Name)
		}
		if declaredType, ok := getString(obj, "Type"); ok && declaredType != "" && declaredType != typeDef.Name {
			return nil, fmt.Errorf("selector %q in %s at %s declares type %s; expected %s", include.Selector, file.Path, node.Path.String(), declaredType, typeDef.Name)
		}

		removeTypeKeys(obj)
		items = append(items, obj)
	}

	if len(items) > 1 && typeDef.identifierIsPath() {
		return nil, fmt.Errorf("type %s uses identifier %q, but selector %q in %s matched multiple objects", typeDef.Name, internalconfig.PathIdentifierField, include.Selector, file.Path)
	}

	return items, nil
}

func classifyLogicalDatabaseError(err error) LogicalDatabaseErrorKind {
	if err == nil {
		return LogicalDatabaseErrorInvalidObject
	}
	message := err.Error()
	if strings.HasPrefix(message, "parse ") || strings.HasPrefix(message, "normalize ") {
		return LogicalDatabaseErrorParse
	}
	return LogicalDatabaseErrorInvalidObject
}

func deriveLogicalObjectID(typeDef *diffSnapshotType, fields map[string]any, sourcePath string) (string, error) {
	if typeDef == nil {
		return "", fmt.Errorf("type definition is required")
	}
	if typeDef.identifierIsPath() {
		return normalizeDiscoveredPathIdentifier(sourcePath)
	}

	raw, ok := fields[typeDef.IdentifierField]
	if !ok {
		return "", fmt.Errorf("missing field %q", typeDef.IdentifierField)
	}

	str, ok := scalar.AsString(raw)
	if !ok {
		return "", fmt.Errorf("field %q must be a non-empty string or number", typeDef.IdentifierField)
	}

	switch typeDef.IdentifierFieldType {
	case "integer":
		if _, ok := logicalInt64(raw); !ok {
			return "", fmt.Errorf("field %q must be an integer", typeDef.IdentifierField)
		}
	case "number":
		if _, ok := logicalFloat64(raw); !ok {
			return "", fmt.Errorf("field %q must be a number", typeDef.IdentifierField)
		}
	}

	return str, nil
}

func canonicalizeLogicalFields(fields map[string]any) (string, error) {
	data, err := json.Marshal(fields)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func logicalObjectMapKey(typeName, id string) string {
	return typeName + "\x00" + id
}

func filepathMatch(pattern, path string) (bool, error) {
	return filepath.Match(pattern, path)
}

func logicalInt64(value any) (int64, bool) {
	switch v := value.(type) {
	case int:
		return int64(v), true
	case int8:
		return int64(v), true
	case int16:
		return int64(v), true
	case int32:
		return int64(v), true
	case int64:
		return v, true
	case uint:
		if uint64(v) > math.MaxInt64 {
			return 0, false
		}
		return int64(v), true
	case uint8:
		return int64(v), true
	case uint16:
		return int64(v), true
	case uint32:
		return int64(v), true
	case uint64:
		if v > math.MaxInt64 {
			return 0, false
		}
		return int64(v), true
	case float32:
		if math.Trunc(float64(v)) != float64(v) {
			return 0, false
		}
		return int64(v), true
	case float64:
		if math.Trunc(v) != v {
			return 0, false
		}
		return int64(v), true
	case json.Number:
		if i, err := v.Int64(); err == nil {
			return i, true
		}
		if f, err := v.Float64(); err == nil && math.Trunc(f) == f {
			return int64(f), true
		}
		return 0, false
	case string:
		if v == "" {
			return 0, false
		}
		iv, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return 0, false
		}
		return iv, true
	default:
		return 0, false
	}
}

func logicalFloat64(value any) (float64, bool) {
	switch v := value.(type) {
	case float32:
		return float64(v), true
	case float64:
		if math.IsNaN(v) || math.IsInf(v, 0) {
			return 0, false
		}
		return v, true
	case int:
		return float64(v), true
	case int8:
		return float64(v), true
	case int16:
		return float64(v), true
	case int32:
		return float64(v), true
	case int64:
		return float64(v), true
	case uint:
		return float64(v), true
	case uint8:
		return float64(v), true
	case uint16:
		return float64(v), true
	case uint32:
		return float64(v), true
	case uint64:
		return float64(v), true
	case json.Number:
		f, err := v.Float64()
		if err != nil {
			return 0, false
		}
		if math.IsNaN(f) || math.IsInf(f, 0) {
			return 0, false
		}
		return f, true
	case string:
		if v == "" {
			return 0, false
		}
		f, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return 0, false
		}
		if math.IsNaN(f) || math.IsInf(f, 0) {
			return 0, false
		}
		return f, true
	default:
		return 0, false
	}
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

func removeTypeKeys(m map[string]any) {
	if m == nil {
		return
	}
	delete(m, "type")
	delete(m, "Type")
	delete(m, internalconfig.PathIdentifierField)
}

func normalizeDiscoveredPathIdentifier(value string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", fmt.Errorf("field %q must be a non-empty string", internalconfig.PathIdentifierField)
	}
	if filepath.IsAbs(value) {
		return "", fmt.Errorf("field %q must be a relative path", internalconfig.PathIdentifierField)
	}
	cleaned := filepath.Clean(filepath.FromSlash(value))
	if cleaned == "." || cleaned == "" {
		return "", fmt.Errorf("field %q must be a non-empty string", internalconfig.PathIdentifierField)
	}
	return filepath.ToSlash(cleaned), nil
}
