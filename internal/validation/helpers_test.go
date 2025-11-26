package validation

import (
	"strings"
	"testing"

	"github.com/mergewayhq/mergeway-cli/internal/config"
)

func TestNormalizedUniqueKeyCoversTypes(t *testing.T) {
	cases := []struct {
		name  string
		input any
	}{
		{name: "string", input: "value"},
		{name: "bool", input: true},
		{name: "int", input: 7},
		{name: "float", input: 3.14},
		{name: "stringer", input: testStringer("ref")},
		{name: "slice", input: []any{"a", "b"}},
		{name: "map", input: map[string]any{"a": 1}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if key := normalizedUniqueKey(tc.input); key == "" {
				t.Fatalf("expected non-empty key for %s", tc.name)
			}
		})
	}
}

type testStringer string

func (t testStringer) String() string {
	return string(t)
}

func TestCloneValueDeepCopies(t *testing.T) {
	src := map[string]any{
		"name": "example",
		"tags": []any{"a", "b"},
	}
	clone := cloneMap(src)
	clone["name"] = "updated"
	clone["tags"].([]any)[0] = "z"

	if src["name"] != "example" {
		t.Fatalf("expected original map untouched")
	}
	if src["tags"].([]any)[0] != "a" {
		t.Fatalf("expected original slice untouched")
	}
}

func TestTypeErrorIncludesLocation(t *testing.T) {
	obj := &rawObject{
		typeDef: &config.TypeDefinition{Name: "User"},
		id:      "User-1",
		file:    "data/users/user.yaml",
	}
	err := typeError(obj, "name", "string")
	if err == nil {
		t.Fatalf("expected error")
	}
	if !strings.Contains(err.Message, "field \"name\" must be string") {
		t.Fatalf("unexpected message: %s", err.Message)
	}
	if err.Type != "User" || err.ID != "User-1" {
		t.Fatalf("expected metadata to be preserved, got %+v", err)
	}
	if !strings.HasPrefix(err.File, "data/users/user.yaml") {
		t.Fatalf("expected file context, got %s", err.File)
	}
}
