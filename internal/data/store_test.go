package data

import (
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"

	"github.com/mergewayhq/mergeway-cli/internal/config"
)

func TestStoreListAndGet(t *testing.T) {
	store, repo := setupStore(t, "repo")

	ids, err := store.List("User")
	if err != nil {
		t.Fatalf("List returned error: %v", err)
	}

	expectedIDs := []string{"User-Alice", "User-Bob"}
	if !reflect.DeepEqual(ids, expectedIDs) {
		t.Fatalf("expected IDs %v, got %v", expectedIDs, ids)
	}

	obj, err := store.Get("Post", "Post-001")
	if err != nil {
		t.Fatalf("Get returned error: %v", err)
	}

	if obj.Type != "Post" {
		t.Fatalf("expected type Post, got %s", obj.Type)
	}

	if obj.Fields["title"] != "First Post" {
		t.Fatalf("expected title 'First Post', got %v", obj.Fields["title"])
	}

	tags, ok := obj.Fields["tags"].([]any)
	if !ok || len(tags) != 2 {
		t.Fatalf("expected tags array with length 2, got %v", obj.Fields["tags"])
	}

	// Ensure we can obtain configuration again after operations (sanity check for repo path usage).
	if _, err := config.Load(filepath.Join(repo, "mergeway.yaml")); err != nil {
		t.Fatalf("Reload config failed: %v", err)
	}
}

func TestStoreCreateUpdateDelete(t *testing.T) {
	store, repo := setupStore(t, "repo")

	// Create a new tag (single-object file path)
	created, err := store.Create("Tag", map[string]any{
		"id":    "Tag-New",
		"label": "New",
	})
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}

	expectedPath := filepath.Join(repo, "data", "tags", "Tag-New.yaml")
	if created.File != expectedPath {
		t.Fatalf("expected file %s, got %s", expectedPath, created.File)
	}

	if _, err := os.Stat(expectedPath); err != nil {
		t.Fatalf("expected file to exist: %v", err)
	}

	// Merge update on existing user (single-object file)
	updated, err := store.Update("User", "User-Bob", map[string]any{
		"role": "author",
	}, true)
	if err != nil {
		t.Fatalf("Update merge returned error: %v", err)
	}

	if updated.Fields["role"] != "author" {
		t.Fatalf("expected merged role 'author', got %v", updated.Fields["role"])
	}
	if updated.Fields["name"] != "Bob Example" {
		t.Fatalf("expected name to remain 'Bob Example', got %v", updated.Fields["name"])
	}

	// Delete from multi-object file
	if err := store.Delete("Post", "Post-002"); err != nil {
		t.Fatalf("Delete returned error: %v", err)
	}

	postsPath := filepath.Join(repo, "data", "posts", "posts.yaml")
	content, err := os.ReadFile(postsPath)
	if err != nil {
		t.Fatalf("reading posts file: %v", err)
	}

	if strings.Count(string(content), "Post-002") != 0 {
		t.Fatalf("expected Post-002 to be removed from file")
	}

	// Ensure Post-001 still retrievable
	if _, err := store.Get("Post", "Post-001"); err != nil {
		t.Fatalf("expected Post-001 to remain accessible: %v", err)
	}
}

func TestStoreJSONPathIncludes(t *testing.T) {
	store, repo := setupStore(t, "jsonpath")

	ids, err := store.List("User")
	if err != nil {
		t.Fatalf("List returned error: %v", err)
	}

	expectedIDs := []string{"User-A", "User-B"}
	if !reflect.DeepEqual(ids, expectedIDs) {
		t.Fatalf("expected IDs %v, got %v", expectedIDs, ids)
	}

	obj, err := store.Get("User", "User-B")
	if err != nil {
		t.Fatalf("Get returned error: %v", err)
	}

	if obj.Fields["name"] != "Bob" {
		t.Fatalf("expected name 'Bob', got %v", obj.Fields["name"])
	}

	jsonPathFile := filepath.Join(repo, "data", "users.json")
	if obj.File != jsonPathFile {
		t.Fatalf("expected file %s, got %s", jsonPathFile, obj.File)
	}

	if _, err := store.Update("User", "User-A", map[string]any{"name": "Updated"}, true); err == nil || !strings.Contains(err.Error(), "selector include") {
		t.Fatalf("expected update to fail for selector include, got %v", err)
	}

	if _, err := store.Create("User", map[string]any{"id": "User-C", "name": "Carol"}); err == nil || !strings.Contains(err.Error(), "selector") {
		t.Fatalf("expected create to fail for selector include, got %v", err)
	}

	if err := store.Delete("User", "User-A"); err == nil || !strings.Contains(err.Error(), "selector include") {
		t.Fatalf("expected delete to fail for selector include, got %v", err)
	}
}

