package diff

import (
	"reflect"
	"testing"
)

func TestDiffLogicalDatabasesDetectsAddedObject(t *testing.T) {
	left := logicalDatabaseFixture()
	right := logicalDatabaseFixture(
		logicalObjectFixture("Tag", "Tag-New", map[string]any{
			"id":    "Tag-New",
			"label": "New",
		}, "data/tags/tag-new.yaml"),
	)

	result := mustDiffLogicalDatabases(t, left, right)
	expected := DiffResult{
		Entries: []DiffEntry{{
			Kind:     DiffEntryKindAdded,
			Type:     "Tag",
			ObjectID: "Tag-New",
			NewValue: map[string]any{
				"id":    "Tag-New",
				"label": "New",
			},
			NewSources: []LogicalObjectSource{{
				Path: "data/tags/tag-new.yaml",
			}},
		}},
	}
	assertDiffResultEqual(t, result, expected)
}

func TestDiffLogicalDatabasesDetectsRemovedObject(t *testing.T) {
	left := logicalDatabaseFixture(
		logicalObjectFixture("Tag", "Tag-Old", map[string]any{
			"id":    "Tag-Old",
			"label": "Old",
		}, "data/tags/tag-old.yaml"),
	)
	right := logicalDatabaseFixture()

	result := mustDiffLogicalDatabases(t, left, right)
	expected := DiffResult{
		Entries: []DiffEntry{{
			Kind:     DiffEntryKindRemoved,
			Type:     "Tag",
			ObjectID: "Tag-Old",
			OldValue: map[string]any{
				"id":    "Tag-Old",
				"label": "Old",
			},
			OldSources: []LogicalObjectSource{{
				Path: "data/tags/tag-old.yaml",
			}},
		}},
	}
	assertDiffResultEqual(t, result, expected)
}

func TestDiffLogicalDatabasesDetectsScalarFieldModification(t *testing.T) {
	left := logicalDatabaseFixture(
		logicalObjectFixture("User", "User-Alice", map[string]any{
			"id":    "User-Alice",
			"name":  "Alice",
			"email": "alice@example.com",
		}, "data/users/user-alice.yaml"),
	)
	right := logicalDatabaseFixture(
		logicalObjectFixture("User", "User-Alice", map[string]any{
			"id":    "User-Alice",
			"name":  "Alice Updated",
			"email": "alice@example.com",
		}, "data/users/user-alice.yaml"),
	)

	result := mustDiffLogicalDatabases(t, left, right)
	expected := DiffResult{
		Entries: []DiffEntry{{
			Kind:     DiffEntryKindModified,
			Type:     "User",
			ObjectID: "User-Alice",
			OldValue: map[string]any{
				"id":    "User-Alice",
				"name":  "Alice",
				"email": "alice@example.com",
			},
			NewValue: map[string]any{
				"id":    "User-Alice",
				"name":  "Alice Updated",
				"email": "alice@example.com",
			},
			FieldChanges: []DiffFieldChange{{
				Path:     "name",
				OldValue: "Alice",
				NewValue: "Alice Updated",
			}},
			OldSources: []LogicalObjectSource{{Path: "data/users/user-alice.yaml"}},
			NewSources: []LogicalObjectSource{{Path: "data/users/user-alice.yaml"}},
		}},
	}
	assertDiffResultEqual(t, result, expected)
}

func TestDiffLogicalDatabasesDetectsNestedFieldModification(t *testing.T) {
	left := logicalDatabaseFixture(
		logicalObjectFixture("User", "User-Alice", map[string]any{
			"id": "User-Alice",
			"profile": map[string]any{
				"contact": map[string]any{
					"email": "alice@example.com",
				},
			},
		}, "data/users/user-alice.yaml"),
	)
	right := logicalDatabaseFixture(
		logicalObjectFixture("User", "User-Alice", map[string]any{
			"id": "User-Alice",
			"profile": map[string]any{
				"contact": map[string]any{
					"email": "alice+updated@example.com",
				},
			},
		}, "data/users/user-alice.yaml"),
	)

	result := mustDiffLogicalDatabases(t, left, right)
	expected := DiffResult{
		Entries: []DiffEntry{{
			Kind:     DiffEntryKindModified,
			Type:     "User",
			ObjectID: "User-Alice",
			OldValue: map[string]any{
				"id": "User-Alice",
				"profile": map[string]any{
					"contact": map[string]any{
						"email": "alice@example.com",
					},
				},
			},
			NewValue: map[string]any{
				"id": "User-Alice",
				"profile": map[string]any{
					"contact": map[string]any{
						"email": "alice+updated@example.com",
					},
				},
			},
			FieldChanges: []DiffFieldChange{{
				Path:     "profile.contact.email",
				OldValue: "alice@example.com",
				NewValue: "alice+updated@example.com",
			}},
			OldSources: []LogicalObjectSource{{Path: "data/users/user-alice.yaml"}},
			NewSources: []LogicalObjectSource{{Path: "data/users/user-alice.yaml"}},
		}},
	}
	assertDiffResultEqual(t, result, expected)
}

