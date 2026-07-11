package lsp

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"go.lsp.dev/protocol"
	"go.lsp.dev/uri"
)

func TestHandleReferencesUsesSemanticTargets(t *testing.T) {
	server, root := initializeExampleFullServer(t)
	userPath := filepath.Join(root, "data", "users", "alice.yaml")
	userContent := readFile(t, userPath)
	postPath := filepath.Join(root, "data", "posts", "launch.yaml")
	postContent := readFile(t, postPath)

	referencePosition := positionInContent(t, userContent, "user-alice")
	declarationRange := scalarFieldValueRange(t, userContent, "id")
	usageRange := scalarFieldValueRange(t, postContent, "author")

	t.Run("exclude declaration", func(t *testing.T) {
		var result []protocol.Location
		callServer(t, server, protocol.MethodTextDocumentReferences, 2, &protocol.ReferenceParams{
			TextDocumentPositionParams: protocol.TextDocumentPositionParams{
				TextDocument: protocol.TextDocumentIdentifier{URI: protocol.DocumentURI(uri.File(userPath))},
				Position:     referencePosition,
			},
			Context: protocol.ReferenceContext{IncludeDeclaration: false},
		}, &result)

		if len(result) != 1 {
			t.Fatalf("expected one usage, got %d", len(result))
		}
		expectLocation(t, result[0], postPath, usageRange)
	})

	t.Run("include declaration", func(t *testing.T) {
		var result []protocol.Location
		callServer(t, server, protocol.MethodTextDocumentReferences, 3, &protocol.ReferenceParams{
			TextDocumentPositionParams: protocol.TextDocumentPositionParams{
				TextDocument: protocol.TextDocumentIdentifier{URI: protocol.DocumentURI(uri.File(userPath))},
				Position:     referencePosition,
			},
			Context: protocol.ReferenceContext{IncludeDeclaration: true},
		}, &result)

		if len(result) != 2 {
			t.Fatalf("expected declaration and usage, got %d", len(result))
		}
		expectLocations(t, result, []expectedLocation{
			{path: userPath, rng: declarationRange},
			{path: postPath, rng: usageRange},
		})
	})
}

func TestHandleReferencesIncludeParentTypedUsagesForDescendants(t *testing.T) {
	server, root := initializeInheritanceServer(t)
	dogPath := filepath.Join(root, "data", "dogs", "dog.yaml")
	dogContent := readFile(t, dogPath)
	kennelPath := filepath.Join(root, "data", "kennels", "kennel.yaml")
	kennelContent := readFile(t, kennelPath)

	referencePosition := positionInContent(t, dogContent, "dog-1")
	declarationRange := scalarFieldValueRange(t, dogContent, "id")
	usageRange := scalarFieldValueRange(t, kennelContent, "resident")

	var result []protocol.Location
	callServer(t, server, protocol.MethodTextDocumentReferences, 5, &protocol.ReferenceParams{
		TextDocumentPositionParams: protocol.TextDocumentPositionParams{
			TextDocument: protocol.TextDocumentIdentifier{URI: protocol.DocumentURI(uri.File(dogPath))},
			Position:     referencePosition,
		},
		Context: protocol.ReferenceContext{IncludeDeclaration: true},
	}, &result)

	if len(result) != 2 {
		t.Fatalf("expected declaration and parent-typed usage, got %d", len(result))
	}
	expectLocations(t, result, []expectedLocation{
		{path: dogPath, rng: declarationRange},
		{path: kennelPath, rng: usageRange},
	})
}

