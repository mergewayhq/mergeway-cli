package lsp

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"go.lsp.dev/jsonrpc2"
	"go.lsp.dev/protocol"
	"go.lsp.dev/uri"
)

func TestRunInitializeShutdownExit(t *testing.T) {
	clientConn, serverConn := net.Pipe()
	mustCleanupClose(t, clientConn)

	done := make(chan serveResult, 1)
	go func() {
		code, err := Run(context.Background(), serverConn, Options{Logger: testLogger()})
		done <- serveResult{code: code, err: err}
	}()

	client := jsonrpc2.NewConn(jsonrpc2.NewStream(clientConn))
	client.Go(context.Background(), protocol.Handlers(jsonrpc2.MethodNotFoundHandler))
	mustCleanupClose(t, client)

	var result protocol.InitializeResult
	_, err := client.Call(context.Background(), protocol.MethodInitialize, &protocol.InitializeParams{
		RootURI: protocol.DocumentURI(uri.File(t.TempDir())),
		Trace:   protocol.TraceVerbose,
	}, &result)
	if err != nil {
		t.Fatalf("initialize: %v", err)
	}
	if result.ServerInfo == nil || result.ServerInfo.Name != "mergeway-lsp" {
		t.Fatalf("unexpected server info: %+v", result.ServerInfo)
	}
	syncOptions, ok := result.Capabilities.TextDocumentSync.(map[string]any)
	if !ok || syncOptions == nil {
		t.Fatalf("expected text sync options, got %#v", result.Capabilities.TextDocumentSync)
	}
	if syncOptions["openClose"] != true || syncOptions["change"] != float64(protocol.TextDocumentSyncKindFull) {
		t.Fatalf("expected full-document open/close sync, got %+v", syncOptions)
	}
	if result.Capabilities.CompletionProvider == nil {
		t.Fatalf("expected completion capability to be advertised")
	}
	if result.Capabilities.HoverProvider != true {
		t.Fatalf("expected hover capability to be advertised, got %+v", result.Capabilities.HoverProvider)
	}
	if result.Capabilities.DefinitionProvider != true {
		t.Fatalf("expected definition capability to be advertised, got %+v", result.Capabilities.DefinitionProvider)
	}
	if result.Capabilities.Workspace == nil || result.Capabilities.Workspace.WorkspaceFolders == nil || !result.Capabilities.Workspace.WorkspaceFolders.Supported {
		t.Fatalf("expected workspace folder capability to be advertised, got %+v", result.Capabilities.Workspace)
	}

	if _, err := client.Call(context.Background(), protocol.MethodShutdown, nil, nil); err != nil {
		t.Fatalf("shutdown: %v", err)
	}
	if err := client.Notify(context.Background(), protocol.MethodExit, nil); err != nil {
		t.Fatalf("exit notify: %v", err)
	}

	res := waitServeResult(t, done)
	if res.err != nil {
		t.Fatalf("Run returned error: %v", res.err)
	}
	if res.code != 0 {
		t.Fatalf("expected clean shutdown exit code, got %d", res.code)
	}
}

func TestHandleInitializeBuildsRootsFromWorkspaceFolders(t *testing.T) {
	server := NewServer(Options{Logger: testLogger()})
	base := filepath.Join("..", "workspace", "testdata", "phase4")
	rootA := uri.File(filepath.Join(base, "multi-root", "root-a"))
	rootB := uri.File(filepath.Join(base, "multi-root", "root-b"))
	missing := uri.File(filepath.Join(base, "no-config"))

	req, err := jsonrpc2.NewCall(jsonrpc2.NewNumberID(1), protocol.MethodInitialize, &protocol.InitializeParams{
		WorkspaceFolders: []protocol.WorkspaceFolder{
			{URI: string(rootA), Name: "root-a"},
			{URI: string(rootB), Name: "root-b"},
			{URI: string(missing), Name: "no-config"},
		},
	})
	if err != nil {
		t.Fatalf("NewCall: %v", err)
	}

	var result protocol.InitializeResult
	if err := server.Handle(context.Background(), captureReply(t, &result), req); err != nil {
		t.Fatalf("Handle(initialize): %v", err)
	}

	if server.roots == nil {
		t.Fatalf("expected roots to be initialized")
	}
	if got := len(server.roots.Roots); got != 2 {
		t.Fatalf("expected 2 detected roots, got %d", got)
	}
	if got := len(server.roots.MissingRoots); got != 1 {
		t.Fatalf("expected 1 missing root, got %d", got)
	}
	if result.Capabilities.Workspace == nil || result.Capabilities.Workspace.WorkspaceFolders == nil {
		t.Fatalf("expected workspace capabilities in initialize result")
	}
}

