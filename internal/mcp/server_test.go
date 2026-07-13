package mcp

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"
	"time"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestNewServerRegistersInitialReadOnlyTools(t *testing.T) {
	service, err := NewService(filepath.Join("..", "data", "testdata", "repo"), nil)
	if err != nil {
		t.Fatalf("NewService: %v", err)
	}

	session := connectServer(t, NewServer(service))
	defer func() {
		if err := session.Close(); err != nil {
			t.Errorf("session.Close: %v", err)
		}
	}()

	var got []string
	for tool, err := range session.Tools(context.Background(), nil) {
		if err != nil {
			t.Fatalf("Tools iterator: %v", err)
		}
		got = append(got, tool.Name)
	}
	sort.Strings(got)

	want := []string{
		ToolEntityList,
		ToolEntityShow,
		ToolFilesList,
		ToolObjectGet,
		ToolObjectList,
		ToolRepositoryExport,
	}
	sort.Strings(want)
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("expected tools %v, got %v", want, got)
	}
}

func TestServerEntityListAndObjectGet(t *testing.T) {
	service, err := NewService(filepath.Join("..", "data", "testdata", "repo"), nil)
	if err != nil {
		t.Fatalf("NewService: %v", err)
	}

	session := connectServer(t, NewServer(service))
	defer func() {
		if err := session.Close(); err != nil {
			t.Errorf("session.Close: %v", err)
		}
	}()

	entityRes, err := session.CallTool(context.Background(), &sdkmcp.CallToolParams{Name: ToolEntityList})
	if err != nil {
		t.Fatalf("CallTool entity_list: %v", err)
	}

	var listOut entityListOutput
	decodeResultJSON(t, entityRes, &listOut)
	if !reflect.DeepEqual(listOut.Entities, []string{"Post", "Tag", "User"}) {
		t.Fatalf("expected visible entities, got %v", listOut.Entities)
	}

	getRes, err := session.CallTool(context.Background(), &sdkmcp.CallToolParams{
		Name:      ToolObjectGet,
		Arguments: map[string]any{"entity": "User", "id": "User-Alice"},
	})
	if err != nil {
		t.Fatalf("CallTool object_get: %v", err)
	}

	var getOut objectGetOutput
	decodeResultJSON(t, getRes, &getOut)
	if getOut.Object.Type != "User" || getOut.Object.ID != "User-Alice" {
		t.Fatalf("unexpected object result: %+v", getOut.Object)
	}
	if getOut.Object.Fields["name"] != "Alice Example" {
		t.Fatalf("expected Alice fields, got %+v", getOut.Object.Fields)
	}
}

func TestServerObjectListUsesExactEntitySemantics(t *testing.T) {
	service, err := NewService(inheritanceRepo(t), nil)
	if err != nil {
		t.Fatalf("NewService: %v", err)
	}

	session := connectServer(t, NewServer(service))
	defer func() {
		if err := session.Close(); err != nil {
			t.Errorf("session.Close: %v", err)
		}
	}()

	res, err := session.CallTool(context.Background(), &sdkmcp.CallToolParams{
		Name:      ToolObjectList,
		Arguments: map[string]any{"entity": "Animal"},
	})
	if err != nil {
		t.Fatalf("CallTool object_list: %v", err)
	}

	var out objectListOutput
	decodeResultJSON(t, res, &out)
	want := []objectSummary{{
		Type: "Animal",
		ID:   "animal-1",
		File: filepath.Join(service.Root(), "data", "animals", "animal.yaml"),
	}}
	if !reflect.DeepEqual(out.Objects, want) {
		t.Fatalf("expected exact Animal objects, got %+v", out.Objects)
	}
}

