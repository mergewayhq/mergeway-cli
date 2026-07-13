package mcp

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestNewServiceRejectsUnknownAllowedEntity(t *testing.T) {
	root := filepath.Join("..", "data", "testdata", "repo")

	_, err := NewService(root, []string{"Missing"})
	if err == nil || !errors.Is(err, ErrUnknownEntity) {
		t.Fatalf("expected unknown entity error, got %v", err)
	}
}

func TestServiceEntityListFiltersAllowedEntitiesExactly(t *testing.T) {
	root := inheritanceRepo(t)

	service, err := NewService(root, []string{"Animal"})
	if err != nil {
		t.Fatalf("NewService: %v", err)
	}

	names, err := service.EntityList()
	if err != nil {
		t.Fatalf("EntityList: %v", err)
	}
	if !reflect.DeepEqual(names, []string{"Animal"}) {
		t.Fatalf("expected only Animal to be visible, got %v", names)
	}
}

func TestServiceObjectListUsesExactEntitySemantics(t *testing.T) {
	root := inheritanceRepo(t)

	service, err := NewService(root, nil)
	if err != nil {
		t.Fatalf("NewService: %v", err)
	}

	animals, err := service.ObjectList("Animal")
	if err != nil {
		t.Fatalf("ObjectList: %v", err)
	}
	if len(animals) != 1 {
		t.Fatalf("expected 1 exact Animal object, got %d", len(animals))
	}
	if animals[0].Type != "Animal" || animals[0].ID != "animal-1" {
		t.Fatalf("expected exact Animal object, got %+v", animals[0])
	}

	dogs, err := service.ObjectList("Dog")
	if err != nil {
		t.Fatalf("ObjectList Dog: %v", err)
	}
	if len(dogs) != 1 || dogs[0].Type != "Dog" || dogs[0].ID != "dog-1" {
		t.Fatalf("expected exact Dog object, got %+v", dogs)
	}
}

func TestServiceObjectGetRejectsDescendantLookupThroughParentEntity(t *testing.T) {
	root := inheritanceRepo(t)

	service, err := NewService(root, nil)
	if err != nil {
		t.Fatalf("NewService: %v", err)
	}

	_, err = service.ObjectGet("Animal", "dog-1")
	if err == nil || !strings.Contains(err.Error(), `Animal "dog-1" not found`) {
		t.Fatalf("expected exact entity miss, got %v", err)
	}
}

func TestServiceObjectGetRejectsBlockedEntityEvenWhenConfigContainsIt(t *testing.T) {
	root := inheritanceRepo(t)

	service, err := NewService(root, []string{"Animal"})
	if err != nil {
		t.Fatalf("NewService: %v", err)
	}

	_, err = service.ObjectGet("Dog", "dog-1")
	if err == nil || !errors.Is(err, ErrEntityNotAllowed) {
		t.Fatalf("expected not-allowed error, got %v", err)
	}
}

func TestServiceObjectGetSupportsPathIdentifiers(t *testing.T) {
	root := filepath.Join("..", "validation", "testdata", "path_identifier")

	service, err := NewService(root, nil)
	if err != nil {
		t.Fatalf("NewService: %v", err)
	}

	obj, err := service.ObjectGet("Note", "data/notes/alpha.yaml")
	if err != nil {
		t.Fatalf("ObjectGet: %v", err)
	}
	if obj.Type != "Note" || obj.ID != "data/notes/alpha.yaml" {
		t.Fatalf("expected Note path identifier object, got %+v", obj)
	}
	if obj.Fields["title"] != "Alpha" {
		t.Fatalf("expected title Alpha, got %v", obj.Fields["title"])
	}
}

func TestServiceObjectListSupportsExternalRootPathEntities(t *testing.T) {
	root := filepath.Join("..", "..", "examples", "external-root-path", "primary")

	service, err := NewService(root, []string{"Product"})
	if err != nil {
		t.Fatalf("NewService: %v", err)
	}

	objects, err := service.ObjectList("Product")
	if err != nil {
		t.Fatalf("ObjectList: %v", err)
	}
	if len(objects) != 2 {
		t.Fatalf("expected 2 products, got %d", len(objects))
	}
	ids := []string{objects[0].ID, objects[1].ID}
	if !reflect.DeepEqual(ids, []string{"../secondary/products/gadget.yaml", "../secondary/products/widget.yaml"}) {
		t.Fatalf("expected external path ids, got %v", ids)
	}
}