func TestHandleInitializeSupportsNoConfigRootURI(t *testing.T) {
	server := NewServer(Options{Logger: testLogger()})
	missing := uri.File(filepath.Join("..", "workspace", "testdata", "phase4", "no-config"))

	req, err := jsonrpc2.NewCall(jsonrpc2.NewNumberID(1), protocol.MethodInitialize, map[string]any{
		"rootUri": string(missing),
	})
	if err != nil {
		t.Fatalf("NewCall: %v", err)
	}

	if err := server.Handle(context.Background(), captureReply[protocol.InitializeResult](t, nil), req); err != nil {
		t.Fatalf("Handle(initialize): %v", err)
	}
	if server.roots == nil {
		t.Fatalf("expected roots state to be initialized")
	}
	if got := len(server.roots.Roots); got != 0 {
		t.Fatalf("expected no detected roots, got %d", got)
	}
	if got := len(server.roots.MissingRoots); got != 1 {
		t.Fatalf("expected one missing root, got %d", got)
	}
}

func TestHandleDidOpenDidChangeDidClosePreferOpenBuffer(t *testing.T) {
	server := NewServer(Options{Logger: testLogger()})
	root := filepath.Join("..", "workspace", "testdata", "phase4", "valid-basic")
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

	targetPath := filepath.Join(root, "data", "users", "alice.yaml")
	targetURI := uri.File(targetPath)

	openReq, err := jsonrpc2.NewCall(jsonrpc2.NewNumberID(2), protocol.MethodTextDocumentDidOpen, &protocol.DidOpenTextDocumentParams{
		TextDocument: protocol.TextDocumentItem{
			URI:        protocol.DocumentURI(targetURI),
			LanguageID: "yaml",
			Version:    1,
			Text:       "id: user-1\nname: Unsaved Alice\n",
		},
	})
	if err != nil {
		t.Fatalf("NewCall(didOpen): %v", err)
	}
	if err := server.Handle(context.Background(), captureReply[struct{}](t, nil), openReq); err != nil {
		t.Fatalf("Handle(didOpen): %v", err)
	}
	if err := server.runtime.FlushReload(); err != nil {
		t.Fatalf("FlushReload(open): %v", err)
	}

	runtimeRoot := server.runtime.RootByPath(targetPath)
	if runtimeRoot == nil || runtimeRoot.Workspace == nil {
		t.Fatalf("expected runtime root after didOpen")
	}
	if got := runtimeRoot.Workspace.Find("User", "user-1")[0].Fields["name"]; got != "Unsaved Alice" {
		t.Fatalf("expected open buffer value, got %v", got)
	}

	changeReq, err := jsonrpc2.NewCall(jsonrpc2.NewNumberID(3), protocol.MethodTextDocumentDidChange, &protocol.DidChangeTextDocumentParams{
		TextDocument: protocol.VersionedTextDocumentIdentifier{
			TextDocumentIdentifier: protocol.TextDocumentIdentifier{URI: protocol.DocumentURI(targetURI)},
			Version:                2,
		},
		ContentChanges: []protocol.TextDocumentContentChangeEvent{
			{Text: "id: user-1\nname: Changed Alice\n"},
		},
	})
	if err != nil {
		t.Fatalf("NewCall(didChange): %v", err)
	}
	if err := server.Handle(context.Background(), captureReply[struct{}](t, nil), changeReq); err != nil {
		t.Fatalf("Handle(didChange): %v", err)
	}
	if err := server.runtime.FlushReload(); err != nil {
		t.Fatalf("FlushReload(change): %v", err)
	}

	runtimeRoot = server.runtime.RootByPath(targetPath)
	if got := runtimeRoot.Workspace.Find("User", "user-1")[0].Fields["name"]; got != "Changed Alice" {
		t.Fatalf("expected changed buffer value, got %v", got)
	}

	closeReq, err := jsonrpc2.NewCall(jsonrpc2.NewNumberID(4), protocol.MethodTextDocumentDidClose, &protocol.DidCloseTextDocumentParams{
		TextDocument: protocol.TextDocumentIdentifier{URI: protocol.DocumentURI(targetURI)},
	})
	if err != nil {
		t.Fatalf("NewCall(didClose): %v", err)
	}
	if err := server.Handle(context.Background(), captureReply[struct{}](t, nil), closeReq); err != nil {
		t.Fatalf("Handle(didClose): %v", err)
	}
	if err := server.runtime.FlushReload(); err != nil {
		t.Fatalf("FlushReload(close): %v", err)
	}

	runtimeRoot = server.runtime.RootByPath(targetPath)
	if got := runtimeRoot.Workspace.Find("User", "user-1")[0].Fields["name"]; got != "Alice" {
		t.Fatalf("expected disk fallback after close, got %v", got)
	}
}