func TestServerReturnsProtocolErrorsForUnknownEntityAndBlockedEntity(t *testing.T) {
	service, err := NewService(inheritanceRepo(t), []string{"Animal"})
	if err != nil {
		t.Fatalf("NewService: %v", err)
	}

	session := connectServer(t, NewServer(service))
	defer func() {
		if err := session.Close(); err != nil {
			t.Errorf("session.Close: %v", err)
		}
	}()

	_, err = session.CallTool(context.Background(), &sdkmcp.CallToolParams{
		Name:      ToolEntityShow,
		Arguments: map[string]any{"entity": "Missing"},
	})
	if err == nil || !strings.Contains(err.Error(), "unknown entity") {
		t.Fatalf("expected unknown entity protocol error, got %v", err)
	}

	_, err = session.CallTool(context.Background(), &sdkmcp.CallToolParams{
		Name:      ToolObjectGet,
		Arguments: map[string]any{"entity": "Dog", "id": "dog-1"},
	})
	if err == nil || !strings.Contains(err.Error(), "entity not allowed") {
		t.Fatalf("expected blocked entity protocol error, got %v", err)
	}
}

func TestServerRepositoryExportAndFilesList(t *testing.T) {
	service, err := NewService(filepath.Join("..", "..", "examples", "external-root-path", "primary"), []string{"Product"})
	if err != nil {
		t.Fatalf("NewService: %v", err)
	}

	session := connectServer(t, NewServer(service))
	defer func() {
		if err := session.Close(); err != nil {
			t.Errorf("session.Close: %v", err)
		}
	}()

	exportRes, err := session.CallTool(context.Background(), &sdkmcp.CallToolParams{Name: ToolRepositoryExport})
	if err != nil {
		t.Fatalf("CallTool repository_export: %v", err)
	}
	var exportOut repositoryExportOutput
	decodeResultJSON(t, exportRes, &exportOut)
	if len(exportOut.Entities["Product"]) != 2 {
		t.Fatalf("expected exported products, got %+v", exportOut.Entities)
	}

	filesRes, err := session.CallTool(context.Background(), &sdkmcp.CallToolParams{
		Name:      ToolFilesList,
		Arguments: map[string]any{"entity": "Product"},
	})
	if err != nil {
		t.Fatalf("CallTool files_list: %v", err)
	}
	var filesOut filesListOutput
	decodeResultJSON(t, filesRes, &filesOut)
	wantFiles := []FileEntry{
		{Type: "Product", File: "../secondary/products/gadget.yaml"},
		{Type: "Product", File: "../secondary/products/widget.yaml"},
	}
	if !reflect.DeepEqual(filesOut.Files, wantFiles) {
		t.Fatalf("expected files %v, got %v", wantFiles, filesOut.Files)
	}
}

func TestNewHTTPHandlerSupportsMountedBasePath(t *testing.T) {
	service, err := NewService(filepath.Join("..", "data", "testdata", "repo"), nil)
	if err != nil {
		t.Fatalf("NewService: %v", err)
	}

	handler, err := NewHTTPHandler(service, HTTPHandlerOptions{BasePath: "/mcp/"})
	if err != nil {
		t.Fatalf("NewHTTPHandler: %v", err)
	}
	server := httptest.NewServer(handler)
	defer server.Close()

	client := sdkmcp.NewClient(&sdkmcp.Implementation{Name: "mergeway-mcp-http-test-client"}, nil)

	session, err := client.Connect(context.Background(), &sdkmcp.StreamableClientTransport{Endpoint: server.URL + "/mcp"}, nil)
	if err != nil {
		t.Fatalf("client.Connect /mcp: %v", err)
	}
	defer func() {
		if err := session.Close(); err != nil {
			t.Errorf("session.Close: %v", err)
		}
	}()

	res, err := session.CallTool(context.Background(), &sdkmcp.CallToolParams{Name: ToolEntityList})
	if err != nil {
		t.Fatalf("CallTool entity_list: %v", err)
	}
	var out entityListOutput
	decodeResultJSON(t, res, &out)
	if !reflect.DeepEqual(out.Entities, []string{"Post", "Tag", "User"}) {
		t.Fatalf("expected visible entities, got %v", out.Entities)
	}
}

func TestNewHTTPHandlerRejectsRequestsOutsideConfiguredMount(t *testing.T) {
	service, err := NewService(filepath.Join("..", "data", "testdata", "repo"), nil)
	if err != nil {
		t.Fatalf("NewService: %v", err)
	}

	handler, err := NewHTTPHandler(service, HTTPHandlerOptions{BasePath: "/mcp"})
	if err != nil {
		t.Fatalf("NewHTTPHandler: %v", err)
	}
	server := httptest.NewServer(handler)
	defer server.Close()

	req, err := http.NewRequest(http.MethodPost, server.URL+"/wrong", bytes.NewBufferString(`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}`))
	if err != nil {
		t.Fatalf("http.NewRequest: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Accept", "text/event-stream")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("http.Do: %v", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			t.Errorf("resp.Body.Close: %v", err)
		}
	}()
	if resp.StatusCode != http.StatusNotFound {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected 404 outside mount, got %d body=%s", resp.StatusCode, string(body))
	}
}

