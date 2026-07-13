package mcp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"path"
	"strings"
	"time"

	sdkjsonrpc "github.com/modelcontextprotocol/go-sdk/jsonrpc"
	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/mergewayhq/mergeway-cli/internal/version"
)

const (
	ToolEntityList       = "entity_list"
	ToolEntityShow       = "entity_show"
	ToolObjectList       = "object_list"
	ToolObjectGet        = "object_get"
	ToolRepositoryExport = "repository_export"
	ToolFilesList        = "files_list"
)

// RunOptions configures server startup for the currently supported transports.
type RunOptions struct {
	Service      *Service
	Transport    string
	Stdin        io.Reader
	Stdout       io.Writer
	HTTPListen   string
	HTTPBasePath string
}

// NewServer constructs an MCP server exposing the initial read-only Mergeway tools.
func NewServer(service *Service) *sdkmcp.Server {
	if service == nil {
		panic("mcp: service is required")
	}

	server := sdkmcp.NewServer(&sdkmcp.Implementation{
		Name:    "mergeway-mcp",
		Version: version.Number,
	}, &sdkmcp.ServerOptions{
		Instructions: "Read-only Mergeway repository inspection server. Use the provided tools to inspect entities, objects, exports, and backing files. Modification is not supported.",
	})

	sdkmcp.AddTool(server, &sdkmcp.Tool{
		Name:        ToolEntityList,
		Description: "List visible Mergeway entities after allow-list filtering.",
	}, func(ctx context.Context, req *sdkmcp.CallToolRequest, in struct{}) (*sdkmcp.CallToolResult, entityListOutput, error) {
		_ = ctx
		_ = req
		entities, err := service.EntityList()
		if err != nil {
			return nil, entityListOutput{}, protocolError(err)
		}
		return nil, entityListOutput{Entities: entities}, nil
	})

	sdkmcp.AddTool(server, &sdkmcp.Tool{
		Name:        ToolEntityShow,
		Description: "Show the schema/config details for one exact Mergeway entity.",
	}, func(ctx context.Context, req *sdkmcp.CallToolRequest, in entityShowInput) (*sdkmcp.CallToolResult, entityShowOutput, error) {
		_ = ctx
		_ = req
		if err := requireEntity(in.Entity); err != nil {
			return nil, entityShowOutput{}, err
		}
		schema, err := service.EntityShow(in.Entity)
		if err != nil {
			return nil, entityShowOutput{}, protocolError(err)
		}
		return nil, entityShowOutput{Entity: in.Entity, Schema: schema}, nil
	})

	sdkmcp.AddTool(server, &sdkmcp.Tool{
		Name:        ToolObjectList,
		Description: "List objects for one exact Mergeway entity without expanding descendants.",
	}, func(ctx context.Context, req *sdkmcp.CallToolRequest, in objectListInput) (*sdkmcp.CallToolResult, objectListOutput, error) {
		_ = ctx
		_ = req
		if err := requireEntity(in.Entity); err != nil {
			return nil, objectListOutput{}, err
		}
		objects, err := service.ObjectList(in.Entity)
		if err != nil {
			return nil, objectListOutput{}, protocolError(err)
		}
		items := make([]objectSummary, len(objects))
		for i, obj := range objects {
			items[i] = objectSummary{
				Type:     obj.Type,
				ID:       obj.ID,
				File:     obj.File,
				Inline:   obj.Inline,
				ReadOnly: obj.ReadOnly,
			}
		}
		return nil, objectListOutput{Entity: in.Entity, Objects: items}, nil
	})

	sdkmcp.AddTool(server, &sdkmcp.Tool{
		Name:        ToolObjectGet,
		Description: "Get one object by exact Mergeway entity and identifier.",
	}, func(ctx context.Context, req *sdkmcp.CallToolRequest, in objectGetInput) (*sdkmcp.CallToolResult, objectGetOutput, error) {
		_ = ctx
		_ = req
		if err := requireEntity(in.Entity); err != nil {
			return nil, objectGetOutput{}, err
		}
		if err := requireIdentifier(in.ID); err != nil {
			return nil, objectGetOutput{}, err
		}
		obj, err := service.ObjectGet(in.Entity, in.ID)
		if err != nil {
			return nil, objectGetOutput{}, protocolError(err)
		}
		return nil, objectGetOutput{
			Object: objectRecord{
				Type:     obj.Type,
				ID:       obj.ID,
				File:     obj.File,
				Inline:   obj.Inline,
				ReadOnly: obj.ReadOnly,
				Fields:   cloneMap(obj.Fields),
			},
		}, nil
	})

	sdkmcp.AddTool(server, &sdkmcp.Tool{
		Name:        ToolRepositoryExport,
		Description: "Export visible Mergeway entities as a structured read-only snapshot.",
	}, func(ctx context.Context, req *sdkmcp.CallToolRequest, in repositoryExportInput) (*sdkmcp.CallToolResult, repositoryExportOutput, error) {
		_ = ctx
		_ = req
		exported, err := service.RepositoryExport(in.Entities)
		if err != nil {
			return nil, repositoryExportOutput{}, protocolError(err)
		}
		return nil, repositoryExportOutput{Entities: exported}, nil
	})

	sdkmcp.AddTool(server, &sdkmcp.Tool{
		Name:        ToolFilesList,
		Description: "List configured backing files for visible Mergeway entities.",
	}, func(ctx context.Context, req *sdkmcp.CallToolRequest, in filesListInput) (*sdkmcp.CallToolResult, filesListOutput, error) {
		_ = ctx
		_ = req
		if in.Entity != "" {
			if err := requireEntity(in.Entity); err != nil {
				return nil, filesListOutput{}, err
			}
		}
		files, err := service.FilesList(in.Entity)
		if err != nil {
			return nil, filesListOutput{}, protocolError(err)
		}
		return nil, filesListOutput{Files: files}, nil
	})

	return server
}