func TestHandleDidChangePartialDocumentsAreRecoverable(t *testing.T) {
	server := NewServer(Options{Logger: testLogger()})
	root := filepath.Join("..", "workspace", "testdata", "phase4", "unknown-reference")
	rootURI := uri.File(root)
	targetPath := filepath.Join(root, "data", "posts", "post.yaml")
	targetURI := uri.File(targetPath)

	initReq, err := jsonrpc2.NewCall(jsonrpc2.NewNumberID(1), protocol.MethodInitialize, map[string]any{
		"rootUri": string(rootURI),
	})
	if err != nil {
		t.Fatalf("NewCall(initialize): %v", err)
	}
	if err := server.Handle(context.Background(), captureReply[protocol.InitializeResult](t, nil), initReq); err != nil {
		t.Fatalf("Handle(initialize): %v", err)
	}

	openReq, err := jsonrpc2.NewCall(jsonrpc2.NewNumberID(2), protocol.MethodTextDocumentDidOpen, &protocol.DidOpenTextDocumentParams{
		TextDocument: protocol.TextDocumentItem{
			URI:        protocol.DocumentURI(targetURI),
			LanguageID: "yaml",
			Version:    1,
			Text:       "id: post-1\nauthor: missing-user\ntitle: Missing Author\n",
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
			{Text: "id: post-1\nauthor: [\n"},
		},
	})
	if err != nil {
		t.Fatalf("NewCall(didChange): %v", err)
	}
	if err := server.Handle(context.Background(), captureReply[struct{}](t, nil), changeReq); err != nil {
		t.Fatalf("Handle(didChange): %v", err)
	}
	if err := server.runtime.FlushReload(); err != nil {
		t.Fatalf("FlushReload(partial): %v", err)
	}

	runtimeRoot := server.runtime.RootByPath(targetPath)
	if runtimeRoot == nil {
		t.Fatalf("expected runtime root for partial document")
	}
	if runtimeRoot.Workspace != nil {
		t.Fatalf("expected invalid partial document to drop current workspace snapshot")
	}
	if runtimeRoot.LoadErr == nil {
		t.Fatalf("expected recoverable load error for partial document")
	}
}