func TestHandleDocumentSymbolMapsEntitiesAndFields(t *testing.T) {
	t.Run("single object document", func(t *testing.T) {
		server, root := initializeExampleFullServer(t)
		postPath := filepath.Join(root, "data", "posts", "launch.yaml")

		var result []protocol.DocumentSymbol
		callServer(t, server, protocol.MethodTextDocumentDocumentSymbol, 2, &protocol.DocumentSymbolParams{
			TextDocument: protocol.TextDocumentIdentifier{URI: protocol.DocumentURI(uri.File(postPath))},
		}, &result)

		if len(result) != 1 {
			t.Fatalf("expected one top-level symbol, got %d", len(result))
		}
		if result[0].Name != "post-001" {
			t.Fatalf("expected top-level symbol post-001, got %q", result[0].Name)
		}
		if !strings.Contains(result[0].Detail, "Post") {
			t.Fatalf("expected Post detail, got %q", result[0].Detail)
		}
		if !containsDocumentSymbol(result[0].Children, "title") || !containsDocumentSymbol(result[0].Children, "author") {
			t.Fatalf("expected title and author child symbols, got %+v", documentSymbolNames(result[0].Children))
		}
	})

	t.Run("multi object document", func(t *testing.T) {
		server, root := initializeExampleFullServer(t)
		tagPath := filepath.Join(root, "data", "tags", "product.yaml")
		tagContent := readFile(t, tagPath)

		var result []protocol.DocumentSymbol
		callServer(t, server, protocol.MethodTextDocumentDocumentSymbol, 3, &protocol.DocumentSymbolParams{
			TextDocument: protocol.TextDocumentIdentifier{URI: protocol.DocumentURI(uri.File(tagPath))},
		}, &result)

		if len(result) != 2 {
			t.Fatalf("expected two top-level tag symbols, got %d", len(result))
		}
		if result[0].Name != "tag-product" || result[1].Name != "tag-update" {
			t.Fatalf("unexpected symbol names: %v", documentSymbolNames(result))
		}
		wantSelection := sequenceItemFieldValueRange(t, tagContent, 0, "id")
		if result[0].SelectionRange != wantSelection {
			t.Fatalf("expected first selection range %+v, got %+v", wantSelection, result[0].SelectionRange)
		}
		if !containsDocumentSymbol(result[0].Children, "label") {
			t.Fatalf("expected label child symbol, got %+v", documentSymbolNames(result[0].Children))
		}
	})
}

func TestHandleWorkspaceSymbolFindsByNameAndKeepsRootsSeparate(t *testing.T) {
	t.Run("find by name", func(t *testing.T) {
		server, root := initializeExampleFullServer(t)
		userPath := filepath.Join(root, "data", "users", "alice.yaml")

		var result []protocol.SymbolInformation
		callServer(t, server, protocol.MethodWorkspaceSymbol, 2, &protocol.WorkspaceSymbolParams{
			Query: "Alice Example",
		}, &result)

		if len(result) != 1 {
			t.Fatalf("expected one symbol for Alice Example, got %d", len(result))
		}
		if result[0].Name != "user-alice" {
			t.Fatalf("expected symbol name user-alice, got %q", result[0].Name)
		}
		if got := result[0].Location.URI.Filename(); got != userPath {
			t.Fatalf("expected Alice location %s, got %s", userPath, got)
		}
	})

	t.Run("keep roots separate", func(t *testing.T) {
		server, rootA, rootB := initializeMultiRootServer(t)
		pathA := filepath.Join(rootA, "data", "users", "alice.yaml")
		pathB := filepath.Join(rootB, "data", "users", "alice.yaml")

		var result []protocol.SymbolInformation
		callServer(t, server, protocol.MethodWorkspaceSymbol, 3, &protocol.WorkspaceSymbolParams{
			Query: "shared-user",
		}, &result)

		if len(result) != 2 {
			t.Fatalf("expected two shared-user symbols, got %d", len(result))
		}
		expectSymbolPaths(t, result, []string{pathA, pathB})
		if !strings.Contains(result[0].ContainerName, "root-") || !strings.Contains(result[1].ContainerName, "root-") {
			t.Fatalf("expected root-qualified containers, got %q and %q", result[0].ContainerName, result[1].ContainerName)
		}
	})

	t.Run("deduplicate descendant symbols", func(t *testing.T) {
		server, _ := initializeInheritanceServer(t)

		var result []protocol.SymbolInformation
		callServer(t, server, protocol.MethodWorkspaceSymbol, 6, &protocol.WorkspaceSymbolParams{
			Query: "dog-1",
		}, &result)

		if len(result) != 1 {
			t.Fatalf("expected one symbol for dog-1, got %d", len(result))
		}
		if result[0].Name != "dog-1" {
			t.Fatalf("expected symbol dog-1, got %q", result[0].Name)
		}
	})
}