// HTTPHandlerOptions configures the mounted HTTP handler.
type HTTPHandlerOptions struct {
	BasePath string
}

// NewHTTPHandler constructs the single supported HTTP transport mode:
// streamable HTTP mounted at one base path in stateless JSON-response mode.
func NewHTTPHandler(service *Service, opts HTTPHandlerOptions) (http.Handler, error) {
	if service == nil {
		return nil, errors.New("mcp: service is required")
	}

	basePath, err := normalizeMountedBasePath(opts.BasePath)
	if err != nil {
		return nil, err
	}

	server := NewServer(service)
	handler := sdkmcp.NewStreamableHTTPHandler(func(*http.Request) *sdkmcp.Server {
		return server
	}, &sdkmcp.StreamableHTTPOptions{
		Stateless:    true,
		JSONResponse: true,
	})

	return mountedHTTPHandler{
		basePath: basePath,
		handler:  handler,
	}, nil
}

// Run starts the MCP server over the currently supported transport.
func Run(ctx context.Context, opts RunOptions) error {
	if opts.Service == nil {
		return errors.New("mcp: service is required")
	}

	server := NewServer(opts.Service)
	switch opts.Transport {
	case "", "stdio":
		if opts.Stdin == nil || opts.Stdout == nil {
			return errors.New("mcp: stdio transport requires stdin and stdout")
		}
		return server.Run(ctx, &sdkmcp.IOTransport{
			Reader: nopReadCloser{Reader: opts.Stdin},
			Writer: nopWriteCloser{Writer: opts.Stdout},
		})
	case "http":
		return runHTTP(ctx, opts)
	default:
		return fmt.Errorf("mcp: unsupported transport %q", opts.Transport)
	}
}

func runHTTP(ctx context.Context, opts RunOptions) error {
	handler, err := NewHTTPHandler(opts.Service, HTTPHandlerOptions{BasePath: opts.HTTPBasePath})
	if err != nil {
		return err
	}

	listenAddr := strings.TrimSpace(opts.HTTPListen)
	if listenAddr == "" {
		return errors.New("mcp: http listen address is required")
	}

	listener, err := net.Listen("tcp", listenAddr)
	if err != nil {
		return fmt.Errorf("mcp: listen %s: %w", listenAddr, err)
	}

	server := &http.Server{
		Handler: handler,
	}

	errCh := make(chan error, 1)
	go func() {
		err := server.Serve(listener)
		if errors.Is(err, http.ErrServerClosed) {
			errCh <- nil
			return
		}
		errCh <- err
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil && !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("mcp: shutdown http server: %w", err)
		}
		if err := <-errCh; err != nil {
			return err
		}
		return nil
	}
}

type entityListOutput struct {
	Entities []string `json:"entities" jsonschema:"visible exact Mergeway entities"`
}

type entityShowInput struct {
	Entity string `json:"entity" jsonschema:"exact Mergeway entity name"`
}

type entityShowOutput struct {
	Entity string `json:"entity" jsonschema:"exact Mergeway entity name"`
	Schema any    `json:"schema" jsonschema:"normalized Mergeway schema for the entity"`
}

type objectListInput struct {
	Entity string `json:"entity" jsonschema:"exact Mergeway entity name"`
}

