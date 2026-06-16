package lsp

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"go.lsp.dev/jsonrpc2"
	"go.lsp.dev/protocol"
	"go.lsp.dev/uri"
	"gopkg.in/yaml.v3"
)

func TestHandleCompletionProvidesContextAwareSuggestions(t *testing.T) {
	t.Run("field names", func(t *testing.T) {
		server, root := initializeExampleFullServer(t)
		path := filepath.Join(root, "data", "users", "alice.yaml")
		text, pos := cursorContent("id: user-alice\nna|: \n")
		openDocument(t, server, path, "yaml", 1, text)

		var result protocol.CompletionList
		callServer(t, server, protocol.MethodTextDocumentCompletion, 2, &protocol.CompletionParams{
			TextDocumentPositionParams: protocol.TextDocumentPositionParams{
				TextDocument: protocol.TextDocumentIdentifier{URI: protocol.DocumentURI(uri.File(path))},
				Position:     pos,
			},
		}, &result)

		labels := completionLabels(result.Items)
		if !contains(labels, "name") {
			t.Fatalf("expected name field completion, got %v", labels)
		}
		if contains(labels, "id") {
			t.Fatalf("expected existing id field to be omitted, got %v", labels)
		}
	})

	t.Run("enum values", func(t *testing.T) {
		server, root := initializeExampleFullServer(t)
		path := filepath.Join(root, "data", "posts", "launch.yaml")
		text, pos := cursorContent("id: post-001\ntitle: Launch Day\nstatus: P|")
		openDocument(t, server, path, "yaml", 1, text)

		var result protocol.CompletionList
		callServer(t, server, protocol.MethodTextDocumentCompletion, 2, &protocol.CompletionParams{
			TextDocumentPositionParams: protocol.TextDocumentPositionParams{
				TextDocument: protocol.TextDocumentIdentifier{URI: protocol.DocumentURI(uri.File(path))},
				Position:     pos,
			},
		}, &result)

		labels := completionLabels(result.Items)
		if len(labels) != 1 || labels[0] != "PUBLISHED" {
			t.Fatalf("expected enum completion [PUBLISHED], got %v", labels)
		}
	})

	t.Run("references", func(t *testing.T) {
		server, root := initializeExampleFullServer(t)
		path := filepath.Join(root, "data", "posts", "launch.yaml")
		text, pos := cursorContent("id: post-001\nauthor: user-|")
		openDocument(t, server, path, "yaml", 1, text)

		var result protocol.CompletionList
		callServer(t, server, protocol.MethodTextDocumentCompletion, 2, &protocol.CompletionParams{
			TextDocumentPositionParams: protocol.TextDocumentPositionParams{
				TextDocument: protocol.TextDocumentIdentifier{URI: protocol.DocumentURI(uri.File(path))},
				Position:     pos,
			},
		}, &result)

		labels := completionLabels(result.Items)
		if !contains(labels, "user-alice") || !contains(labels, "user-bob") {
			t.Fatalf("expected user reference completions, got %v", labels)
		}
	})

	t.Run("config type values", func(t *testing.T) {
		server, root := initializeExampleFullServer(t)
		path := filepath.Join(root, "entities", "Post.yaml")
		text, pos := cursorContent(strings.TrimSpace(`
mergeway:
  version: 1

entities:
  Post:
    fields:
      sample:
        type: U|
`) + "\n")
		openDocument(t, server, path, "yaml", 1, text)

		var result protocol.CompletionList
		callServer(t, server, protocol.MethodTextDocumentCompletion, 2, &protocol.CompletionParams{
			TextDocumentPositionParams: protocol.TextDocumentPositionParams{
				TextDocument: protocol.TextDocumentIdentifier{URI: protocol.DocumentURI(uri.File(path))},
				Position:     pos,
			},
		}, &result)

		labels := completionLabels(result.Items)
		if !contains(labels, "User") {
			t.Fatalf("expected User type completion, got %v", labels)
		}
	})
}