func TestDiffLogicalDatabasesTreatsPathOnlyMoveAsRelocation(t *testing.T) {
	left := logicalDatabaseFixture(
		logicalObjectFixture("User", "User-Bob", map[string]any{
			"id":   "User-Bob",
			"name": "Bob",
		}, "data/users/user-bob.yaml"),
	)
	right := logicalDatabaseFixture(
		logicalObjectFixture("User", "User-Bob", map[string]any{
			"id":   "User-Bob",
			"name": "Bob",
		}, "data/users/bob-renamed.yaml"),
	)

	result := mustDiffLogicalDatabases(t, left, right)
	expected := DiffResult{
		Entries: []DiffEntry{{
			Kind:     DiffEntryKindRelocated,
			Type:     "User",
			ObjectID: "User-Bob",
			OldValue: map[string]any{
				"id":   "User-Bob",
				"name": "Bob",
			},
			NewValue: map[string]any{
				"id":   "User-Bob",
				"name": "Bob",
			},
			OldSources: []LogicalObjectSource{{Path: "data/users/user-bob.yaml"}},
			NewSources: []LogicalObjectSource{{Path: "data/users/bob-renamed.yaml"}},
		}},
	}
	assertDiffResultEqual(t, result, expected)
}

func TestDiffLogicalDatabasesIgnoresReorderOnlyChanges(t *testing.T) {
	left := logicalDatabaseFixture(
		logicalObjectFixture("Post", "Post-001", map[string]any{
			"id": "Post-001",
			"tags": []any{
				"Tag-A",
				"Tag-B",
			},
		}, "data/posts/post-001.yaml"),
		logicalObjectFixture("User", "User-Alice", map[string]any{
			"id":   "User-Alice",
			"name": "Alice",
		}, "data/users/user-alice.yaml"),
	)
	right := logicalDatabaseFixture(
		logicalObjectFixture("User", "User-Alice", map[string]any{
			"id":   "User-Alice",
			"name": "Alice",
		}, "data/users/user-alice.yaml"),
		logicalObjectFixture("Post", "Post-001", map[string]any{
			"id": "Post-001",
			"tags": []any{
				"Tag-B",
				"Tag-A",
			},
		}, "data/posts/post-001.yaml"),
	)

	result := mustDiffLogicalDatabases(t, left, right)
	assertDiffResultEqual(t, result, DiffResult{})
}

func TestDiffLogicalDatabasesReportsMultipleIndependentChanges(t *testing.T) {
	left := logicalDatabaseFixture(
		logicalObjectFixture("Tag", "Tag-Old", map[string]any{
			"id":    "Tag-Old",
			"label": "Old",
		}, "data/tags/tag-old.yaml"),
		logicalObjectFixture("User", "User-Alice", map[string]any{
			"id":   "User-Alice",
			"name": "Alice",
		}, "data/users/user-alice.yaml"),
	)
	right := logicalDatabaseFixture(
		logicalObjectFixture("Tag", "Tag-New", map[string]any{
			"id":    "Tag-New",
			"label": "New",
		}, "data/tags/tag-new.yaml"),
		logicalObjectFixture("User", "User-Alice", map[string]any{
			"id":   "User-Alice",
			"name": "Alice Updated",
		}, "data/users/user-alice.yaml"),
	)

	result := mustDiffLogicalDatabases(t, left, right)
	expectedKeys := []string{
		"added:Tag:Tag-New",
		"removed:Tag:Tag-Old",
		"modified:User:User-Alice",
	}
	if got := diffEntryKeys(result); !reflect.DeepEqual(got, expectedKeys) {
		t.Fatalf("expected entry keys %v, got %v", expectedKeys, got)
	}
}

