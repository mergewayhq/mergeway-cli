package diff

import (
	"encoding/json"
	"sort"
)

type diffJSONDocument struct {
	Version int             `json:"version"`
	Entries []diffJSONEntry `json:"entries"`
}

type diffJSONEntry struct {
	Kind      DiffEntryKind         `json:"kind"`
	Type      string                `json:"type"`
	ObjectID  string                `json:"object_id"`
	Value     map[string]any        `json:"value,omitempty"`
	OldValue  map[string]any        `json:"old_value,omitempty"`
	NewValue  map[string]any        `json:"new_value,omitempty"`
	Changes   []diffJSONFieldChange `json:"changes,omitempty"`
	Sources   []diffJSONSource      `json:"sources,omitempty"`
	OldSource []diffJSONSource      `json:"old_sources,omitempty"`
	NewSource []diffJSONSource      `json:"new_sources,omitempty"`
}

type diffJSONFieldChange struct {
	Path   string `json:"path"`
	Before any    `json:"before"`
	After  any    `json:"after"`
}

type diffJSONSource struct {
	Path     string `json:"path"`
	Selector string `json:"selector,omitempty"`
	ReadOnly bool   `json:"read_only,omitempty"`
}

func marshalDiffResultJSON(result DiffResult) ([]byte, error) {
	doc := diffJSONDocument{
		Version: 1,
		Entries: make([]diffJSONEntry, 0, len(result.Entries)),
	}

	for _, entry := range result.Entries {
		doc.Entries = append(doc.Entries, diffJSONEntryFromDiffEntry(entry))
	}

	return json.MarshalIndent(doc, "", "  ")
}

func diffJSONEntryFromDiffEntry(entry DiffEntry) diffJSONEntry {
	out := diffJSONEntry{
		Kind:     entry.Kind,
		Type:     entry.Type,
		ObjectID: entry.ObjectID,
	}

	switch entry.Kind {
	case DiffEntryKindAdded:
		out.Value = cloneMap(entry.NewValue)
		out.Sources = jsonSources(entry.NewSources)
	case DiffEntryKindRemoved:
		out.Value = cloneMap(entry.OldValue)
		out.Sources = jsonSources(entry.OldSources)
	case DiffEntryKindModified:
		out.OldValue = cloneMap(entry.OldValue)
		out.NewValue = cloneMap(entry.NewValue)
		out.Changes = jsonFieldChanges(entry.FieldChanges)
		out.OldSource = jsonSources(entry.OldSources)
		out.NewSource = jsonSources(entry.NewSources)
	case DiffEntryKindRelocated:
		out.Value = cloneMap(entry.NewValue)
		if out.Value == nil {
			out.Value = cloneMap(entry.OldValue)
		}
		out.OldSource = jsonSources(entry.OldSources)
		out.NewSource = jsonSources(entry.NewSources)
	default:
		out.OldValue = cloneMap(entry.OldValue)
		out.NewValue = cloneMap(entry.NewValue)
		out.Changes = jsonFieldChanges(entry.FieldChanges)
		out.OldSource = jsonSources(entry.OldSources)
		out.NewSource = jsonSources(entry.NewSources)
	}

	return out
}

func jsonFieldChanges(changes []DiffFieldChange) []diffJSONFieldChange {
	if len(changes) == 0 {
		return nil
	}

	sortedChanges := sortedFieldChanges(changes)
	out := make([]diffJSONFieldChange, 0, len(sortedChanges))
	for _, change := range sortedChanges {
		out = append(out, diffJSONFieldChange{
			Path:   change.Path,
			Before: cloneValue(change.OldValue),
			After:  cloneValue(change.NewValue),
		})
	}
	return out
}

func jsonSources(sources []LogicalObjectSource) []diffJSONSource {
	if len(sources) == 0 {
		return nil
	}

	sortedSources := append([]LogicalObjectSource(nil), sources...)
	sort.Slice(sortedSources, func(i, j int) bool {
		if sortedSources[i].Path != sortedSources[j].Path {
			return sortedSources[i].Path < sortedSources[j].Path
		}
		if sortedSources[i].Selector != sortedSources[j].Selector {
			return sortedSources[i].Selector < sortedSources[j].Selector
		}
		if sortedSources[i].ReadOnly == sortedSources[j].ReadOnly {
			return false
		}
		return !sortedSources[i].ReadOnly && sortedSources[j].ReadOnly
	})

	out := make([]diffJSONSource, 0, len(sortedSources))
	for _, source := range sortedSources {
		out = append(out, diffJSONSource(source))
	}
	return out
}
