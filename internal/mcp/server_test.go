package mcp

import (
	"context"
	"encoding/json"
	"errors"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"

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

func TestRunSupportsStdioAndRejectsHTTPForNow(t *testing.T) {
	service, err := NewService(filepath.Join("..", "data", "testdata", "repo"), nil)
	if err != nil {
		t.Fatalf("NewService: %v", err)
	}

	if err := Run(context.Background(), RunOptions{Service: service, Transport: "http"}); err == nil || !strings.Contains(err.Error(), "http transport not implemented yet") {
		t.Fatalf("expected http unimplemented error, got %v", err)
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
