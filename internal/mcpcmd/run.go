package mcpcmd

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/mergewayhq/mergeway-cli/internal/config"
	"github.com/mergewayhq/mergeway-cli/internal/workspace"
)

const (
	defaultRoot         = "."
	defaultTransport    = "stdio"
	defaultHTTPListen   = "127.0.0.1:8080"
	defaultHTTPBasePath = "/"
)

// Invocation captures the validated startup contract for mergeway-mcp.
type Invocation struct {
	Root         string
	ConfigPath   string
	Transport    string
	HTTPListen   string
	HTTPBasePath string
	Entities     []string
	Stdin        io.Reader
	Stdout       io.Writer
	Stderr       io.Writer
}

// StartFunc launches the server using a validated invocation.
type StartFunc func(context.Context, Invocation) error

// Options configures the command runner.
type Options struct {
	Start StartFunc
}

// Run parses args, validates startup config, and launches the server.
func Run(ctx context.Context, args []string, stdin io.Reader, stdout, stderr io.Writer, opts Options) int {
	invocation, helpShown, err := parseInvocation(args, stdin, stdout, stderr)
	if helpShown {
		return 0
	}
	if err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0
		}
		_, _ = fmt.Fprintf(stderr, "mergeway-mcp: %v\n", err)
		return 2
	}

	if opts.Start == nil {
		_, _ = fmt.Fprintln(stderr, "mergeway-mcp: server implementation not configured")
		return 1
	}

	if err := opts.Start(ctx, invocation); err != nil {
		_, _ = fmt.Fprintf(stderr, "mergeway-mcp: %v\n", err)
		return 1
	}

	return 0
}

func parseInvocation(args []string, stdin io.Reader, stdout, stderr io.Writer) (Invocation, bool, error) {
	var (
		root       = defaultRoot
		transport  = defaultTransport
		httpListen = defaultHTTPListen
		basePath   = defaultHTTPBasePath
		entities   stringListFlag
	)

	fs := flag.NewFlagSet("mergeway-mcp", flag.ContinueOnError)
	fs.SetOutput(stderr)
	fs.Usage = func() {
		_, _ = fmt.Fprintln(stderr, "Usage: mergeway-mcp [flags]")
		_, _ = fmt.Fprintln(stderr)
		_, _ = fmt.Fprintln(stderr, "Run the Mergeway MCP server.")
		_, _ = fmt.Fprintln(stderr)
		_, _ = fmt.Fprintln(stderr, "Flags:")
		fs.PrintDefaults()
	}

	fs.StringVar(&root, "root", defaultRoot, "Root folder of the Mergeway repository")
	fs.StringVar(&transport, "transport", defaultTransport, "Transport protocol (stdio|http)")
	fs.StringVar(&httpListen, "http-listen", defaultHTTPListen, "Listen address for HTTP transport")
	fs.StringVar(&basePath, "http-base-path", defaultHTTPBasePath, "Base path for HTTP transport")
	fs.Var(&entities, "entity", "Allow only the named entity (repeatable)")

	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return Invocation{}, true, nil
		}
		return Invocation{}, false, err
	}

	if remaining := fs.Args(); len(remaining) > 0 {
		return Invocation{}, false, fmt.Errorf("unexpected positional arguments: %s", strings.Join(remaining, " "))
	}

	visited := visitedFlags(fs)

	resolvedRoot, err := resolveRoot(root)
	if err != nil {
		return Invocation{}, false, err
	}

	normalizedTransport, err := normalizeTransport(transport)
	if err != nil {
		return Invocation{}, false, err
	}

	normalizedBasePath, err := normalizeHTTPBasePath(basePath)
	if err != nil {
		return Invocation{}, false, err
	}

	if err := validateTransportFlags(normalizedTransport, visited); err != nil {
		return Invocation{}, false, err
	}
	if normalizedTransport == "http" {
		if err := validateHTTPListen(httpListen); err != nil {
			return Invocation{}, false, err
		}
	}

	configPath, cfg, err := loadRepositoryConfig(resolvedRoot)
	if err != nil {
		return Invocation{}, false, err
	}

	allowedEntities, err := normalizeEntities(entities.values, cfg)
	if err != nil {
		return Invocation{}, false, err
	}

	return Invocation{
		Root:         resolvedRoot,
		ConfigPath:   configPath,
		Transport:    normalizedTransport,
		HTTPListen:   httpListen,
		HTTPBasePath: normalizedBasePath,
		Entities:     allowedEntities,
		Stdin:        stdin,
		Stdout:       stdout,
		Stderr:       stderr,
	}, false, nil
}