func TestHandleHoverReturnsSchemaAndEntitySummaries(t *testing.T) {
	server, root := initializeExampleFullServer(t)
	postPath := filepath.Join(root, "data", "posts", "launch.yaml")
	postContent := readFile(t, postPath)

	t.Run("known field", func(t *testing.T) {
		var result protocol.Hover
		callServer(t, server, protocol.MethodTextDocumentHover, 2, &protocol.HoverParams{
			TextDocumentPositionParams: protocol.TextDocumentPositionParams{
				TextDocument: protocol.TextDocumentIdentifier{URI: protocol.DocumentURI(uri.File(postPath))},
				Position:     positionInContent(t, postContent, "status"),
			},
		}, &result)

		if !strings.Contains(result.Contents.Value, "**status**") || !strings.Contains(result.Contents.Value, "Publication lifecycle state") {
			t.Fatalf("unexpected field hover: %q", result.Contents.Value)
		}
		if !strings.Contains(result.Contents.Value, "`DRAFT`, `PUBLISHED`") {
			t.Fatalf("expected enum values in hover, got %q", result.Contents.Value)
		}
	})

	t.Run("entity id", func(t *testing.T) {
		userPath := filepath.Join(root, "data", "users", "alice.yaml")
		userContent := readFile(t, userPath)

		var result protocol.Hover
		callServer(t, server, protocol.MethodTextDocumentHover, 3, &protocol.HoverParams{
			TextDocumentPositionParams: protocol.TextDocumentPositionParams{
				TextDocument: protocol.TextDocumentIdentifier{URI: protocol.DocumentURI(uri.File(userPath))},
				Position:     positionInContent(t, userContent, "user-alice"),
			},
		}, &result)

		if !strings.Contains(result.Contents.Value, "**User** `user-alice`") || !strings.Contains(result.Contents.Value, "alice@example.com") {
			t.Fatalf("unexpected entity hover: %q", result.Contents.Value)
		}
	})

	t.Run("reference", func(t *testing.T) {
		var result protocol.Hover
		callServer(t, server, protocol.MethodTextDocumentHover, 4, &protocol.HoverParams{
			TextDocumentPositionParams: protocol.TextDocumentPositionParams{
				TextDocument: protocol.TextDocumentIdentifier{URI: protocol.DocumentURI(uri.File(postPath))},
				Position:     positionInContent(t, postContent, "user-alice"),
			},
		}, &result)

		if !strings.Contains(result.Contents.Value, "**User** `user-alice`") || !strings.Contains(result.Contents.Value, "Alice Example") {
			t.Fatalf("unexpected reference hover: %q", result.Contents.Value)
		}
	})
}

func TestHandleDefinitionResolvesSavedAndUnsavedReferences(t *testing.T) {
	t.Run("saved reference", func(t *testing.T) {
		server, root := initializeExampleFullServer(t)
		postPath := filepath.Join(root, "data", "posts", "launch.yaml")
		postContent := readFile(t, postPath)
		targetPath := filepath.Join(root, "data", "users", "alice.yaml")
		targetContent := readFile(t, targetPath)

		var result []protocol.Location
		callServer(t, server, protocol.MethodTextDocumentDefinition, 2, &protocol.DefinitionParams{
			TextDocumentPositionParams: protocol.TextDocumentPositionParams{
				TextDocument: protocol.TextDocumentIdentifier{URI: protocol.DocumentURI(uri.File(postPath))},
				Position:     positionInContent(t, postContent, "user-alice"),
			},
		}, &result)

		if len(result) != 1 {
			t.Fatalf("expected one definition target, got %d", len(result))
		}
		if got := result[0].URI.Filename(); got != targetPath {
			t.Fatalf("expected target %s, got %s", targetPath, got)
		}
		wantRange := scalarFieldValueRange(t, targetContent, "id")
		if result[0].Range != wantRange {
			t.Fatalf("expected range %+v, got %+v", wantRange, result[0].Range)
		}
	})

	t.Run("unsaved overlay reference", func(t *testing.T) {
		server, root := initializeExampleFullServer(t)
		postPath := filepath.Join(root, "data", "posts", "launch.yaml")
		tagPath := filepath.Join(root, "data", "tags", "product.yaml")
		postText, pos := cursorContent("id: post-001\ntitle: Launch Day\ntags:\n  - tag-new|")
		tagText := strings.TrimSpace(`
type: Tag
items:
  - id: tag-new
    label: Temporary Tag
  - id: tag-update
    label: Product Update
`) + "\n"

		openDocument(t, server, postPath, "yaml", 1, postText)
		openDocument(t, server, tagPath, "yaml", 1, tagText)
		if err := server.runtime.FlushReload(); err != nil {
			t.Fatalf("FlushReload: %v", err)
		}

		var result []protocol.Location
		callServer(t, server, protocol.MethodTextDocumentDefinition, 4, &protocol.DefinitionParams{
			TextDocumentPositionParams: protocol.TextDocumentPositionParams{
				TextDocument: protocol.TextDocumentIdentifier{URI: protocol.DocumentURI(uri.File(postPath))},
				Position:     pos,
			},
		}, &result)

		if len(result) != 1 {
			t.Fatalf("expected one definition target, got %d", len(result))
		}
		if got := result[0].URI.Filename(); got != tagPath {
			t.Fatalf("expected target %s, got %s", tagPath, got)
		}
		wantRange := sequenceItemFieldValueRange(t, tagText, 0, "id")
		if result[0].Range != wantRange {
			t.Fatalf("expected range %+v, got %+v", wantRange, result[0].Range)
		}
	})
}