func TestRunSupportsStdioAndHTTP(t *testing.T) {
	service, err := NewService(filepath.Join("..", "data", "testdata", "repo"), nil)
	if err != nil {
		t.Fatalf("NewService: %v", err)
	}

	err = Run(context.Background(), RunOptions{
		Service:   service,
		Transport: "stdio",
		Stdin:     strings.NewReader(""),
		Stdout:    discardWriter{},
	})
	if err != nil && !errors.Is(err, context.Canceled) && !strings.Contains(err.Error(), "EOF") {
		t.Fatalf("expected stdio run to terminate cleanly on EOF, got %v", err)
	}

	addr := freeTCPAddress(t)
	ctx, cancel := context.WithCancel(context.Background())
	errCh := make(chan error, 1)
	go func() {
		errCh <- Run(ctx, RunOptions{
			Service:      service,
			Transport:    "http",
			HTTPListen:   addr,
			HTTPBasePath: "/mcp",
		})
	}()

	waitForHTTPReady(t, "http://"+addr+"/mcp")
	cancel()

	select {
	case err := <-errCh:
		if err != nil {
			t.Fatalf("Run http: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for http Run to return")
	}
}

func connectServer(t *testing.T, server *sdkmcp.Server) *sdkmcp.ClientSession {
	t.Helper()

	ctx := context.Background()
	serverTransport, clientTransport := sdkmcp.NewInMemoryTransports()
	serverSession, err := server.Connect(ctx, serverTransport, nil)
	if err != nil {
		t.Fatalf("server.Connect: %v", err)
	}
	t.Cleanup(func() {
		if err := serverSession.Close(); err != nil {
			t.Errorf("serverSession.Close: %v", err)
		}
		if err := serverSession.Wait(); err != nil {
			t.Errorf("serverSession.Wait: %v", err)
		}
	})

	client := sdkmcp.NewClient(&sdkmcp.Implementation{Name: "mergeway-mcp-test-client"}, nil)
	session, err := client.Connect(ctx, clientTransport, nil)
	if err != nil {
		t.Fatalf("client.Connect: %v", err)
	}
	return session
}

func decodeResultJSON(t *testing.T, res *sdkmcp.CallToolResult, target any) {
	t.Helper()

	if res == nil {
		t.Fatal("expected non-nil call tool result")
	}
	if res.IsError {
		t.Fatalf("expected non-error result, got %+v", res)
	}
	if len(res.Content) == 0 {
		t.Fatalf("expected text content, got %+v", res)
	}
	text, ok := res.Content[0].(*sdkmcp.TextContent)
	if !ok {
		t.Fatalf("expected text content, got %T", res.Content[0])
	}
	if err := json.Unmarshal([]byte(text.Text), target); err != nil {
		t.Fatalf("unmarshal result JSON: %v\ntext=%s", err, text.Text)
	}
}

type discardWriter struct{}

func (discardWriter) Write(p []byte) (int, error) { return len(p), nil }

func freeTCPAddress(t *testing.T) string {
	t.Helper()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("net.Listen: %v", err)
	}
	addr := listener.Addr().String()
	if err := listener.Close(); err != nil {
		t.Fatalf("listener.Close: %v", err)
	}
	return addr
}

func waitForHTTPReady(t *testing.T, endpoint string) {
	t.Helper()

	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		req, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewBufferString(`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}`))
		if err != nil {
			t.Fatalf("http.NewRequest: %v", err)
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Add("Accept", "application/json")
		req.Header.Add("Accept", "text/event-stream")

		resp, err := http.DefaultClient.Do(req)
		if err == nil {
			_ = resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				return
			}
		}
		time.Sleep(25 * time.Millisecond)
	}
	t.Fatalf("http endpoint %s did not become ready", endpoint)
}
