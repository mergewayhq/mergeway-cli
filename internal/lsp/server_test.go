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