func TestHandleDidChangePartialJSONDocumentsAreRecoverable(t *testing.T) {
	server := NewServer(Options{Logger: testLogger()})
	root := filepath.Join("..", "..", "examples", "jsonpath")
	rootURI := uri.File(root)
	targetPath := filepath.Join(root, "data", "users.json")
	targetURI := uri.File(targetPath)

	initReq, err := jsonrpc2.NewCall(jsonrpc2.NewNumberID(1), protocol.MethodInitialize, map[string]any{
		"rootUri": string(rootURI),
	})
	if err != nil {
		t.Fatalf("NewCall(initialize): %v", err)
	}
	if err := server.Handle(context.Background(), captureReply[protocol.InitializeResult](t, nil), initReq); err != nil {
		t.Fatalf("Handle(initialize): %v", err)
	}

	openReq, err := jsonrpc2.NewCall(jsonrpc2.NewNumberID(2), protocol.MethodTextDocumentDidOpen, &protocol.DidOpenTextDocumentParams{
		TextDocument: protocol.TextDocumentItem{
			URI:        protocol.DocumentURI(targetURI),
			LanguageID: "json",
			Version:    1,
			Text:       "{\n  \"users\": [\n    {\n      \"id\": \"User-001\",\n      \"name\": \"Ada\"\n    }\n  ]\n}\n",
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
			{Text: "{\n  \"users\": [\n"},
		},
	})
	if err != nil {
		t.Fatalf("NewCall(didChange): %v", err)
	}
	if err := server.Handle(context.Background(), captureReply[struct{}](t, nil), changeReq); err != nil {
		t.Fatalf("Handle(didChange): %v", err)
	}
	if err := server.runtime.FlushReload(); err != nil {
		t.Fatalf("FlushReload(partial json): %v", err)
	}

	runtimeRoot := server.runtime.RootByPath(targetPath)
	if runtimeRoot == nil {
		t.Fatalf("expected runtime root for partial json document")
	}
	if runtimeRoot.Workspace != nil {
		t.Fatalf("expected invalid partial json document to drop current workspace snapshot")
	}
	if runtimeRoot.LoadErr == nil {
		t.Fatalf("expected recoverable load error for partial json document")
	}
}

func TestRunExitBeforeShutdownReturnsNonZero(t *testing.T) {
	clientConn, serverConn := net.Pipe()
	mustCleanupClose(t, clientConn)

	done := make(chan serveResult, 1)
	go func() {
		code, err := Run(context.Background(), serverConn, Options{Logger: testLogger()})
		done <- serveResult{code: code, err: err}
	}()

	client := jsonrpc2.NewConn(jsonrpc2.NewStream(clientConn))
	client.Go(context.Background(), protocol.Handlers(jsonrpc2.MethodNotFoundHandler))
	mustCleanupClose(t, client)

	if err := client.Notify(context.Background(), protocol.MethodExit, nil); err != nil {
		t.Fatalf("exit notify: %v", err)
	}

	res := waitServeResult(t, done)
	if res.err != nil {
		t.Fatalf("Run returned error: %v", res.err)
	}
	if res.code != 1 {
		t.Fatalf("expected non-zero exit without shutdown, got %d", res.code)
	}
}

func TestRunRejectsRequestsBeforeInitialize(t *testing.T) {
	clientConn, serverConn := net.Pipe()
	mustCleanupClose(t, clientConn)

	done := make(chan serveResult, 1)
	go func() {
		code, err := Run(context.Background(), serverConn, Options{Logger: testLogger()})
		done <- serveResult{code: code, err: err}
	}()

	client := jsonrpc2.NewConn(jsonrpc2.NewStream(clientConn))
	client.Go(context.Background(), protocol.Handlers(jsonrpc2.MethodNotFoundHandler))
	mustCleanupClose(t, client)

	if _, err := client.Call(context.Background(), protocol.MethodShutdown, nil, nil); err == nil || !strings.Contains(err.Error(), "server not initialized") {
		t.Fatalf("expected server-not-initialized error, got %v", err)
	}

	if err := client.Notify(context.Background(), protocol.MethodExit, nil); err != nil {
		t.Fatalf("exit notify: %v", err)
	}

	res := waitServeResult(t, done)
	if res.err != nil {
		t.Fatalf("Run returned error: %v", res.err)
	}
	if res.code != 1 {
		t.Fatalf("expected exit code 1, got %d", res.code)
	}
}