func resolveRoot(root string) (string, error) {
	resolvedRoot, err := filepath.Abs(root)
	if err != nil {
		return "", fmt.Errorf("resolve root: %w", err)
	}

	info, err := os.Stat(resolvedRoot)
	if err != nil {
		return "", fmt.Errorf("stat root %s: %w", resolvedRoot, err)
	}
	if !info.IsDir() {
		return "", fmt.Errorf("root %s is not a directory", resolvedRoot)
	}

	return resolvedRoot, nil
}

func normalizeTransport(transport string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(transport)) {
	case "stdio":
		return "stdio", nil
	case "http":
		return "http", nil
	default:
		return "", fmt.Errorf("invalid --transport %q; must be stdio or http", transport)
	}
}

func normalizeHTTPBasePath(basePath string) (string, error) {
	trimmed := strings.TrimSpace(basePath)
	if trimmed == "" {
		return "", errors.New("--http-base-path cannot be empty")
	}
	if !strings.HasPrefix(trimmed, "/") {
		return "", fmt.Errorf("invalid --http-base-path %q; must start with /", basePath)
	}

	cleaned := path.Clean(trimmed)
	if !strings.HasPrefix(cleaned, "/") {
		cleaned = "/" + cleaned
	}
	return cleaned, nil
}

func validateTransportFlags(transport string, visited map[string]struct{}) error {
	if transport == "http" {
		return nil
	}

	if _, ok := visited["http-listen"]; ok {
		return errors.New("--http-listen requires --transport=http")
	}
	if _, ok := visited["http-base-path"]; ok {
		return errors.New("--http-base-path requires --transport=http")
	}
	return nil
}

func validateHTTPListen(listen string) error {
	if strings.TrimSpace(listen) == "" {
		return errors.New("--http-listen cannot be empty")
	}
	if _, err := net.ResolveTCPAddr("tcp", listen); err != nil {
		return fmt.Errorf("invalid --http-listen %q: %w", listen, err)
	}
	return nil
}

func loadRepositoryConfig(root string) (string, *config.Config, error) {
	configPath, found, err := workspace.DetectConfigPath(root)
	if err != nil {
		return "", nil, err
	}
	if !found {
		return "", nil, fmt.Errorf("no mergeway config found under root %s", root)
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		return "", nil, err
	}
	return configPath, cfg, nil
}

func normalizeEntities(values []string, cfg *config.Config) ([]string, error) {
	if len(values) == 0 {
		return nil, nil
	}
	if cfg == nil {
		return nil, errors.New("config is required for entity validation")
	}

	seen := make(map[string]struct{}, len(values))
	entities := make([]string, 0, len(values))
	for _, raw := range values {
		name := strings.TrimSpace(raw)
		if name == "" {
			return nil, errors.New("--entity cannot be empty")
		}
		if _, ok := cfg.Types[name]; !ok {
			return nil, fmt.Errorf("unknown --entity %q", name)
		}
		if _, dup := seen[name]; dup {
			continue
		}
		seen[name] = struct{}{}
		entities = append(entities, name)
	}
	return entities, nil
}

func visitedFlags(fs *flag.FlagSet) map[string]struct{} {
	visited := make(map[string]struct{})
	fs.Visit(func(f *flag.Flag) {
		visited[f.Name] = struct{}{}
	})
	return visited
}

type stringListFlag struct {
	values []string
}

func (f *stringListFlag) String() string {
	return strings.Join(f.values, ",")
}

func (f *stringListFlag) Set(value string) error {
	f.values = append(f.values, value)
	return nil
}