func TestServiceRepositoryExportFiltersAllowedEntities(t *testing.T) {
	root := inheritanceRepo(t)

	service, err := NewService(root, []string{"Dog"})
	if err != nil {
		t.Fatalf("NewService: %v", err)
	}

	exported, err := service.RepositoryExport(nil)
	if err != nil {
		t.Fatalf("RepositoryExport: %v", err)
	}
	if len(exported) != 1 {
		t.Fatalf("expected one exported entity, got %d", len(exported))
	}
	if _, ok := exported["Dog"]; !ok {
		t.Fatalf("expected Dog export, got %+v", exported)
	}
	if _, ok := exported["Animal"]; ok {
		t.Fatalf("did not expect Animal export, got %+v", exported)
	}
}

func TestServiceFilesListReturnsWorkspaceRelativePaths(t *testing.T) {
	root := filepath.Join("..", "..", "examples", "external-root-path", "primary")

	service, err := NewService(root, []string{"Product"})
	if err != nil {
		t.Fatalf("NewService: %v", err)
	}

	files, err := service.FilesList("Product")
	if err != nil {
		t.Fatalf("FilesList: %v", err)
	}

	expected := []FileEntry{
		{Type: "Product", File: "../secondary/products/gadget.yaml"},
		{Type: "Product", File: "../secondary/products/widget.yaml"},
	}
	if !reflect.DeepEqual(files, expected) {
		t.Fatalf("expected files %v, got %v", expected, files)
	}
}

func TestServiceReloadsRepositoryStatePerRequest(t *testing.T) {
	root := copyFixture(t, filepath.Join("..", "data", "testdata", "repo"))

	service, err := NewService(root, nil)
	if err != nil {
		t.Fatalf("NewService: %v", err)
	}

	before, err := service.ObjectList("User")
	if err != nil {
		t.Fatalf("ObjectList before: %v", err)
	}
	if len(before) != 2 {
		t.Fatalf("expected 2 users before change, got %d", len(before))
	}

	target := filepath.Join(root, "data", "users", "user-zed.yaml")
	body := "id: User-Zed\nname: Zed Example\nemail: zed@example.com\n"
	if err := os.WriteFile(target, []byte(body), 0o644); err != nil {
		t.Fatalf("write new user: %v", err)
	}

	after, err := service.ObjectList("User")
	if err != nil {
		t.Fatalf("ObjectList after: %v", err)
	}
	if len(after) != 3 {
		t.Fatalf("expected 3 users after change, got %d", len(after))
	}
	if after[2].ID != "User-Zed" {
		t.Fatalf("expected newly loaded user, got %+v", after)
	}
}

func inheritanceRepo(t *testing.T) string {
	t.Helper()

	root := t.TempDir()
	cfg := `mergeway:
  version: 1

entities:
  Animal:
    identifier: id
    include:
      - data/animals/*.yaml
    fields:
      id: string
      name: string
  Dog:
    extends: Animal
    include:
      - data/dogs/*.yaml
    fields:
      breed: string
`
	if err := os.WriteFile(filepath.Join(root, "mergeway.yaml"), []byte(cfg), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(root, "data", "animals"), 0o755); err != nil {
		t.Fatalf("mkdir animals: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(root, "data", "dogs"), 0o755); err != nil {
		t.Fatalf("mkdir dogs: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "data", "animals", "animal.yaml"), []byte("id: animal-1\nname: Generic\n"), 0o644); err != nil {
		t.Fatalf("write animal: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "data", "dogs", "dog.yaml"), []byte("id: dog-1\nname: Fido\nbreed: collie\n"), 0o644); err != nil {
		t.Fatalf("write dog: %v", err)
	}

	return root
}

func copyFixture(t *testing.T, src string) string {
	t.Helper()

	dest := t.TempDir()
	if err := filepath.WalkDir(src, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dest, rel)
		if d.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(target, data, 0o644)
	}); err != nil {
		t.Fatalf("copy fixture: %v", err)
	}

	return dest
}
