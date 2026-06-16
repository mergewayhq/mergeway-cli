package lsp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"sync"

	"github.com/mergewayhq/mergeway-cli/internal/version"
	"go.lsp.dev/jsonrpc2"
	"go.lsp.dev/protocol"
	"go.lsp.dev/uri"
)

// Options configures the LSP transport service.
type Options struct {
	Logger *slog.Logger
}

// Server handles the minimal LSP lifecycle required for phase 3.
type Server struct {
	logger *slog.Logger

	mu                sync.Mutex
	initialized       bool
	shutdownRequested bool
	exitCode          int
	trace             protocol.TraceValue
	rootURI           uri.URI
	workspaceFolders  []uri.URI
}

// Run serves LSP traffic over a stdio-compatible connection until the client
// sends exit or the transport fails.
func Run(ctx context.Context, conn io.ReadWriteCloser, opts Options) (int, error) {
	if conn == nil {
		return 1, errors.New("lsp: connection is required")
	}

	server := NewServer(opts)
	stream := jsonrpc2.NewStream(conn)
	handler := protocol.CancelHandler(jsonrpc2.ReplyHandler(server.Handle))

	for {
		msg, _, err := stream.Read(ctx)
		if err != nil {
			switch {
			case errors.Is(err, io.EOF), isClosedConnError(err), errors.Is(err, context.Canceled):
				return server.ExitCode(), nil
			default:
				return server.ExitCode(), err
			}
		}

		req, ok := msg.(jsonrpc2.Request)
		if !ok {
			continue
		}

		err = handler(ctx, replier(stream, req), req)
		switch {
		case err == nil:
			continue
		case errors.Is(err, errExitRequested):
			return server.ExitCode(), nil
		default:
			return server.ExitCode(), err
		}
	}
}

// NewServer constructs a minimal mergeway LSP lifecycle server.
func NewServer(opts Options) *Server {
	logger := opts.Logger
	if logger == nil {
		logger = slog.New(slog.NewTextHandler(io.Discard, nil))
	}

	return &Server{
		logger:   logger,
		exitCode: 1,
		trace:    protocol.TraceOff,
	}
}

// ExitCode reports the exit code selected by the server lifecycle so far.
func (s *Server) ExitCode() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.exitCode
}

// Handle dispatches the minimal lifecycle methods needed for the initial LSP surface.
func (s *Server) Handle(ctx context.Context, reply jsonrpc2.Replier, req jsonrpc2.Request) error {
	switch req.Method() {
	case protocol.MethodInitialize:
		return s.handleInitialize(ctx, reply, req)
	case protocol.MethodInitialized:
		return s.handleInitialized(ctx, reply, req)
	case protocol.MethodShutdown:
		return s.handleShutdown(ctx, reply, req)
	case protocol.MethodExit:
		return s.handleExit(ctx, reply, req)
	case protocol.MethodSetTrace:
		return s.handleSetTrace(ctx, reply, req)
	default:
		if !s.isInitialized() {
			return reply(ctx, nil, jsonrpc2.NewError(jsonrpc2.ServerNotInitialized, "server not initialized"))
		}
		if s.isShuttingDown() {
			return reply(ctx, nil, jsonrpc2.NewError(jsonrpc2.InvalidRequest, "server is shutting down"))
		}
		return reply(ctx, nil, fmt.Errorf("%q: %w", req.Method(), jsonrpc2.ErrMethodNotFound))
	}
}

func (s *Server) handleInitialize(ctx context.Context, reply jsonrpc2.Replier, req jsonrpc2.Request) error {
	var params protocol.InitializeParams
	if err := decodeParams(req.Params(), &params); err != nil {
		return reply(ctx, nil, fmt.Errorf("%s: %w", jsonrpc2.ErrParse, err))
	}
	legacy, err := decodeLegacyInitialize(req.Params())
	if err != nil {
		return reply(ctx, nil, fmt.Errorf("%s: %w", jsonrpc2.ErrParse, err))
	}

	s.mu.Lock()
	if s.initialized {
		s.mu.Unlock()
		return reply(ctx, nil, jsonrpc2.NewError(jsonrpc2.InvalidRequest, "server already initialized"))
	}

	s.trace = params.Trace
	s.rootURI = resolveRootURI(&params, legacy)
	s.workspaceFolders = resolveWorkspaceFolders(&params)
	s.initialized = true
	s.shutdownRequested = false
	s.exitCode = 1
	s.mu.Unlock()

	s.logger.Debug("initialize",
		slog.String("root_uri", string(s.rootURI)),
		slog.Int("workspace_folders", len(s.workspaceFolders)),
		slog.String("trace", string(s.trace)),
	)

	result := &protocol.InitializeResult{
		Capabilities: protocol.ServerCapabilities{},
		ServerInfo: &protocol.ServerInfo{
			Name:    "mergeway-lsp",
			Version: version.Number,
		},
	}
	return reply(ctx, result, nil)
}

