package diff

import "testing"

func TestRenderDiffResultSingleModifiedObjectIsDeterministic(t *testing.T) {
	result := DiffResult{
		Entries: []DiffEntry{{
			Kind:     DiffEntryKindModified,
			Type:     "User",
			ObjectID: "User-Alice",
			FieldChanges: []DiffFieldChange{
				{
					Path:     "profile.name",
					OldValue: "Alice",
					NewValue: "Alice Example",
				},
				{
					Path:     "email",
					OldValue: "old@example.com",
					NewValue: "new@example.com",
				},
			},
		}},
	}

	got := renderDiffResult(result)
	want := "" +
		"MODIFIED User[User-Alice]\n" +
		"  email: \"old@example.com\" -> \"new@example.com\"\n" +
		"  profile.name: \"Alice\" -> \"Alice Example\"\n"
	if got != want {
		t.Fatalf("unexpected rendered diff\nwant:\n%s\ngot:\n%s", want, got)
	}
}

func TestRenderDiffResultMultipleEntriesIsDeterministic(t *testing.T) {
	result := DiffResult{
		Entries: []DiffEntry{
			{
				Kind:     DiffEntryKindAdded,
				Type:     "Tag",
				ObjectID: "Tag-New",
				NewSources: []LogicalObjectSource{
					{Path: "data/tags/tag-new.yaml"},
				},
			},
			{
				Kind:     DiffEntryKindModified,
				Type:     "User",
				ObjectID: "User-Alice",
				FieldChanges: []DiffFieldChange{
					{
						Path:     "name",
						OldValue: "Alice",
						NewValue: "Alice Updated",
					},
				},
			},
			{
				Kind:     DiffEntryKindRemoved,
				Type:     "Tag",
				ObjectID: "Tag-Old",
				OldSources: []LogicalObjectSource{
					{Path: "data/tags/tag-old.yaml"},
				},
			},
		},
	}

	got := renderDiffResult(result)
	want := "" +
		"ADDED Tag[Tag-New]\n" +
		"  at: data/tags/tag-new.yaml\n" +
		"\n" +
		"MODIFIED User[User-Alice]\n" +
		"  name: \"Alice\" -> \"Alice Updated\"\n" +
		"\n" +
		"REMOVED Tag[Tag-Old]\n" +
		"  from: data/tags/tag-old.yaml\n"
	if got != want {
		t.Fatalf("unexpected rendered diff\nwant:\n%s\ngot:\n%s", want, got)
	}
}

func TestRenderDiffResultEmptyDiffIsStable(t *testing.T) {
	got := renderDiffResult(DiffResult{})
	want := "No changes.\n"
	if got != want {
		t.Fatalf("unexpected rendered diff\nwant:\n%s\ngot:\n%s", want, got)
	}
}

func TestRenderDiffResultAddRemoveRenderingIsStable(t *testing.T) {
	result := DiffResult{
		Entries: []DiffEntry{
			{
				Kind:     DiffEntryKindAdded,
				Type:     "Service",
				ObjectID: "svc-new",
				NewSources: []LogicalObjectSource{
					{Path: "data/services/new.yaml"},
				},
			},
			{
				Kind:     DiffEntryKindRemoved,
				Type:     "Service",
				ObjectID: "svc-old",
				OldSources: []LogicalObjectSource{
					{Path: "data/services/old.yaml"},
				},
			},
		},
	}

	got := renderDiffResult(result)
	want := "" +
		"ADDED Service[svc-new]\n" +
		"  at: data/services/new.yaml\n" +
		"\n" +
		"REMOVED Service[svc-old]\n" +
		"  from: data/services/old.yaml\n"
	if got != want {
		t.Fatalf("unexpected rendered diff\nwant:\n%s\ngot:\n%s", want, got)
	}
}

func TestRenderDiffResultRelocationRenderingIsStable(t *testing.T) {
	result := DiffResult{
		Entries: []DiffEntry{{
			Kind:     DiffEntryKindRelocated,
			Type:     "Customer",
			ObjectID: "42",
			OldSources: []LogicalObjectSource{
				{Path: "data/customers.yaml"},
			},
			NewSources: []LogicalObjectSource{
				{Path: "archive/customers.yaml"},
			},
		}},
	}

	got := renderDiffResult(result)
	want := "" +
		"RELOCATED Customer[42]\n" +
		"  from: data/customers.yaml\n" +
		"  to: archive/customers.yaml\n"
	if got != want {
		t.Fatalf("unexpected rendered diff\nwant:\n%s\ngot:\n%s", want, got)
	}
}
