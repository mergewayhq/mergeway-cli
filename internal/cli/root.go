package cli

import (
	"flag"
	"fmt"
	"io"
	"path/filepath"
	"strings"
)

type Context struct {
	Root     string
	Config   string
	Format   string
	FailFast bool
	Yes      bool
	Verbose  bool
	Stdout   io.Writer
	Stderr   io.Writer
}

// Run executes the CLI. It returns an exit code.
func Run(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("mw", flag.ContinueOnError)
	fs.SetOutput(stderr)

	root := fs.String("root", ".", "Repository root containing config and data directories")
	configPath := fs.String("config", "", "Path to configuration entry file")
	format := fs.String("format", "yaml", "Output format (yaml|json)")
	failFast := fs.Bool("fail-fast", false, "Stop validation on first error")
	yes := fs.Bool("yes", false, "Auto-confirm prompts")
	verbose := fs.Bool("verbose", false, "Enable verbose logging")

	if err := fs.Parse(args); err != nil {
		return 1
	}

	remaining := fs.Args()
	if len(remaining) == 0 {
		printUsage(stdout)
		return 1
	}

	ctx := &Context{
		Root:     *root,
		Config:   *configPath,
		Format:   strings.ToLower(*format),
		FailFast: *failFast,
		Yes:      *yes,
		Verbose:  *verbose,
		Stdout:   stdout,
		Stderr:   stderr,
	}

	if ctx.Config == "" {
		ctx.Config = filepath.Join(ctx.Root, "mergeway.yaml")
	}

	switch remaining[0] {
	case "init":
		return cmdInit(ctx, remaining[1:])
	case "type":
		return cmdType(ctx, remaining[1:])
	case "list":
		return cmdList(ctx, remaining[1:])
	case "get":
		return cmdGet(ctx, remaining[1:])
	case "create":
		return cmdCreate(ctx, remaining[1:])
	case "update":
		return cmdUpdate(ctx, remaining[1:])
	case "delete":
		return cmdDelete(ctx, remaining[1:])
	case "validate":
		return cmdValidate(ctx, remaining[1:])
	case "config":
		return cmdConfig(ctx, remaining[1:])
	case "version":
		return cmdVersion(ctx, remaining[1:])
	case "help", "--help", "-h":
		printUsage(stdout)
		return 0
	default:
		_, _ = fmt.Fprintf(stderr, "unknown command: %s\n", remaining[0])
		printUsage(stderr)
		return 1
	}
}

func printUsage(w io.Writer) {
	_, _ = fmt.Fprintln(w, "Usage: mw [global flags] <command> [args]")
	_, _ = fmt.Fprintln(w, "\nCommands:")
	_, _ = fmt.Fprintln(w, "  init                      Scaffold repository structure")
	_, _ = fmt.Fprintln(w, "  type list                 List known types")
	_, _ = fmt.Fprintln(w, "  type show <type>          Show schema for a type")
	_, _ = fmt.Fprintln(w, "  list                      List object identifiers")
	_, _ = fmt.Fprintln(w, "  get                       Get an object")
	_, _ = fmt.Fprintln(w, "  create                    Create an object")
	_, _ = fmt.Fprintln(w, "  update                    Update an object")
	_, _ = fmt.Fprintln(w, "  delete                    Delete an object")
	_, _ = fmt.Fprintln(w, "  validate                  Validate repository contents")
	_, _ = fmt.Fprintln(w, "  config lint               Validate configuration files")
	_, _ = fmt.Fprintln(w, "  config export             Export entity spec as JSON Schema")
	_, _ = fmt.Fprintln(w, "  version                   Display CLI build information")
}