type expectedLocation struct {
	path string
	rng  protocol.Range
}

func initializeMultiRootServer(t *testing.T) (*Server, string, string) {
	t.Helper()

	server := NewServer(Options{Logger: testLogger()})
	base, err := filepath.Abs(filepath.Join("..", "workspace", "testdata", "phase4", "multi-root"))
	if err != nil {
		t.Fatalf("filepath.Abs(base): %v", err)
	}
	rootA := filepath.Join(base, "root-a")
	rootB := filepath.Join(base, "root-b")

	callServer(t, server, protocol.MethodInitialize, 1, &protocol.InitializeParams{
		WorkspaceFolders: []protocol.WorkspaceFolder{
			{URI: string(uri.File(rootA)), Name: "root-a"},
			{URI: string(uri.File(rootB)), Name: "root-b"},
		},
	}, (*protocol.InitializeResult)(nil))

	return server, rootA, rootB
}

func initializeInheritanceServer(t *testing.T) (*Server, string) {
	t.Helper()

	root := t.TempDir()
	cfg := `mergeway:
  version: 1

entities:
  Animal:
    identifier: id
    fields:
      id: string
      name: string
  Dog:
    extends: Animal
    include:
      - data/dogs/*.yaml
    fields:
      breed: string
  Kennel:
    identifier: id
    include:
      - data/kennels/*.yaml
    fields:
      id: string
      resident:
        type: Animal
`
	if err := os.WriteFile(filepath.Join(root, "mergeway.yaml"), []byte(cfg), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(root, "data", "dogs"), 0o755); err != nil {
		t.Fatalf("mkdir dogs: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(root, "data", "kennels"), 0o755); err != nil {
		t.Fatalf("mkdir kennels: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "data", "dogs", "dog.yaml"), []byte("id: dog-1\nname: Fido\nbreed: collie\n"), 0o644); err != nil {
		t.Fatalf("write dog: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "data", "kennels", "kennel.yaml"), []byte("id: kennel-1\nresident: dog-1\n"), 0o644); err != nil {
		t.Fatalf("write kennel: %v", err)
	}

	server := NewServer(Options{Logger: testLogger()})
	callServer(t, server, protocol.MethodInitialize, 1, map[string]any{
		"rootUri": string(uri.File(root)),
	}, (*protocol.InitializeResult)(nil))
	return server, root
}

func expectLocation(t *testing.T, location protocol.Location, path string, rng protocol.Range) {
	t.Helper()

	if got := location.URI.Filename(); got != path {
		t.Fatalf("expected location path %s, got %s", path, got)
	}
	if location.Range != rng {
		t.Fatalf("expected location range %+v, got %+v", rng, location.Range)
	}
}

func expectLocations(t *testing.T, locations []protocol.Location, expected []expectedLocation) {
	t.Helper()

	if len(locations) != len(expected) {
		t.Fatalf("expected %d locations, got %d", len(expected), len(locations))
	}
	for _, want := range expected {
		found := false
		for _, location := range locations {
			if location.URI.Filename() == want.path && location.Range == want.rng {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("expected location %s %+v in %+v", want.path, want.rng, locations)
		}
	}
}

func containsDocumentSymbol(symbols []protocol.DocumentSymbol, want string) bool {
	for _, symbol := range symbols {
		if symbol.Name == want {
			return true
		}
	}
	return false
}

func documentSymbolNames(symbols []protocol.DocumentSymbol) []string {
	names := make([]string, 0, len(symbols))
	for _, symbol := range symbols {
		names = append(names, symbol.Name)
	}
	return names
}

func expectSymbolPaths(t *testing.T, symbols []protocol.SymbolInformation, paths []string) {
	t.Helper()

	if len(symbols) != len(paths) {
		t.Fatalf("expected %d symbols, got %d", len(paths), len(symbols))
	}

	seen := make(map[string]struct{}, len(symbols))
	for _, symbol := range symbols {
		seen[symbol.Location.URI.Filename()] = struct{}{}
	}
	for _, path := range paths {
		if _, ok := seen[path]; !ok {
			t.Fatalf("expected symbol path %s in result set", path)
		}
	}
}
