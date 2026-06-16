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
	"github.com/mergewayhq/mergeway-cli/internal/workspace"
	"go.lsp.dev/jsonrpc2"
	"go.lsp.dev/protocol"
	"go.lsp.dev/uri"
)

// Options configures the LSP transport service.
type Options struct {
	Logger *slog.Logger

	// PublishDiagnostics optionally overrides outbound diagnostics publishing.
	// It is primarily intended for tests.
	PublishDiagnostics func(context.Context, *protocol.PublishDiagnosticsParams) error
}

// Server handles the minimal LSP lifecycle required for phase 3.
type Server struct {
	logger *slog.Logger

	mu                 sync.Mutex
	initialized        bool
	shutdownRequested  bool
	exitCode           int
	trace              protocol.TraceValue
	rootURI            uri.URI
	workspaceFolders   []uri.URI
	roots              *workspace.RootSet
	runtime            *workspace.Runtime
	publishDiagnostics diagnosticPublisher
	publishMu          sync.Mutex
	publishedPaths     map[string]struct{}
}

// Run serves LSP traffic over a stdio-compatible connection until the client
// sends exit or the transport fails.
func Run(ctx context.Context, conn io.ReadWriteCloser, opts Options) (int, error) {
	if conn == nil {
		return 1, errors.New("lsp: connection is required")
	}

	server := NewServer(opts)
	stream := jsonrpc2.NewStream(conn)
	writeMu := &sync.Mutex{}
	if server.publishDiagnostics == nil {
		server.publishDiagnostics = streamDiagnosticPublisher(stream, writeMu)
	}
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

		err = handler(ctx, replier(stream, writeMu, req), req)
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
		logger:             logger,
		exitCode:           1,
		trace:              protocol.TraceOff,
		publishDiagnostics: opts.PublishDiagnostics,
		publishedPaths:     make(map[string]struct{}),
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
	case protocol.MethodTextDocumentDidOpen:
		return s.handleDidOpen(ctx, reply, req)
	case protocol.MethodTextDocumentDidChange:
		return s.handleDidChange(ctx, reply, req)
	case protocol.MethodTextDocumentDidClose:
		return s.handleDidClose(ctx, reply, req)
	case protocol.MethodTextDocumentCompletion:
		return s.handleCompletion(ctx, reply, req)
	case protocol.MethodTextDocumentHover:
		return s.handleHover(ctx, reply, req)
	case protocol.MethodTextDocumentDefinition:
		return s.handleDefinition(ctx, reply, req)
	case protocol.MethodTextDocumentReferences:
		return s.handleReferences(ctx, reply, req)
	case protocol.MethodTextDocumentDocumentSymbol:
		return s.handleDocumentSymbol(ctx, reply, req)
	case protocol.MethodWorkspaceSymbol:
		return s.handleWorkspaceSymbol(ctx, reply, req)
	case protocol.MethodTextDocumentCodeAction:
		return s.handleCodeAction(ctx, reply, req)
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
	roots, err := workspace.OpenRoots(resolveRootCandidates(s.rootURI, s.workspaceFolders))
	if err != nil {
		s.mu.Unlock()
		return reply(ctx, nil, err)
	}
	s.roots = roots
	s.runtime = workspace.NewRuntime(roots)
	s.runtime.SetReloadHook(func() {
		if err := s.publishWorkspaceDiagnostics(context.Background()); err != nil {
			s.logger.Error("publish_diagnostics", slog.Any("error", err))
		}
	})
	s.initialized = true
	s.shutdownRequested = false
	s.exitCode = 1
	s.mu.Unlock()

	if err := s.runtime.FlushReload(); err != nil {
		return reply(ctx, nil, err)
	}

	s.logger.Debug("initialize",
		slog.String("root_uri", string(s.rootURI)),
		slog.Int("workspace_folders", len(s.workspaceFolders)),
		slog.Int("detected_roots", len(roots.Roots)),
		slog.Int("missing_roots", len(roots.MissingRoots)),
		slog.String("trace", string(s.trace)),
	)

	result := &protocol.InitializeResult{
		Capabilities: protocol.ServerCapabilities{
			TextDocumentSync: &protocol.TextDocumentSyncOptions{
				OpenClose: true,
				Change:    protocol.TextDocumentSyncKindFull,
			},
			CompletionProvider:      &protocol.CompletionOptions{},
			HoverProvider:           true,
			DefinitionProvider:      true,
			ReferencesProvider:      true,
			DocumentSymbolProvider:  true,
			WorkspaceSymbolProvider: true,
			CodeActionProvider:      true,
			Workspace: &protocol.ServerCapabilitiesWorkspace{
				WorkspaceFolders: &protocol.ServerCapabilitiesWorkspaceFolders{
					Supported:           true,
					ChangeNotifications: false,
				},
			},
		},
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

func (s *Server) handleDidOpen(ctx context.Context, reply jsonrpc2.Replier, req jsonrpc2.Request) error {
	var params protocol.DidOpenTextDocumentParams
	if err := decodeParams(req.Params(), &params); err != nil {
		return reply(ctx, nil, fmt.Errorf("%s: %w", jsonrpc2.ErrParse, err))
	}

	if s.runtime == nil {
		return reply(ctx, nil, jsonrpc2.NewError(jsonrpc2.ServerNotInitialized, "server not initialized"))
	}

	path := params.TextDocument.URI.Filename()
	err := s.runtime.DidOpen(&workspace.OpenDocument{
		URI:        string(params.TextDocument.URI),
		Path:       path,
		LanguageID: string(params.TextDocument.LanguageID),
		Version:    params.TextDocument.Version,
		Text:       params.TextDocument.Text,
	})
	if err != nil {
		return reply(ctx, nil, err)
	}
	return reply(ctx, nil, nil)
}

func (s *Server) handleDidChange(ctx context.Context, reply jsonrpc2.Replier, req jsonrpc2.Request) error {
	var params protocol.DidChangeTextDocumentParams
	if err := decodeParams(req.Params(), &params); err != nil {
		return reply(ctx, nil, fmt.Errorf("%s: %w", jsonrpc2.ErrParse, err))
	}

	if s.runtime == nil {
		return reply(ctx, nil, jsonrpc2.NewError(jsonrpc2.ServerNotInitialized, "server not initialized"))
	}
	if len(params.ContentChanges) == 0 {
		return reply(ctx, nil, nil)
	}

	change := params.ContentChanges[len(params.ContentChanges)-1]
	err := s.runtime.DidChange(params.TextDocument.URI.Filename(), params.TextDocument.Version, change.Text)
	if err != nil {
		return reply(ctx, nil, err)
	}
	return reply(ctx, nil, nil)
}

func (s *Server) handleDidClose(ctx context.Context, reply jsonrpc2.Replier, req jsonrpc2.Request) error {
	var params protocol.DidCloseTextDocumentParams
	if err := decodeParams(req.Params(), &params); err != nil {
		return reply(ctx, nil, fmt.Errorf("%s: %w", jsonrpc2.ErrParse, err))
	}

	if s.runtime == nil {
		return reply(ctx, nil, jsonrpc2.NewError(jsonrpc2.ServerNotInitialized, "server not initialized"))
	}

	s.runtime.DidClose(params.TextDocument.URI.Filename())
	return reply(ctx, nil, nil)
}

func (s *Server) handleCompletion(ctx context.Context, reply jsonrpc2.Replier, req jsonrpc2.Request) error {
	var params protocol.CompletionParams
	if err := decodeParams(req.Params(), &params); err != nil {
		return reply(ctx, nil, fmt.Errorf("%s: %w", jsonrpc2.ErrParse, err))
	}
	if !s.isInitialized() {
		return reply(ctx, nil, jsonrpc2.NewError(jsonrpc2.ServerNotInitialized, "server not initialized"))
	}
	if s.isShuttingDown() {
		return reply(ctx, nil, jsonrpc2.NewError(jsonrpc2.InvalidRequest, "server is shutting down"))
	}

	result, err := s.completion(ctx, &params)
	return reply(ctx, result, err)
}

func (s *Server) handleHover(ctx context.Context, reply jsonrpc2.Replier, req jsonrpc2.Request) error {
	var params protocol.HoverParams
	if err := decodeParams(req.Params(), &params); err != nil {
		return reply(ctx, nil, fmt.Errorf("%s: %w", jsonrpc2.ErrParse, err))
	}
	if !s.isInitialized() {
		return reply(ctx, nil, jsonrpc2.NewError(jsonrpc2.ServerNotInitialized, "server not initialized"))
	}
	if s.isShuttingDown() {
		return reply(ctx, nil, jsonrpc2.NewError(jsonrpc2.InvalidRequest, "server is shutting down"))
	}

	result, err := s.hover(ctx, &params)
	return reply(ctx, result, err)
}

func (s *Server) handleDefinition(ctx context.Context, reply jsonrpc2.Replier, req jsonrpc2.Request) error {
	var params protocol.DefinitionParams
	if err := decodeParams(req.Params(), &params); err != nil {
		return reply(ctx, nil, fmt.Errorf("%s: %w", jsonrpc2.ErrParse, err))
	}
	if !s.isInitialized() {
		return reply(ctx, nil, jsonrpc2.NewError(jsonrpc2.ServerNotInitialized, "server not initialized"))
	}
	if s.isShuttingDown() {
		return reply(ctx, nil, jsonrpc2.NewError(jsonrpc2.InvalidRequest, "server is shutting down"))
	}

	result, err := s.definition(ctx, &params)
	return reply(ctx, result, err)
}

func (s *Server) handleReferences(ctx context.Context, reply jsonrpc2.Replier, req jsonrpc2.Request) error {
	var params protocol.ReferenceParams
	if err := decodeParams(req.Params(), &params); err != nil {
		return reply(ctx, nil, fmt.Errorf("%s: %w", jsonrpc2.ErrParse, err))
	}
	if !s.isInitialized() {
		return reply(ctx, nil, jsonrpc2.NewError(jsonrpc2.ServerNotInitialized, "server not initialized"))
	}
	if s.isShuttingDown() {
		return reply(ctx, nil, jsonrpc2.NewError(jsonrpc2.InvalidRequest, "server is shutting down"))
	}

	result, err := s.references(ctx, &params)
	return reply(ctx, result, err)
}

func (s *Server) handleDocumentSymbol(ctx context.Context, reply jsonrpc2.Replier, req jsonrpc2.Request) error {
	var params protocol.DocumentSymbolParams
	if err := decodeParams(req.Params(), &params); err != nil {
		return reply(ctx, nil, fmt.Errorf("%s: %w", jsonrpc2.ErrParse, err))
	}
	if !s.isInitialized() {
		return reply(ctx, nil, jsonrpc2.NewError(jsonrpc2.ServerNotInitialized, "server not initialized"))
	}
	if s.isShuttingDown() {
		return reply(ctx, nil, jsonrpc2.NewError(jsonrpc2.InvalidRequest, "server is shutting down"))
	}

	result, err := s.documentSymbols(ctx, &params)
	return reply(ctx, result, err)
}

func (s *Server) handleWorkspaceSymbol(ctx context.Context, reply jsonrpc2.Replier, req jsonrpc2.Request) error {
	var params protocol.WorkspaceSymbolParams
	if err := decodeParams(req.Params(), &params); err != nil {
		return reply(ctx, nil, fmt.Errorf("%s: %w", jsonrpc2.ErrParse, err))
	}
	if !s.isInitialized() {
		return reply(ctx, nil, jsonrpc2.NewError(jsonrpc2.ServerNotInitialized, "server not initialized"))
	}
	if s.isShuttingDown() {
		return reply(ctx, nil, jsonrpc2.NewError(jsonrpc2.InvalidRequest, "server is shutting down"))
	}

	result, err := s.workspaceSymbols(ctx, &params)
	return reply(ctx, result, err)
}

func (s *Server) handleCodeAction(ctx context.Context, reply jsonrpc2.Replier, req jsonrpc2.Request) error {
	var params protocol.CodeActionParams
	if err := decodeParams(req.Params(), &params); err != nil {
		return reply(ctx, nil, fmt.Errorf("%s: %w", jsonrpc2.ErrParse, err))
	}
	if !s.isInitialized() {
		return reply(ctx, nil, jsonrpc2.NewError(jsonrpc2.ServerNotInitialized, "server not initialized"))
	}
	if s.isShuttingDown() {
		return reply(ctx, nil, jsonrpc2.NewError(jsonrpc2.InvalidRequest, "server is shutting down"))
	}

	result, err := s.codeActions(ctx, &params)
	return reply(ctx, result, err)
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

func resolveRootCandidates(rootURI uri.URI, workspaceFolders []uri.URI) []string {
	if len(workspaceFolders) > 0 {
		roots := make([]string, 0, len(workspaceFolders))
		for _, folder := range workspaceFolders {
			if folder == "" {
				continue
			}
			roots = append(roots, folder.Filename())
		}
		return roots
	}
	if rootURI != "" {
		return []string{rootURI.Filename()}
	}
	return nil
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

type jsonrpc2StreamWriter interface {
	Write(context.Context, jsonrpc2.Message) (int64, error)
}

type locker interface {
	Lock()
	Unlock()
}

func replier(stream jsonrpc2StreamWriter, writeMu locker, req jsonrpc2.Request) jsonrpc2.Replier {
	return func(ctx context.Context, result interface{}, err error) error {
		call, ok := req.(*jsonrpc2.Call)
		if !ok {
			return nil
		}

		response, responseErr := jsonrpc2.NewResponse(call.ID(), result, err)
		if responseErr != nil {
			return responseErr
		}
		return writeJSONRPCMessage(ctx, stream, writeMu, response)
	}
}

func newNotification(method string, params interface{}) (*jsonrpc2.Notification, error) {
	return jsonrpc2.NewNotification(method, params)
}

func writeJSONRPCMessage(ctx context.Context, stream jsonrpc2StreamWriter, writeMu locker, msg jsonrpc2.Message) error {
	writeMu.Lock()
	defer writeMu.Unlock()
	_, err := stream.Write(ctx, msg)
	return err
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