func TestRunStdioTranscriptUsesProtocolFramesOnly(t *testing.T) {
	clientConn, serverConn := net.Pipe()
	mustCleanupClose(t, clientConn)

	done := make(chan serveResult, 1)
	go func() {
		code, err := Run(context.Background(), serverConn, Options{Logger: testLogger()})
		done <- serveResult{code: code, err: err}
	}()

	reader := bufio.NewReader(clientConn)
	root := uri.File(t.TempDir())

	initialize := map[string]any{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  protocol.MethodInitialize,
		"params": map[string]any{
			"rootUri": string(root),
		},
	}
	if err := writeLSPFrame(clientConn, initialize); err != nil {
		t.Fatalf("write initialize: %v", err)
	}

	transcript, body := readLSPFrame(t, reader)
	if !strings.HasPrefix(transcript, "Content-Length: ") {
		t.Fatalf("expected protocol framing, got %q", transcript)
	}
	if strings.Contains(transcript, "mergeway-lsp:") {
		t.Fatalf("expected stdout to contain protocol bytes only, got %q", transcript)
	}

	var response struct {
		ID     int                       `json:"id"`
		Result protocol.InitializeResult `json:"result"`
	}
	if err := json.Unmarshal(body, &response); err != nil {
		t.Fatalf("decode initialize response: %v\nbody:\n%s", err, string(body))
	}
	if response.ID != 1 {
		t.Fatalf("expected initialize response id 1, got %d", response.ID)
	}

	if err := writeLSPFrame(clientConn, map[string]any{
		"jsonrpc": "2.0",
		"id":      2,
		"method":  protocol.MethodShutdown,
	}); err != nil {
		t.Fatalf("write shutdown: %v", err)
	}
	_, _ = readLSPFrame(t, reader)
	if err := writeLSPFrame(clientConn, map[string]any{
		"jsonrpc": "2.0",
		"method":  protocol.MethodExit,
	}); err != nil {
		t.Fatalf("write exit: %v", err)
	}

	res := waitServeResult(t, done)
	if res.err != nil {
		t.Fatalf("Run returned error: %v", res.err)
	}
	if res.code != 0 {
		t.Fatalf("expected exit code 0, got %d", res.code)
	}
}

type serveResult struct {
	code int
	err  error
}

func waitServeResult(t *testing.T, ch <-chan serveResult) serveResult {
	t.Helper()
	select {
	case res := <-ch:
		return res
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for server to stop")
		return serveResult{}
	}
}

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func captureReply[T any](t *testing.T, target *T) jsonrpc2.Replier {
	t.Helper()
	return func(_ context.Context, result interface{}, err error) error {
		if err != nil {
			return err
		}
		if target == nil || result == nil {
			return nil
		}

		body, marshalErr := json.Marshal(result)
		if marshalErr != nil {
			t.Fatalf("marshal reply: %v", marshalErr)
		}
		if unmarshalErr := json.Unmarshal(body, target); unmarshalErr != nil {
			t.Fatalf("unmarshal reply: %v", unmarshalErr)
		}
		return nil
	}
}

func mustCleanupClose(t *testing.T, closer io.Closer) {
	t.Helper()
	t.Cleanup(func() {
		if err := closer.Close(); err != nil && !isExpectedCloseError(err) {
			t.Errorf("close: %v", err)
		}
	})
}

func isExpectedCloseError(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "closed pipe") || strings.Contains(msg, "use of closed network connection")
}

func writeLSPFrame(w io.Writer, payload interface{}) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "Content-Length: %d\r\n\r\n", len(body)); err != nil {
		return err
	}
	if _, err := w.Write(body); err != nil {
		return err
	}
	return nil
}

func readLSPFrame(t *testing.T, r *bufio.Reader) (string, []byte) {
	t.Helper()

	var transcript bytes.Buffer
	contentLength := -1

	for {
		line, err := r.ReadString('\n')
		if err != nil {
			t.Fatalf("read header: %v", err)
		}
		transcript.WriteString(line)
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			break
		}
		if strings.HasPrefix(trimmed, "Content-Length:") {
			value := strings.TrimSpace(strings.TrimPrefix(trimmed, "Content-Length:"))
			n, err := strconv.Atoi(value)
			if err != nil {
				t.Fatalf("parse Content-Length %q: %v", value, err)
			}
			contentLength = n
		}
	}

	if contentLength <= 0 {
		t.Fatalf("missing Content-Length header in %q", transcript.String())
	}

	body := make([]byte, contentLength)
	if _, err := io.ReadFull(r, body); err != nil {
		t.Fatalf("read body: %v", err)
	}
	transcript.Write(body)
	return transcript.String(), body
}