func initializeExampleFullServer(t *testing.T) (*Server, string) {
	t.Helper()

	server := NewServer(Options{Logger: testLogger()})
	root, err := filepath.Abs(filepath.Join("..", "..", "examples", "full"))
	if err != nil {
		t.Fatalf("filepath.Abs(root): %v", err)
	}
	callServer(t, server, protocol.MethodInitialize, 1, map[string]any{
		"rootUri": string(uri.File(root)),
	}, (*protocol.InitializeResult)(nil))
	return server, root
}

func callServer[T any](t *testing.T, server *Server, method string, id int32, params any, target *T) {
	t.Helper()

	req, err := jsonrpc2.NewCall(jsonrpc2.NewNumberID(id), method, params)
	if err != nil {
		t.Fatalf("NewCall(%s): %v", method, err)
	}
	if err := server.Handle(context.Background(), captureReply(t, target), req); err != nil {
		t.Fatalf("Handle(%s): %v", method, err)
	}
}

func openDocument(t *testing.T, server *Server, path, languageID string, version int32, text string) {
	t.Helper()

	callServer(t, server, protocol.MethodTextDocumentDidOpen, 1, &protocol.DidOpenTextDocumentParams{
		TextDocument: protocol.TextDocumentItem{
			URI:        protocol.DocumentURI(uri.File(path)),
			LanguageID: protocol.LanguageIdentifier(languageID),
			Version:    version,
			Text:       text,
		},
	}, (*struct{})(nil))
}

func cursorContent(text string) (string, protocol.Position) {
	lines := strings.Split(text, "\n")
	for lineIndex, line := range lines {
		if column := strings.Index(line, "|"); column >= 0 {
			lines[lineIndex] = strings.Replace(line, "|", "", 1)
			return strings.Join(lines, "\n"), protocol.Position{
				Line:      uint32(lineIndex),
				Character: uint32(column),
			}
		}
	}
	panic("cursor marker not found")
}

func completionLabels(items []protocol.CompletionItem) []string {
	labels := make([]string, 0, len(items))
	for _, item := range items {
		labels = append(labels, item.Label)
	}
	return labels
}

func contains(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

func readFile(t *testing.T, path string) string {
	t.Helper()

	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile(%s): %v", path, err)
	}
	return string(body)
}

func positionInContent(t *testing.T, content, needle string) protocol.Position {
	t.Helper()

	offset := strings.Index(content, needle)
	if offset < 0 {
		t.Fatalf("needle %q not found", needle)
	}

	before := content[:offset]
	line := strings.Count(before, "\n")
	column := len(before) - strings.LastIndex(before, "\n") - 1
	return protocol.Position{Line: uint32(line), Character: uint32(column)}
}

func scalarFieldValueRange(t *testing.T, content, field string) protocol.Range {
	t.Helper()

	doc, ok := parseDocumentNode([]byte(content))
	if !ok {
		t.Fatalf("parseDocumentNode failed for field %s", field)
	}
	root := documentRoot(doc)
	_, valueNode := mappingEntry(root, field)
	if valueNode == nil {
		t.Fatalf("field %s not found", field)
	}
	return nodeRange(valueNode)
}

func sequenceItemFieldValueRange(t *testing.T, content string, itemIndex int, field string) protocol.Range {
	t.Helper()

	doc, ok := parseDocumentNode([]byte(content))
	if !ok {
		t.Fatalf("parseDocumentNode failed for field %s", field)
	}
	root := documentRoot(doc)
	_, itemsNode := mappingEntry(root, "items")
	if itemsNode == nil || itemsNode.Kind != yaml.SequenceNode || itemIndex >= len(itemsNode.Content) {
		t.Fatalf("items[%d] not found", itemIndex)
	}
	_, valueNode := mappingEntry(itemsNode.Content[itemIndex], field)
	if valueNode == nil {
		t.Fatalf("field %s not found on items[%d]", field, itemIndex)
	}
	return nodeRange(valueNode)
}
