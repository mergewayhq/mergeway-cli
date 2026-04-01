package diff

import "testing"

func TestMarshalDiffResultJSONIsDeterministic(t *testing.T) {
	result := DiffResult{
		Entries: []DiffEntry{
			{
				Kind:     DiffEntryKindAdded,
				Type:     "Tag",
				ObjectID: "Tag-New",
				NewValue: map[string]any{
					"id":    "Tag-New",
					"label": "New",
				},
				NewSources: []LogicalObjectSource{
					{Path: "data/tags/tag-new.yaml"},
				},
			},
			{
				Kind:     DiffEntryKindModified,
				Type:     "User",
				ObjectID: "User-Alice",
				OldValue: map[string]any{
					"id":    "User-Alice",
					"email": "alice@example.com",
					"name":  "Alice",
				},
				NewValue: map[string]any{
					"id":    "User-Alice",
					"email": "alice+updated@example.com",
					"name":  "Alice Updated",
				},
				FieldChanges: []DiffFieldChange{
					{
						Path:     "name",
						OldValue: "Alice",
						NewValue: "Alice Updated",
					},
					{
						Path:     "email",
						OldValue: "alice@example.com",
						NewValue: "alice+updated@example.com",
					},
				},
				OldSources: []LogicalObjectSource{{Path: "data/users/user-alice.yaml"}},
				NewSources: []LogicalObjectSource{{Path: "data/users/user-alice.yaml"}},
			},
		},
	}

	payload, err := marshalDiffResultJSON(result)
	if err != nil {
		t.Fatalf("marshal diff result json: %v", err)
	}

	want := `{
  "version": 1,
  "entries": [
    {
      "kind": "added",
      "type": "Tag",
      "object_id": "Tag-New",
      "value": {
        "id": "Tag-New",
        "label": "New"
      },
      "sources": [
        {
          "path": "data/tags/tag-new.yaml"
        }
      ]
    },
    {
      "kind": "modified",
      "type": "User",
      "object_id": "User-Alice",
      "old_value": {
        "email": "alice@example.com",
        "id": "User-Alice",
        "name": "Alice"
      },
      "new_value": {
        "email": "alice+updated@example.com",
        "id": "User-Alice",
        "name": "Alice Updated"
      },
      "changes": [
        {
          "path": "email",
          "before": "alice@example.com",
          "after": "alice+updated@example.com"
        },
        {
          "path": "name",
          "before": "Alice",
          "after": "Alice Updated"
        }
      ],
      "old_sources": [
        {
          "path": "data/users/user-alice.yaml"
        }
      ],
      "new_sources": [
        {
          "path": "data/users/user-alice.yaml"
        }
      ]
    }
  ]
}`
	if string(payload) != want {
		t.Fatalf("unexpected json serialization\nwant:\n%s\ngot:\n%s", want, string(payload))
	}
}

func TestMarshalDiffResultJSONAddRemoveEntriesSerializeCorrectly(t *testing.T) {
	result := DiffResult{
		Entries: []DiffEntry{
			{
				Kind:     DiffEntryKindAdded,
				Type:     "Service",
				ObjectID: "svc-new",
				NewValue: map[string]any{
					"id": "svc-new",
				},
				NewSources: []LogicalObjectSource{{Path: "data/services/new.yaml"}},
			},
			{
				Kind:     DiffEntryKindRemoved,
				Type:     "Service",
				ObjectID: "svc-old",
				OldValue: map[string]any{
					"id": "svc-old",
				},
				OldSources: []LogicalObjectSource{{Path: "data/services/old.yaml"}},
			},
		},
	}

	payload, err := marshalDiffResultJSON(result)
	if err != nil {
		t.Fatalf("marshal diff result json: %v", err)
	}

	want := `{
  "version": 1,
  "entries": [
    {
      "kind": "added",
      "type": "Service",
      "object_id": "svc-new",
      "value": {
        "id": "svc-new"
      },
      "sources": [
        {
          "path": "data/services/new.yaml"
        }
      ]
    },
    {
      "kind": "removed",
      "type": "Service",
      "object_id": "svc-old",
      "value": {
        "id": "svc-old"
      },
      "sources": [
        {
          "path": "data/services/old.yaml"
        }
      ]
    }
  ]
}`
	if string(payload) != want {
		t.Fatalf("unexpected json serialization\nwant:\n%s\ngot:\n%s", want, string(payload))
	}
}

func TestMarshalDiffResultJSONEmptyDiffSerializesCorrectly(t *testing.T) {
	payload, err := marshalDiffResultJSON(DiffResult{})
	if err != nil {
		t.Fatalf("marshal diff result json: %v", err)
	}

	want := `{
  "version": 1,
  "entries": []
}`
	if string(payload) != want {
		t.Fatalf("unexpected json serialization\nwant:\n%s\ngot:\n%s", want, string(payload))
	}
}