func TestDiffLogicalDatabasesOrdersEntriesDeterministically(t *testing.T) {
	left := logicalDatabaseFixture(
		logicalObjectFixture("User", "User-B", map[string]any{
			"id":   "User-B",
			"name": "B",
		}, "data/users/user-b.yaml"),
		logicalObjectFixture("Tag", "Tag-Z", map[string]any{
			"id":    "Tag-Z",
			"label": "Z",
		}, "data/tags/tag-z.yaml"),
	)
	right := logicalDatabaseFixture(
		logicalObjectFixture("User", "User-B", map[string]any{
			"id":   "User-B",
			"name": "B Updated",
		}, "data/users/user-b.yaml"),
		logicalObjectFixture("User", "User-A", map[string]any{
			"id":   "User-A",
			"name": "A",
		}, "data/users/user-a.yaml"),
	)

	result := mustDiffLogicalDatabases(t, left, right)
	want := []string{
		"removed:Tag:Tag-Z",
		"added:User:User-A",
		"modified:User:User-B",
	}
	if got := diffEntryKeys(result); !reflect.DeepEqual(got, want) {
		t.Fatalf("expected deterministic diff order %v, got %v", want, got)
	}
}

func TestDiffLogicalDatabasesEmptyVsEmptyIsEmpty(t *testing.T) {
	result := mustDiffLogicalDatabases(t, logicalDatabaseFixture(), logicalDatabaseFixture())
	assertDiffResultEqual(t, result, DiffResult{})
}

func TestDiffLogicalDatabasesIdenticalDatabasesYieldEmptyDiff(t *testing.T) {
	left := logicalDatabaseFixture(
		logicalObjectFixture("User", "User-Alice", map[string]any{
			"id":   "User-Alice",
			"name": "Alice",
		}, "data/users/user-alice.yaml"),
	)
	right := logicalDatabaseFixture(
		logicalObjectFixture("User", "User-Alice", map[string]any{
			"id":   "User-Alice",
			"name": "Alice",
		}, "data/users/user-alice.yaml"),
	)

	result := mustDiffLogicalDatabases(t, left, right)
	assertDiffResultEqual(t, result, DiffResult{})
}

func TestDiffLogicalDatabasesMatchesGoldenStructuredResult(t *testing.T) {
	left := logicalDatabaseFixture(
		logicalObjectFixture("Tag", "Tag-Removed", map[string]any{
			"id":    "Tag-Removed",
			"label": "Removed",
		}, "data/tags/tag-removed.yaml"),
		logicalObjectFixture("User", "User-Alice", map[string]any{
			"id":      "User-Alice",
			"name":    "Alice",
			"profile": map[string]any{"email": "alice@example.com"},
		}, "data/users/user-alice.yaml"),
		logicalObjectFixture("User", "User-Bob", map[string]any{
			"id":   "User-Bob",
			"name": "Bob",
		}, "data/users/user-bob.yaml"),
	)
	right := logicalDatabaseFixture(
		logicalObjectFixture("Tag", "Tag-Added", map[string]any{
			"id":    "Tag-Added",
			"label": "Added",
		}, "data/tags/tag-added.yaml"),
		logicalObjectFixture("User", "User-Alice", map[string]any{
			"id":      "User-Alice",
			"name":    "Alice Updated",
			"profile": map[string]any{"email": "alice+updated@example.com"},
		}, "data/users/user-alice.yaml"),
		logicalObjectFixture("User", "User-Bob", map[string]any{
			"id":   "User-Bob",
			"name": "Bob",
		}, "data/users/bob-renamed.yaml"),
	)

	result := mustDiffLogicalDatabases(t, left, right)
	expected := DiffResult{
		Entries: []DiffEntry{
			{
				Kind:     DiffEntryKindAdded,
				Type:     "Tag",
				ObjectID: "Tag-Added",
				NewValue: map[string]any{
					"id":    "Tag-Added",
					"label": "Added",
				},
				NewSources: []LogicalObjectSource{{Path: "data/tags/tag-added.yaml"}},
			},
			{
				Kind:     DiffEntryKindRemoved,
				Type:     "Tag",
				ObjectID: "Tag-Removed",
				OldValue: map[string]any{
					"id":    "Tag-Removed",
					"label": "Removed",
				},
				OldSources: []LogicalObjectSource{{Path: "data/tags/tag-removed.yaml"}},
			},
			{
				Kind:     DiffEntryKindModified,
				Type:     "User",
				ObjectID: "User-Alice",
				OldValue: map[string]any{
					"id":      "User-Alice",
					"name":    "Alice",
					"profile": map[string]any{"email": "alice@example.com"},
				},
				NewValue: map[string]any{
					"id":      "User-Alice",
					"name":    "Alice Updated",
					"profile": map[string]any{"email": "alice+updated@example.com"},
				},
				FieldChanges: []DiffFieldChange{
					{
						Path:     "name",
						OldValue: "Alice",
						NewValue: "Alice Updated",
					},
					{
						Path:     "profile.email",
						OldValue: "alice@example.com",
						NewValue: "alice+updated@example.com",
					},
				},
				OldSources: []LogicalObjectSource{{Path: "data/users/user-alice.yaml"}},
				NewSources: []LogicalObjectSource{{Path: "data/users/user-alice.yaml"}},
			},
			{
				Kind:     DiffEntryKindRelocated,
				Type:     "User",
				ObjectID: "User-Bob",
				OldValue: map[string]any{
					"id":   "User-Bob",
					"name": "Bob",
				},
				NewValue: map[string]any{
					"id":   "User-Bob",
					"name": "Bob",
				},
				OldSources: []LogicalObjectSource{{Path: "data/users/user-bob.yaml"}},
				NewSources: []LogicalObjectSource{{Path: "data/users/bob-renamed.yaml"}},
			},
		},
	}
	assertDiffResultEqual(t, result, expected)
}

