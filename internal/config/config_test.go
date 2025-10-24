package config

import (
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

	if post.Identifier.Field != "id" {
		t.Fatalf("expected Post identifier field 'id', got %q", post.Identifier.Field)
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