type objectListOutput struct {
	Entity  string          `json:"entity" jsonschema:"exact Mergeway entity name"`
	Objects []objectSummary `json:"objects" jsonschema:"objects declared exactly as this entity"`
}

type objectSummary struct {
	Type     string `json:"type" jsonschema:"concrete Mergeway entity type"`
	ID       string `json:"id" jsonschema:"object identifier"`
	File     string `json:"file,omitempty" jsonschema:"backing file path, if file-backed"`
	Inline   bool   `json:"inline,omitempty" jsonschema:"whether the object is defined inline in config"`
	ReadOnly bool   `json:"readOnly,omitempty" jsonschema:"whether the object is read-only because of its source"`
}

type objectGetInput struct {
	Entity string `json:"entity" jsonschema:"exact Mergeway entity name"`
	ID     string `json:"id" jsonschema:"object identifier"`
}

type objectGetOutput struct {
	Object objectRecord `json:"object" jsonschema:"full Mergeway object record"`
}

type objectRecord struct {
	Type     string         `json:"type" jsonschema:"concrete Mergeway entity type"`
	ID       string         `json:"id" jsonschema:"object identifier"`
	File     string         `json:"file,omitempty" jsonschema:"backing file path, if file-backed"`
	Inline   bool           `json:"inline,omitempty" jsonschema:"whether the object is defined inline in config"`
	ReadOnly bool           `json:"readOnly,omitempty" jsonschema:"whether the object is read-only because of its source"`
	Fields   map[string]any `json:"fields" jsonschema:"object fields including derived read-only fields"`
}

type repositoryExportInput struct {
	Entities []string `json:"entities,omitempty" jsonschema:"optional exact Mergeway entity names to export; omit for all visible entities"`
}

type repositoryExportOutput struct {
	Entities map[string][]map[string]any `json:"entities" jsonschema:"exported entities keyed by exact entity name"`
}

type filesListInput struct {
	Entity string `json:"entity,omitempty" jsonschema:"optional exact Mergeway entity name"`
}

type filesListOutput struct {
	Files []FileEntry `json:"files" jsonschema:"configured backing files for visible entities"`
}

func requireEntity(entity string) error {
	if entity == "" {
		return newProtocolError(sdkjsonrpc.CodeInvalidParams, "invalid_arguments", "entity is required")
	}
	return nil
}

func requireIdentifier(id string) error {
	if id == "" {
		return newProtocolError(sdkjsonrpc.CodeInvalidParams, "invalid_arguments", "id is required")
	}
	return nil
}

func protocolError(err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, ErrUnknownEntity):
		return newProtocolError(sdkjsonrpc.CodeInvalidParams, "unknown_entity", err.Error())
	case errors.Is(err, ErrEntityNotAllowed):
		return newProtocolError(sdkjsonrpc.CodeInvalidParams, "entity_not_allowed", err.Error())
	case errors.Is(err, ErrObjectNotFound):
		return newProtocolError(sdkjsonrpc.CodeInvalidParams, "object_not_found", err.Error())
	default:
		return newProtocolError(sdkjsonrpc.CodeInternalError, "repository_error", err.Error())
	}
}

func newProtocolError(code int64, kind, message string) error {
	data, err := json.Marshal(map[string]any{"kind": kind})
	if err != nil {
		data = nil
	}
	return &sdkjsonrpc.Error{
		Code:    code,
		Message: message,
		Data:    data,
	}
}

type nopReadCloser struct {
	io.Reader
}

func (nopReadCloser) Close() error { return nil }

type nopWriteCloser struct {
	io.Writer
}

func (nopWriteCloser) Close() error { return nil }

type mountedHTTPHandler struct {
	basePath string
	handler  http.Handler
}

func (h mountedHTTPHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if !matchesMountedBasePath(req.URL.Path, h.basePath) {
		http.NotFound(w, req)
		return
	}
	h.handler.ServeHTTP(w, req)
}

func normalizeMountedBasePath(basePath string) (string, error) {
	trimmed := strings.TrimSpace(basePath)
	if trimmed == "" {
		trimmed = "/"
	}
	if !strings.HasPrefix(trimmed, "/") {
		return "", fmt.Errorf("mcp: invalid http base path %q", basePath)
	}

	cleaned := path.Clean(trimmed)
	if !strings.HasPrefix(cleaned, "/") {
		cleaned = "/" + cleaned
	}
	return cleaned, nil
}

func matchesMountedBasePath(requestPath, basePath string) bool {
	if basePath == "/" {
		return requestPath == "/"
	}

	requestClean := path.Clean(requestPath)
	if !strings.HasPrefix(requestClean, "/") {
		requestClean = "/" + requestClean
	}
	return requestClean == basePath
}