func TestStoreWithInlineConfig(t *testing.T) {
	store, _ := setupStore(t, "inline")

	// Ensure list works with inline config
	ids, err := store.List("Post")
	if err != nil {
		t.Fatalf("List returned error: %v", err)
	}

	expected := []string{"Post-001"}
	if !reflect.DeepEqual(ids, expected) {
		t.Fatalf("expected IDs %v, got %v", expected, ids)
	}

	// Update single-object file
	_, err = store.Update("Post", "Post-001", map[string]any{
		"title": "Inline Updated",
	}, true)
	if err != nil {
		t.Fatalf("Update returned error: %v", err)
	}

	obj, err := store.Get("Post", "Post-001")
	if err != nil {
		t.Fatalf("Get returned error after update: %v", err)
	}

	if obj.Fields["title"] != "Inline Updated" {
		t.Fatalf("expected updated title, got %v", obj.Fields["title"])
	}

	tagIDs, err := store.List("Tag")
	if err != nil {
		t.Fatalf("List tags returned error: %v", err)
	}

	expectedTagIDs := []string{"Tag-Docs", "Tag-Inline"}
	if !reflect.DeepEqual(tagIDs, expectedTagIDs) {
		t.Fatalf("expected tag IDs %v, got %v", expectedTagIDs, tagIDs)
	}

	inlineTag, err := store.Get("Tag", "Tag-Inline")
	if err != nil {
		t.Fatalf("Get inline tag returned error: %v", err)
	}

	if inlineTag.Fields["label"] != "Inline Tag" {
		t.Fatalf("expected inline tag label 'Inline Tag', got %v", inlineTag.Fields["label"])
	}

	if _, err := store.Update("Tag", "Tag-Inline", map[string]any{"label": "New"}, false); err == nil || !strings.Contains(err.Error(), "inline") {
		t.Fatalf("expected inline update to fail, got %v", err)
	}

	if err := store.Delete("Tag", "Tag-Inline"); err == nil || !strings.Contains(err.Error(), "inline") {
		t.Fatalf("expected inline delete to fail, got %v", err)
	}
}

func TestStoreNumericIdentifiers(t *testing.T) {
	store, repo := setupStore(t, "numeric")

	ids, err := store.List("Person")
	if err != nil {
		t.Fatalf("List returned error: %v", err)
	}

	expected := []string{"1"}
	if !reflect.DeepEqual(ids, expected) {
		t.Fatalf("expected IDs %v, got %v", expected, ids)
	}

	obj, err := store.Get("Person", "1")
	if err != nil {
		t.Fatalf("Get returned error: %v", err)
	}

	if _, ok := obj.Fields["id"].(int); !ok {
		t.Fatalf("expected stored id to be an integer, got %T", obj.Fields["id"])
	}

	created, err := store.Create("Person", map[string]any{
		"id":   "2",
		"name": "Babbage",
	})
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}

	if created.ID != "2" {
		t.Fatalf("expected created ID '2', got %q", created.ID)
	}

	if _, ok := created.Fields["id"].(int64); !ok {
		t.Fatalf("expected stored id to coerce to int64, got %T", created.Fields["id"])
	}

	path := filepath.Join(repo, "data", "user-2.yaml")
	contents, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading generated file: %v", err)
	}

	if string(contents) == "" || !strings.Contains(string(contents), "id: 2") {
		t.Fatalf("expected file contents to include 'id: 2', got %q", string(contents))
	}

	if _, err := store.Get("Person", "2"); err != nil {
		t.Fatalf("expected to retrieve newly created Person: %v", err)
	}
}

func setupStore(t *testing.T, fixture string) (*Store, string) {
	t.Helper()
	repo := copyRepo(t, fixture)

	cfg, err := config.Load(filepath.Join(repo, "mergeway.yaml"))
	if err != nil {
		t.Fatalf("load config: %v", err)
	}

	store, err := NewStore(repo, cfg)
	if err != nil {
		t.Fatalf("new store: %v", err)
	}

	return store, repo
}

func copyRepo(t *testing.T, fixture string) string {
	t.Helper()
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	base := filepath.Dir(filename)
	src := filepath.Join(base, "testdata", fixture)
	dest := t.TempDir()

	if err := filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}

		target := filepath.Join(dest, rel)

		if info.IsDir() {
			return os.MkdirAll(target, info.Mode())
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		return os.WriteFile(target, data, info.Mode())
	}); err != nil {
		t.Fatalf("copy repo: %v", err)
	}

	return dest
}
