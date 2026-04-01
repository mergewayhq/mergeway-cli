package diff

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

type DiffResult struct {
	Entries []DiffEntry
}

type DiffEntryKind string

const (
	DiffEntryKindAdded     DiffEntryKind = "added"
	DiffEntryKindRemoved   DiffEntryKind = "removed"
	DiffEntryKindModified  DiffEntryKind = "modified"
	DiffEntryKindRelocated DiffEntryKind = "relocated"
)

type DiffEntry struct {
	Kind         DiffEntryKind
	Type         string
	ObjectID     string
	OldValue     map[string]any
	NewValue     map[string]any
	FieldChanges []DiffFieldChange
	OldSources   []LogicalObjectSource
	NewSources   []LogicalObjectSource
}

type DiffFieldChange struct {
	Path     string
	OldValue any
	NewValue any
}

// diffLogicalDatabases converts two normalized logical databases into semantic
// facts that are stable enough to feed rendering, automation, and later merge
// planning.
//
// Invariants:
//   - configuration changes are excluded before this layer
//   - object comparison is keyed by logical identity, not source path
//   - path-only moves are preserved as relocation metadata and never reported
//     as value modification
//   - entry ordering is deterministic across runs
func diffLogicalDatabases(left, right LogicalDatabase) (DiffResult, error) {
	leftByKey := logicalObjectsByKey(left)
	rightByKey := logicalObjectsByKey(right)

	keys := make(map[string]struct{}, len(leftByKey)+len(rightByKey))
	for key := range leftByKey {
		keys[key] = struct{}{}
	}
	for key := range rightByKey {
		keys[key] = struct{}{}
	}

	sortedKeys := sortedKeys(keys)
	result := DiffResult{
		Entries: make([]DiffEntry, 0, len(sortedKeys)),
	}
	for _, key := range sortedKeys {
		leftObj, leftOK := leftByKey[key]
		rightObj, rightOK := rightByKey[key]

		switch {
		case leftOK && !rightOK:
			result.Entries = append(result.Entries, DiffEntry{
				Kind:       DiffEntryKindRemoved,
				Type:       leftObj.Type,
				ObjectID:   leftObj.ID,
				OldValue:   cloneMap(leftObj.Fields),
				OldSources: cloneLogicalSources(leftObj.Sources),
			})
		case !leftOK && rightOK:
			result.Entries = append(result.Entries, DiffEntry{
				Kind:       DiffEntryKindAdded,
				Type:       rightObj.Type,
				ObjectID:   rightObj.ID,
				NewValue:   cloneMap(rightObj.Fields),
				NewSources: cloneLogicalSources(rightObj.Sources),
			})
		case leftOK && rightOK:
			fieldChanges, err := diffObjectFields("", leftObj.Fields, rightObj.Fields)
			if err != nil {
				return DiffResult{}, fmt.Errorf("diff: compare %s %q: %w", leftObj.Type, leftObj.ID, err)
			}

			relocated := !semanticSourcesEqual(leftObj.Sources, rightObj.Sources)
			switch {
			case len(fieldChanges) > 0:
				result.Entries = append(result.Entries, DiffEntry{
					Kind:         DiffEntryKindModified,
					Type:         leftObj.Type,
					ObjectID:     leftObj.ID,
					OldValue:     cloneMap(leftObj.Fields),
					NewValue:     cloneMap(rightObj.Fields),
					FieldChanges: fieldChanges,
					OldSources:   cloneLogicalSources(leftObj.Sources),
					NewSources:   cloneLogicalSources(rightObj.Sources),
				})
			case relocated:
				result.Entries = append(result.Entries, DiffEntry{
					Kind:       DiffEntryKindRelocated,
					Type:       leftObj.Type,
					ObjectID:   leftObj.ID,
					OldValue:   cloneMap(leftObj.Fields),
					NewValue:   cloneMap(rightObj.Fields),
					OldSources: cloneLogicalSources(leftObj.Sources),
					NewSources: cloneLogicalSources(rightObj.Sources),
				})
			}
		}
	}

	return result, nil
}

