package lsp

import (
	"path/filepath"
	"strings"
	"testing"

	"go.lsp.dev/protocol"
	"go.lsp.dev/uri"
)

func TestHandleCodeActionOffersConservativeQuickFixes(t *testing.T) {
	t.Run("misspelled field rename", func(t *testing.T) {
		root := filepath.Join("..", "workspace", "testdata", "phase4", "valid-basic")
		server, capture, absRoot := initializeCodeActionServer(t, root)
		path := filepath.Join(absRoot, "data", "users", "alice.yaml")
		text := "id: user-1\nnmae: Alice\n"

		actions := openDocumentAndCodeActions(t, server, capture, path, text)
		action := requireCodeAction(t, actions, `Rename field "nmae" to "name"`)
		edit := requireSingleEdit(t, action, path)

		if edit.NewText != "name" {
			t.Fatalf("expected rename replacement name, got %q", edit.NewText)
		}
		if edit.Range.Start.Line != 1 {
			t.Fatalf("expected rename on second line, got %+v", edit.Range)
		}
	})

	t.Run("missing required field insertion", func(t *testing.T) {
		server, capture, root := initializeCodeActionServer(t, filepath.Join("..", "..", "examples", "full"))
		path := filepath.Join(root, "data", "posts", "launch.yaml")
		text := "id: post-1\nstatus: DRAFT\nauthor: user-alice\n"

		actions := openDocumentAndCodeActions(t, server, capture, path, text)
		action := requireCodeAction(t, actions, `Insert missing field "title"`)
		edit := requireSingleEdit(t, action, path)

		if edit.NewText != "title: \"\"\n" {
			t.Fatalf("expected title insertion, got %q", edit.NewText)
		}
		if edit.Range.Start.Line != 1 || edit.Range.Start.Character != 0 {
			t.Fatalf("expected insertion before status line, got %+v", edit.Range)
		}
	})

	t.Run("enum replacement", func(t *testing.T) {
		server, capture, root := initializeCodeActionServer(t, filepath.Join("..", "..", "examples", "full"))
		path := filepath.Join(root, "data", "posts", "launch.yaml")
		text := "id: post-1\ntitle: Launch Day\nstatus: PUBLISHD\nauthor: user-alice\n"

		actions := openDocumentAndCodeActions(t, server, capture, path, text)
		action := requireCodeAction(t, actions, `Replace with "PUBLISHED"`)
		edit := requireSingleEdit(t, action, path)

		wantRange := scalarFieldValueRange(t, text, "status")
		if edit.Range != wantRange {
			t.Fatalf("expected status value range %+v, got %+v", wantRange, edit.Range)
		}
		if edit.NewText != "PUBLISHED" {
			t.Fatalf("expected enum replacement PUBLISHED, got %q", edit.NewText)
		}
	})

	t.Run("reference replacement", func(t *testing.T) {
		server, capture, root := initializeCodeActionServer(t, filepath.Join("..", "..", "examples", "full"))
		path := filepath.Join(root, "data", "posts", "launch.yaml")
		text := "id: post-1\ntitle: Launch Day\nstatus: PUBLISHED\nauthor: user-ailce\n"

		actions := openDocumentAndCodeActions(t, server, capture, path, text)
		action := requireCodeAction(t, actions, `Replace with "user-alice"`)
		edit := requireSingleEdit(t, action, path)

		wantRange := scalarFieldValueRange(t, text, "author")
		if edit.Range != wantRange {
			t.Fatalf("expected author value range %+v, got %+v", wantRange, edit.Range)
		}
		if edit.NewText != "user-alice" {
			t.Fatalf("expected reference replacement user-alice, got %q", edit.NewText)
		}
	})
}

func initializeCodeActionServer(t *testing.T, root string) (*Server, *diagnosticCapture, string) {
	t.Helper()

	absRoot := absTestPath(t, root)
	capture := &diagnosticCapture{}
	server := NewServer(Options{
		Logger:             testLogger(),
		PublishDiagnostics: capture.PublishDiagnostics,
	})
	initializeServerForDiagnostics(t, server, absRoot)
	capture.Reset()
	return server, capture, absRoot
}

func openDocumentAndCodeActions(t *testing.T, server *Server, capture *diagnosticCapture, path, text string) []protocol.CodeAction {
	t.Helper()

	openDocument(t, server, path, languageForPath(path), 1, text)
	if err := server.runtime.FlushReload(); err != nil {
		t.Fatalf("FlushReload: %v", err)
	}

	params := capture.latestByPath()[path]
	if params == nil || len(params.Diagnostics) == 0 {
		t.Fatalf("expected diagnostics for %s, got %#v", path, params)
	}

	var actions []protocol.CodeAction
	callServer(t, server, protocol.MethodTextDocumentCodeAction, 2, &protocol.CodeActionParams{
		TextDocument: protocol.TextDocumentIdentifier{URI: protocol.DocumentURI(uri.File(path))},
		Range:        params.Diagnostics[0].Range,
		Context: protocol.CodeActionContext{
			Diagnostics: params.Diagnostics,
			Only:        []protocol.CodeActionKind{protocol.QuickFix},
		},
	}, &actions)
	return actions
}

func requireCodeAction(t *testing.T, actions []protocol.CodeAction, title string) protocol.CodeAction {
	t.Helper()

	for _, action := range actions {
		if action.Title == title {
			return action
		}
	}
	t.Fatalf("expected code action %q in %#v", title, actions)
	return protocol.CodeAction{}
}

func requireSingleEdit(t *testing.T, action protocol.CodeAction, path string) protocol.TextEdit {
	t.Helper()

	if action.Edit == nil {
		t.Fatalf("expected edit for action %q", action.Title)
	}

	edits := action.Edit.Changes[protocol.DocumentURI(uri.File(path))]
	if len(edits) != 1 {
		t.Fatalf("expected one edit for %s, got %#v", path, edits)
	}
	return edits[0]
}

func languageForPath(path string) string {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".json":
		return "json"
	default:
		return "yaml"
	}
}
