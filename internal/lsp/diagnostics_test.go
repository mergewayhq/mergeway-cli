package lsp

import (
	"context"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"sync"
	"testing"

	"github.com/mergewayhq/mergeway-cli/internal/config"
	"github.com/mergewayhq/mergeway-cli/internal/validation"
	"go.lsp.dev/jsonrpc2"
	"go.lsp.dev/protocol"
	"go.lsp.dev/uri"
)

func TestHandleInitializePublishesSavedWorkspaceDiagnosticsParity(t *testing.T) {
	fixtures := []string{"format_error", "schema_error", "reference_error"}

	for _, fixture := range fixtures {
		t.Run(fixture, func(t *testing.T) {
			root := filepath.Join("..", "validation", "testdata", fixture)
			capture := &diagnosticCapture{}
			server := NewServer(Options{
				Logger:             testLogger(),
				PublishDiagnostics: capture.PublishDiagnostics,
			})

			initializeServerForDiagnostics(t, server, root)

			got := capture.messagesByPath()
			want := expectedValidationMessagesByPath(t, root)
			if !reflect.DeepEqual(got, want) {
				t.Fatalf("unexpected diagnostics\nwant: %#v\ngot: %#v", want, got)
			}
		})
	}
}

func TestHandleInitializePublishesConfigDiagnosticsForInvalidConfig(t *testing.T) {
	root := t.TempDir()
	configBody := `mergeway:
  version: 1

entities:
  User:
    include:
      - data/users/*.yaml
    identifier: id
    fields:
      id:
        type: string
      owner:
        type: Team
`
	if err := os.WriteFile(filepath.Join(root, "mergeway.yaml"), []byte(configBody), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	capture := &diagnosticCapture{}
	server := NewServer(Options{
		Logger:             testLogger(),
		PublishDiagnostics: capture.PublishDiagnostics,
	})

	initializeServerForDiagnostics(t, server, root)

	diagnostics := capture.latestByPath()[filepath.Join(root, "mergeway.yaml")]
	if diagnostics == nil || len(diagnostics.Diagnostics) != 1 {
		t.Fatalf("expected one config diagnostic, got %#v", diagnostics)
	}
	diag := diagnostics.Diagnostics[0]
	if !strings.Contains(diag.Message, `references unknown type "Team"`) {
		t.Fatalf("unexpected diagnostic message: %s", diag.Message)
	}
	if diag.Range.Start.Line == 0 {
		t.Fatalf("expected config diagnostic to target the owner field, got range %+v", diag.Range)
	}
}

func TestHandleDidChangePublishesOpenDocumentSchemaDiagnosticsAndClears(t *testing.T) {
	root := filepath.Join("..", "workspace", "testdata", "phase4", "valid-basic")
	targetPath := absTestPath(t, filepath.Join(root, "data", "users", "alice.yaml"))
	targetURI := uri.File(targetPath)

	capture := &diagnosticCapture{}
	server := NewServer(Options{
		Logger:             testLogger(),
		PublishDiagnostics: capture.PublishDiagnostics,
	})
	initializeServerForDiagnostics(t, server, root)
	capture.Reset()

	openReq, err := jsonrpc2.NewCall(jsonrpc2.NewNumberID(2), protocol.MethodTextDocumentDidOpen, &protocol.DidOpenTextDocumentParams{
		TextDocument: protocol.TextDocumentItem{
			URI:        protocol.DocumentURI(targetURI),
			LanguageID: "yaml",
			Version:    1,
			Text:       "id: user-1\nname: Alice\n",
		},
	})
	if err != nil {
		t.Fatalf("NewCall(didOpen): %v", err)
	}
	if err := server.Handle(context.Background(), captureReply[struct{}](t, nil), openReq); err != nil {
		t.Fatalf("Handle(didOpen): %v", err)
	}

	changeReq, err := jsonrpc2.NewCall(jsonrpc2.NewNumberID(3), protocol.MethodTextDocumentDidChange, &protocol.DidChangeTextDocumentParams{
		TextDocument: protocol.VersionedTextDocumentIdentifier{
			TextDocumentIdentifier: protocol.TextDocumentIdentifier{URI: protocol.DocumentURI(targetURI)},
			Version:                2,
		},
		ContentChanges: []protocol.TextDocumentContentChangeEvent{
			{Text: "id: user-1\nname: 7\n"},
		},
	})
	if err != nil {
		t.Fatalf("NewCall(didChange invalid): %v", err)
	}
	if err := server.Handle(context.Background(), captureReply[struct{}](t, nil), changeReq); err != nil {
		t.Fatalf("Handle(didChange invalid): %v", err)
	}
	if err := server.runtime.FlushReload(); err != nil {
		t.Fatalf("FlushReload(invalid): %v", err)
	}

	invalid := capture.latestByPath()[targetPath]
	if invalid == nil || len(invalid.Diagnostics) != 1 {
		t.Fatalf("expected one schema diagnostic, got %#v", invalid)
	}
	if got := invalid.Diagnostics[0].Message; !strings.Contains(got, `field "name" must be string`) {
		t.Fatalf("unexpected schema diagnostic: %s", got)
	}
	if invalid.Diagnostics[0].Range.Start.Line != 1 {
		t.Fatalf("expected schema diagnostic on second line, got %+v", invalid.Diagnostics[0].Range)
	}

	capture.Reset()

	fixReq, err := jsonrpc2.NewCall(jsonrpc2.NewNumberID(4), protocol.MethodTextDocumentDidChange, &protocol.DidChangeTextDocumentParams{
		TextDocument: protocol.VersionedTextDocumentIdentifier{
			TextDocumentIdentifier: protocol.TextDocumentIdentifier{URI: protocol.DocumentURI(targetURI)},
			Version:                3,
		},
		ContentChanges: []protocol.TextDocumentContentChangeEvent{
			{Text: "id: user-1\nname: Alice\n"},
		},
	})
	if err != nil {
		t.Fatalf("NewCall(didChange fix): %v", err)
	}
	if err := server.Handle(context.Background(), captureReply[struct{}](t, nil), fixReq); err != nil {
		t.Fatalf("Handle(didChange fix): %v", err)
	}
	if err := server.runtime.FlushReload(); err != nil {
		t.Fatalf("FlushReload(fix): %v", err)
	}

	cleared := capture.latestByPath()[targetPath]
	if cleared == nil {
		t.Fatalf("expected clearing diagnostics publish for %s", targetPath)
	}
	if len(cleared.Diagnostics) != 0 {
		t.Fatalf("expected diagnostics to clear, got %#v", cleared.Diagnostics)
	}
	if cleared.Version != 3 {
		t.Fatalf("expected cleared diagnostics to carry document version 3, got %d", cleared.Version)
	}
}

func TestHandleDidChangePublishesOpenDocumentSyntaxDiagnostics(t *testing.T) {
	root := filepath.Join("..", "workspace", "testdata", "phase4", "valid-basic")
	targetPath := absTestPath(t, filepath.Join(root, "data", "users", "alice.yaml"))
	targetURI := uri.File(targetPath)

	capture := &diagnosticCapture{}
	server := NewServer(Options{
		Logger:             testLogger(),
		PublishDiagnostics: capture.PublishDiagnostics,
	})
	initializeServerForDiagnostics(t, server, root)
	capture.Reset()

	openReq, err := jsonrpc2.NewCall(jsonrpc2.NewNumberID(2), protocol.MethodTextDocumentDidOpen, &protocol.DidOpenTextDocumentParams{
		TextDocument: protocol.TextDocumentItem{
			URI:        protocol.DocumentURI(targetURI),
			LanguageID: "yaml",
			Version:    1,
			Text:       "id: user-1\nname: Alice\n",
		},
	})
	if err != nil {
		t.Fatalf("NewCall(didOpen): %v", err)
	}
	if err := server.Handle(context.Background(), captureReply[struct{}](t, nil), openReq); err != nil {
		t.Fatalf("Handle(didOpen): %v", err)
	}

	changeReq, err := jsonrpc2.NewCall(jsonrpc2.NewNumberID(3), protocol.MethodTextDocumentDidChange, &protocol.DidChangeTextDocumentParams{
		TextDocument: protocol.VersionedTextDocumentIdentifier{
			TextDocumentIdentifier: protocol.TextDocumentIdentifier{URI: protocol.DocumentURI(targetURI)},
			Version:                2,
		},
		ContentChanges: []protocol.TextDocumentContentChangeEvent{
			{Text: "id: user-1\nname: [\n"},
		},
	})
	if err != nil {
		t.Fatalf("NewCall(didChange syntax): %v", err)
	}
	if err := server.Handle(context.Background(), captureReply[struct{}](t, nil), changeReq); err != nil {
		t.Fatalf("Handle(didChange syntax): %v", err)
	}
	if err := server.runtime.FlushReload(); err != nil {
		t.Fatalf("FlushReload(syntax): %v", err)
	}

	diagnostics := capture.latestByPath()[targetPath]
	if diagnostics == nil || len(diagnostics.Diagnostics) != 1 {
		t.Fatalf("expected one syntax diagnostic, got %#v", diagnostics)
	}
	diag := diagnostics.Diagnostics[0]
	if !strings.Contains(diag.Message, "unable to parse") {
		t.Fatalf("unexpected syntax diagnostic: %s", diag.Message)
	}
	if diag.Range.Start.Line != 1 {
		t.Fatalf("expected syntax diagnostic on second line, got %+v", diag.Range)
	}
}

func initializeServerForDiagnostics(t *testing.T, server *Server, root string) {
	t.Helper()

	rootURI := uri.File(root)
	initReq, err := jsonrpc2.NewCall(jsonrpc2.NewNumberID(1), protocol.MethodInitialize, map[string]any{
		"rootUri": string(rootURI),
	})
	if err != nil {
		t.Fatalf("NewCall(initialize): %v", err)
	}
	if err := server.Handle(context.Background(), captureReply[protocol.InitializeResult](t, nil), initReq); err != nil {
		t.Fatalf("Handle(initialize): %v", err)
	}
}

func expectedValidationMessagesByPath(t *testing.T, root string) map[string][]string {
	t.Helper()

	root = absTestPath(t, root)

	cfgPath := filepath.Join(root, "mergeway.yaml")
	cfg, err := config.Load(cfgPath)
	if err != nil {
		t.Fatalf("config.Load(%s): %v", cfgPath, err)
	}

	result, err := validation.Validate(root, cfg, validation.Options{})
	if err != nil {
		t.Fatalf("validation.Validate(%s): %v", root, err)
	}

	grouped := make(map[string][]string)
	for _, errItem := range result.Errors {
		loc, ok := resolveValidationLocation(root, errItem.File)
		if !ok {
			t.Fatalf("resolveValidationLocation(%q): false", errItem.File)
		}
		grouped[loc.path] = append(grouped[loc.path], errItem.Message)
	}

	for path := range grouped {
		sort.Strings(grouped[path])
	}
	return grouped
}

func absTestPath(t *testing.T, path string) string {
	t.Helper()

	resolved, err := filepath.Abs(path)
	if err != nil {
		t.Fatalf("filepath.Abs(%s): %v", path, err)
	}
	return resolved
}

type diagnosticCapture struct {
	mu     sync.Mutex
	params []*protocol.PublishDiagnosticsParams
}

func (c *diagnosticCapture) PublishDiagnostics(_ context.Context, params *protocol.PublishDiagnosticsParams) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	cloned := &protocol.PublishDiagnosticsParams{
		URI:         params.URI,
		Version:     params.Version,
		Diagnostics: append([]protocol.Diagnostic(nil), params.Diagnostics...),
	}
	c.params = append(c.params, cloned)
	return nil
}

func (c *diagnosticCapture) Reset() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.params = nil
}

func (c *diagnosticCapture) latestByPath() map[string]*protocol.PublishDiagnosticsParams {
	c.mu.Lock()
	defer c.mu.Unlock()

	result := make(map[string]*protocol.PublishDiagnosticsParams)
	for _, params := range c.params {
		result[params.URI.Filename()] = params
	}
	return result
}

func (c *diagnosticCapture) messagesByPath() map[string][]string {
	grouped := make(map[string][]string)
	for path, params := range c.latestByPath() {
		for _, diag := range params.Diagnostics {
			grouped[path] = append(grouped[path], diag.Message)
		}
		sort.Strings(grouped[path])
	}
	return grouped
}