func (s *Server) handleInitialized(ctx context.Context, reply jsonrpc2.Replier, req jsonrpc2.Request) error {
	var params protocol.InitializedParams
	if err := decodeParams(req.Params(), &params); err != nil {
		return reply(ctx, nil, fmt.Errorf("%s: %w", jsonrpc2.ErrParse, err))
	}
	s.logger.Debug("initialized")
	return reply(ctx, nil, nil)
}

func (s *Server) handleShutdown(ctx context.Context, reply jsonrpc2.Replier, req jsonrpc2.Request) error {
	if len(req.Params()) > 0 {
		return reply(ctx, nil, fmt.Errorf("expected no params: %w", jsonrpc2.ErrInvalidParams))
	}
	if !s.isInitialized() {
		return reply(ctx, nil, jsonrpc2.NewError(jsonrpc2.ServerNotInitialized, "server not initialized"))
	}

	s.mu.Lock()
	s.shutdownRequested = true
	s.mu.Unlock()

	s.logger.Debug("shutdown")
	return reply(ctx, nil, nil)
}

func (s *Server) handleExit(ctx context.Context, reply jsonrpc2.Replier, req jsonrpc2.Request) error {
	if len(req.Params()) > 0 {
		if err := reply(ctx, nil, fmt.Errorf("expected no params: %w", jsonrpc2.ErrInvalidParams)); err != nil {
			return err
		}
		return errExitRequested
	}

	s.mu.Lock()
	if s.shutdownRequested {
		s.exitCode = 0
	} else {
		s.exitCode = 1
	}
	s.mu.Unlock()

	s.logger.Debug("exit", slog.Int("exit_code", s.ExitCode()))
	if err := reply(ctx, nil, nil); err != nil {
		return err
	}
	return errExitRequested
}

func (s *Server) handleSetTrace(ctx context.Context, reply jsonrpc2.Replier, req jsonrpc2.Request) error {
	var params protocol.SetTraceParams
	if err := decodeParams(req.Params(), &params); err != nil {
		return reply(ctx, nil, fmt.Errorf("%s: %w", jsonrpc2.ErrParse, err))
	}

	s.mu.Lock()
	s.trace = params.Value
	s.mu.Unlock()

	s.logger.Debug("set_trace", slog.String("trace", string(params.Value)))
	return reply(ctx, nil, nil)
}

func (s *Server) isInitialized() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.initialized
}

func (s *Server) isShuttingDown() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.shutdownRequested
}

func resolveRootURI(params *protocol.InitializeParams, legacy legacyInitializeParams) uri.URI {
	if params == nil {
		return ""
	}
	if len(params.WorkspaceFolders) > 0 {
		return uri.URI(params.WorkspaceFolders[0].URI)
	}
	if legacy.RootURI != "" {
		return uri.URI(legacy.RootURI)
	}
	if legacy.RootPath != "" {
		return uri.File(legacy.RootPath)
	}
	return ""
}

func resolveWorkspaceFolders(params *protocol.InitializeParams) []uri.URI {
	if params == nil || len(params.WorkspaceFolders) == 0 {
		return nil
	}

	folders := make([]uri.URI, 0, len(params.WorkspaceFolders))
	for _, folder := range params.WorkspaceFolders {
		if folder.URI == "" {
			continue
		}
		folders = append(folders, uri.URI(folder.URI))
	}
	return folders
}

func decodeParams(raw json.RawMessage, dst interface{}) error {
	if len(raw) == 0 {
		return nil
	}
	if err := json.Unmarshal(raw, dst); err != nil {
		return err
	}
	return nil
}

type legacyInitializeParams struct {
	RootURI  protocol.DocumentURI `json:"rootUri,omitempty"`
	RootPath string               `json:"rootPath,omitempty"`
}

func decodeLegacyInitialize(raw json.RawMessage) (legacyInitializeParams, error) {
	var params legacyInitializeParams
	if len(raw) == 0 {
		return params, nil
	}
	if err := json.Unmarshal(raw, &params); err != nil {
		return legacyInitializeParams{}, err
	}
	return params, nil
}

func replier(stream jsonrpc2.Stream, req jsonrpc2.Request) jsonrpc2.Replier {
	return func(ctx context.Context, result interface{}, err error) error {
		call, ok := req.(*jsonrpc2.Call)
		if !ok {
			return nil
		}

		response, responseErr := jsonrpc2.NewResponse(call.ID(), result, err)
		if responseErr != nil {
			return responseErr
		}
		_, responseErr = stream.Write(ctx, response)
		return responseErr
	}
}

func isClosedConnError(err error) bool {
	if err == nil {
		return false
	}
	message := err.Error()
	return message == "io: read/write on closed pipe" ||
		message == "read/write on closed pipe" ||
		message == "EOF"
}

var errExitRequested = errors.New("lsp: exit requested")
