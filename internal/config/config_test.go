package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadValidConfig(t *testing.T) {
	path := filepath.Join("testdata", "valid", "mergeway.yaml")

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	if cfg.Version != 1 {
		t.Fatalf("expected version 1, got %d", cfg.Version)
	}

	if len(cfg.Types) != 3 {
		t.Fatalf("expected 3 types, got %d", len(cfg.Types))
	}

	post, ok := cfg.Types["Post"]
	if !ok {
		t.Fatalf("expected type 'Post' to be present")
	}

	if post.Description != "Primary blog post content type" {
		t.Fatalf("expected Post description to be present")
	}

	if post.Identifier.Field != "id" {
		t.Fatalf("expected Post identifier field 'id', got %q", post.Identifier.Field)
	}

	title, ok := post.Fields["title"]
	if !ok {
		t.Fatalf("expected Post field 'title' to be present")
	}

	if title.Description != "Human readable title shown in listings" {
		t.Fatalf("expected title description, got %q", title.Description)
	}

	tags, ok := post.Fields["tags"]
	if !ok {
		t.Fatalf("expected Post field 'tags' to be present")
	}

	if !tags.Repeated {
		t.Fatalf("expected tags to be marked repeated")
	}

	if tags.Type != "Tag" {
		t.Fatalf("expected tags to reference Tag, got %q", tags.Type)
	}

	author, ok := post.Fields["author"]
	if !ok {
		t.Fatalf("expected Post field 'author' to be present")
	}

	if author.Type != "User" {
		t.Fatalf("expected author to reference User, got %q", author.Type)
	}

	tag, ok := cfg.Types["Tag"]
	if !ok {
		t.Fatalf("expected type 'Tag' to be present")
	}

	if len(tag.InlineData) != 1 {
		t.Fatalf("expected Tag to have 1 inline data item, got %d", len(tag.InlineData))
	}

	if value := tag.InlineData[0]["label"]; value != "Inline Tag" {
		t.Fatalf("expected inline tag label 'Inline Tag', got %v", value)
	}
}

func TestLoadRepeatedUniqueError(t *testing.T) {
	path := filepath.Join("testdata", "repeated_unique", "mergeway.yaml")
	_, err := Load(path)
	if err == nil {
		t.Fatalf("expected error for repeated unique field")
	}

	if got := err.Error(); got == "" || !strings.Contains(got, "cannot declare unique") {
		t.Fatalf("expected unique/repeated error, got %q", got)
	}
}

func TestLoadInvalidIdentifier(t *testing.T) {
	path := filepath.Join("testdata", "invalid_identifier", "mergeway.yaml")
	_, err := Load(path)
	if err == nil {
		t.Fatalf("expected error for invalid identifier")
	}

	if got := err.Error(); got == "" || !strings.Contains(got, "invalid type identifier") {
		t.Fatalf("expected invalid identifier error, got %q", got)
	}
}

func TestLoadUnknownReference(t *testing.T) {
	path := filepath.Join("testdata", "unknown_reference", "mergeway.yaml")
	_, err := Load(path)
	if err == nil {
		t.Fatalf("expected error for unknown reference type")
	}

	if got := err.Error(); got == "" || !strings.Contains(got, "references unknown type") {
		t.Fatalf("expected unknown type error, got %q", got)
	}
}

func TestLoadShorthandFieldDefinitions(t *testing.T) {
	path := filepath.Join("testdata", "shorthand", "mergeway.yaml")
	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	person, ok := cfg.Types["Person"]
	if !ok {
		t.Fatalf("expected type 'Person' to be present")
	}

	idField, ok := person.Fields["id"]
	if !ok {
		t.Fatalf("expected field 'id'")
	}
	if idField.Type != "string" {
		t.Fatalf("expected id field type 'string', got %q", idField.Type)
	}
	if idField.Required {
		t.Fatalf("expected shorthand id field to default to optional")
	}

	ageField, ok := person.Fields["age"]
	if !ok {
		t.Fatalf("expected field 'age'")
	}
	if ageField.Type != "integer" {
		t.Fatalf("expected age field type 'integer', got %q", ageField.Type)
	}
	if ageField.Required {
		t.Fatalf("expected shorthand age field to default to optional")
	}

	nameField, ok := person.Fields["name"]
	if !ok {
		t.Fatalf("expected field 'name'")
	}
	if !nameField.Required {
		t.Fatalf("expected explicit mapping to preserve required=true")
	}
}

func TestLoadIncludeWithJSONPath(t *testing.T) {
	path := filepath.Join("testdata", "jsonpath", "mergeway.yaml")

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	item, ok := cfg.Types["Item"]
	if !ok {
		t.Fatalf("expected type 'Item' to be present")
	}

	if len(item.Include) != 1 {
		t.Fatalf("expected one include, got %d", len(item.Include))
	}

	include := item.Include[0]
	if include.Path != "data/items.json" {
		t.Fatalf("expected include path 'data/items.json', got %q", include.Path)
	}
	if include.Selector != "$.items[*]" {
		t.Fatalf("expected selector '$.items[*]', got %q", include.Selector)
	}
}

func TestLoadMissingMergewayBlock(t *testing.T) {
	path := filepath.Join(t.TempDir(), "mergeway.yaml")
	content := []byte(`entities:
  Foo:
    identifier: id
    fields:
      id: string
`)
	if err := os.WriteFile(path, content, 0o644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	_, err := Load(path)
	if err == nil {
		t.Fatalf("expected error for missing mergeway block")
	}

	if got := err.Error(); !strings.Contains(got, "mergeway block is required") {
		t.Fatalf("expected mergeway block error, got %q", got)
	}
}

func TestLoadUnsupportedConfigVersion(t *testing.T) {
	path := filepath.Join(t.TempDir(), "mergeway.yaml")
	content := []byte(`mergeway:
  version: 2

entities:
  Foo:
    identifier: id
    fields:
      id: string
`)
	if err := os.WriteFile(path, content, 0o644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	_, err := Load(path)
	if err == nil {
		t.Fatalf("expected error for unsupported version")
	}

	if got := err.Error(); !strings.Contains(got, "mergeway.version must be") {
		t.Fatalf("expected unsupported version error, got %q", got)
	}
}
