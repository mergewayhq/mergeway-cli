package data

import (
	"path/filepath"
	"testing"

	"github.com/mergewayhq/mergeway-cli/internal/config"
)

func TestStoreLoadExamples(t *testing.T) {
	root := rootExamplesDir(t)
	cfgPath := filepath.Join(root, "mergeway.yaml")
	cfg, err := config.Load(cfgPath)
	if err != nil {
		t.Fatalf("load config: %v", err)
	}

	store, err := NewStore(root, cfg)
	if err != nil {
		t.Fatalf("new store: %v", err)
	}

	posts, err := store.LoadAll("Post")
	if err != nil {
		t.Fatalf("load posts: %v", err)
	}
	if len(posts) != 1 {
		t.Fatalf("expected 1 post, got %d", len(posts))
	}

	comments, err := store.LoadAll("Comment")
	if err != nil {
		t.Fatalf("load comments: %v", err)
	}
	if len(comments) != 1 {
		t.Fatalf("expected 1 comment, got %d", len(comments))
	}
}

func rootExamplesDir(t *testing.T) string {
	path := filepath.Join("..", "..", "examples", "full")
	abs, err := filepath.Abs(path)
	if err != nil {
		t.Fatalf("abs path: %v", err)
	}
	return abs
}