func logicalObjectsByKey(db LogicalDatabase) map[string]LogicalObject {
	objects := make(map[string]LogicalObject, len(db.Objects))
	for _, obj := range db.Objects {
		objects[logicalObjectMapKey(obj.Type, obj.ID)] = obj
	}
	return objects
}

func diffObjectFields(prefix string, left, right map[string]any) ([]DiffFieldChange, error) {
	keySet := make(map[string]struct{}, len(left)+len(right))
	for key := range left {
		keySet[key] = struct{}{}
	}
	for key := range right {
		keySet[key] = struct{}{}
	}

	keys := sortedKeys(keySet)
	changes := make([]DiffFieldChange, 0)
	for _, key := range keys {
		path := key
		if prefix != "" {
			path = prefix + "." + key
		}

		leftValue, leftOK := left[key]
		rightValue, rightOK := right[key]
		switch {
		case !leftOK:
			changes = append(changes, DiffFieldChange{
				Path:     path,
				OldValue: nil,
				NewValue: cloneValue(rightValue),
			})
		case !rightOK:
			changes = append(changes, DiffFieldChange{
				Path:     path,
				OldValue: cloneValue(leftValue),
				NewValue: nil,
			})
		default:
			fieldChanges, err := diffLogicalValues(path, leftValue, rightValue)
			if err != nil {
				return nil, err
			}
			changes = append(changes, fieldChanges...)
		}
	}

	return changes, nil
}

func diffLogicalValues(path string, left, right any) ([]DiffFieldChange, error) {
	equal, err := semanticValuesEqual(left, right)
	if err != nil {
		return nil, err
	}
	if equal {
		return nil, nil
	}

	leftMap, leftIsMap := left.(map[string]any)
	rightMap, rightIsMap := right.(map[string]any)
	if leftIsMap && rightIsMap {
		return diffObjectFields(path, leftMap, rightMap)
	}

	return []DiffFieldChange{{
		Path:     path,
		OldValue: cloneValue(left),
		NewValue: cloneValue(right),
	}}, nil
}

func semanticValuesEqual(left, right any) (bool, error) {
	leftCanonical, err := semanticCanonicalValue(left)
	if err != nil {
		return false, err
	}
	rightCanonical, err := semanticCanonicalValue(right)
	if err != nil {
		return false, err
	}
	return leftCanonical == rightCanonical, nil
}

func semanticCanonicalValue(value any) (string, error) {
	switch v := value.(type) {
	case nil:
		return "null", nil
	case map[string]any:
		keys := make([]string, 0, len(v))
		for key := range v {
			keys = append(keys, key)
		}
		sort.Strings(keys)

		parts := make([]string, 0, len(keys))
		for _, key := range keys {
			childCanonical, err := semanticCanonicalValue(v[key])
			if err != nil {
				return "", err
			}
			parts = append(parts, key+":"+childCanonical)
		}
		return "{" + strings.Join(parts, ",") + "}", nil
	case []any:
		items := make([]string, 0, len(v))
		for _, item := range v {
			itemCanonical, err := semanticCanonicalValue(item)
			if err != nil {
				return "", err
			}
			items = append(items, itemCanonical)
		}
		sort.Strings(items)
		return "[" + strings.Join(items, ",") + "]", nil
	default:
		data, err := json.Marshal(v)
		if err != nil {
			return "", err
		}
		return string(data), nil
	}
}

func semanticSourcesEqual(left, right []LogicalObjectSource) bool {
	if len(left) != len(right) {
		return false
	}
	if len(left) == 0 {
		return true
	}

	leftValues := logicalSourceKeys(left)
	rightValues := logicalSourceKeys(right)
	for idx := range leftValues {
		if leftValues[idx] != rightValues[idx] {
			return false
		}
	}
	return true
}

func logicalSourceKeys(sources []LogicalObjectSource) []string {
	values := make([]string, 0, len(sources))
	for _, source := range sources {
		values = append(values, source.Path+"\x00"+source.Selector+"\x00"+fmt.Sprintf("%t", source.ReadOnly))
	}
	sort.Strings(values)
	return values
}

func cloneLogicalSources(sources []LogicalObjectSource) []LogicalObjectSource {
	if len(sources) == 0 {
		return nil
	}
	cloned := make([]LogicalObjectSource, len(sources))
	copy(cloned, sources)
	return cloned
}
