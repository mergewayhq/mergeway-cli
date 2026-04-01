package diff

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

func renderDiffResult(result DiffResult) string {
	if len(result.Entries) == 0 {
		return "No changes.\n"
	}

	var b strings.Builder
	for idx, entry := range result.Entries {
		if idx > 0 {
			b.WriteByte('\n')
		}

		switch entry.Kind {
		case DiffEntryKindAdded:
			fmt.Fprintf(&b, "ADDED %s[%s]\n", entry.Type, entry.ObjectID)
			fmt.Fprintf(&b, "  at: %s\n", summarizeSources(entry.NewSources))
		case DiffEntryKindRemoved:
			fmt.Fprintf(&b, "REMOVED %s[%s]\n", entry.Type, entry.ObjectID)
			fmt.Fprintf(&b, "  from: %s\n", summarizeSources(entry.OldSources))
		case DiffEntryKindModified:
			fmt.Fprintf(&b, "MODIFIED %s[%s]\n", entry.Type, entry.ObjectID)
			if !semanticSourcesEqual(entry.OldSources, entry.NewSources) {
				fmt.Fprintf(&b, "  from: %s\n", summarizeSources(entry.OldSources))
				fmt.Fprintf(&b, "  to: %s\n", summarizeSources(entry.NewSources))
			}
			for _, change := range sortedFieldChanges(entry.FieldChanges) {
				fmt.Fprintf(
					&b,
					"  %s: %s -> %s\n",
					change.Path,
					formatDiffValue(change.OldValue),
					formatDiffValue(change.NewValue),
				)
			}
		case DiffEntryKindRelocated:
			fmt.Fprintf(&b, "RELOCATED %s[%s]\n", entry.Type, entry.ObjectID)
			fmt.Fprintf(&b, "  from: %s\n", summarizeSources(entry.OldSources))
			fmt.Fprintf(&b, "  to: %s\n", summarizeSources(entry.NewSources))
		default:
			fmt.Fprintf(&b, "UNKNOWN %s[%s]\n", entry.Type, entry.ObjectID)
		}
	}

	return b.String()
}

func sortedFieldChanges(changes []DiffFieldChange) []DiffFieldChange {
	if len(changes) == 0 {
		return nil
	}

	sortedChanges := append([]DiffFieldChange(nil), changes...)
	sort.Slice(sortedChanges, func(i, j int) bool {
		return sortedChanges[i].Path < sortedChanges[j].Path
	})
	return sortedChanges
}

func formatDiffValue(value any) string {
	if value == nil {
		return "null"
	}

	encoded, err := json.Marshal(value)
	if err != nil {
		return fmt.Sprintf("%v", value)
	}
	return string(encoded)
}

func summarizeSources(sources []LogicalObjectSource) string {
	if len(sources) == 0 {
		return "(unknown)"
	}

	values := make([]string, 0, len(sources))
	for _, source := range sources {
		if source.Selector != "" {
			values = append(values, source.Path+" @ "+source.Selector)
			continue
		}
		values = append(values, source.Path)
	}
	sort.Strings(values)
	return strings.Join(values, ", ")
}