func TestDiffLogicalDatabasesFromSameBasePreservesModifiedAndRelocatedFacts(t *testing.T) {
	base := logicalDatabaseFixture(
		logicalObjectFixture("User", "User-Alice", map[string]any{
			"id":    "User-Alice",
			"name":  "Alice",
			"email": "alice@example.com",
		}, "data/users/user-alice.yaml"),
	)

	modified := logicalDatabaseFixture(
		logicalObjectFixture("User", "User-Alice", map[string]any{
			"id":    "User-Alice",
			"name":  "Alice Updated",
			"email": "alice@example.com",
		}, "data/users/user-alice.yaml"),
	)

	relocated := logicalDatabaseFixture(
		logicalObjectFixture("User", "User-Alice", map[string]any{
			"id":    "User-Alice",
			"name":  "Alice",
			"email": "alice@example.com",
		}, "archive/users/user-alice.yaml"),
	)

	modifiedResult := mustDiffLogicalDatabases(t, base, modified)
	relocatedResult := mustDiffLogicalDatabases(t, base, relocated)

	if got := diffEntryKeys(modifiedResult); !reflect.DeepEqual(got, []string{"modified:User:User-Alice"}) {
		t.Fatalf("expected base->modified to stay a modification fact, got %v", got)
	}
	if got := diffEntryKeys(relocatedResult); !reflect.DeepEqual(got, []string{"relocated:User:User-Alice"}) {
		t.Fatalf("expected base->relocated to stay a relocation fact, got %v", got)
	}

	modifiedEntry := modifiedResult.Entries[0]
	if len(modifiedEntry.FieldChanges) != 1 || modifiedEntry.FieldChanges[0].Path != "name" {
		t.Fatalf("expected only semantic field change in modified result, got %+v", modifiedEntry.FieldChanges)
	}

	relocatedEntry := relocatedResult.Entries[0]
	if relocatedEntry.Kind != DiffEntryKindRelocated {
		t.Fatalf("expected relocation entry, got %+v", relocatedEntry)
	}
	if !semanticSourcesEqual(relocatedEntry.OldSources, []LogicalObjectSource{{Path: "data/users/user-alice.yaml"}}) {
		t.Fatalf("expected original relocation metadata, got %+v", relocatedEntry.OldSources)
	}
	if !semanticSourcesEqual(relocatedEntry.NewSources, []LogicalObjectSource{{Path: "archive/users/user-alice.yaml"}}) {
		t.Fatalf("expected new relocation metadata, got %+v", relocatedEntry.NewSources)
	}
}

func mustDiffLogicalDatabases(t *testing.T, left, right LogicalDatabase) DiffResult {
	t.Helper()
	result, err := diffLogicalDatabases(left, right)
	if err != nil {
		t.Fatalf("diff logical databases: %v", err)
	}
	return result
}

func logicalDatabaseFixture(objects ...LogicalObject) LogicalDatabase {
	db := LogicalDatabase{
		Objects: append([]LogicalObject(nil), objects...),
	}
	return db
}

func logicalObjectFixture(typeName, id string, fields map[string]any, path string) LogicalObject {
	canonical, err := canonicalizeLogicalFields(fields)
	if err != nil {
		panic(err)
	}

	return LogicalObject{
		Type:      typeName,
		ID:        id,
		Fields:    cloneMap(fields),
		Canonical: canonical,
		Sources: []LogicalObjectSource{{
			Path: path,
		}},
	}
}

func diffEntryKeys(result DiffResult) []string {
	keys := make([]string, 0, len(result.Entries))
	for _, entry := range result.Entries {
		keys = append(keys, string(entry.Kind)+":"+entry.Type+":"+entry.ObjectID)
	}
	return keys
}

func assertDiffResultEqual(t *testing.T, got, want DiffResult) {
	t.Helper()
	if len(got.Entries) == 0 {
		got.Entries = nil
	}
	if len(want.Entries) == 0 {
		want.Entries = nil
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected diff result\nwant: %#v\ngot:  %#v", want, got)
	}
}
