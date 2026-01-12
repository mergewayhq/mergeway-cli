package cli

import (
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
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
	cmd := newRootCommand(stdout, stderr)
	cmd.SetArgs(args)

	if err := cmd.Execute(); err != nil {
		var exitErr exitError
		if errors.As(err, &exitErr) {
			return exitErr.Code()
		}
		_, _ = fmt.Fprintln(stderr, err.Error())
		return 1
	}
	return 0
}

type exitError struct {
	code int
}

func (e exitError) Error() string {
	return fmt.Sprintf("exit code %d", e.code)
}

func (e exitError) Code() int {
	if e.code == 0 {
		return 1
	}
	return e.code
}

func newExitError(code int) error {
	return exitError{code: code}
}

func newRootCommand(stdout, stderr io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:           "mw",
		Short:         "Manage mergeway repositories",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			_ = cmd.Help()
			return newExitError(1)
		},
	}
	cmd.SetOut(stdout)
	cmd.SetErr(stderr)

	flags := cmd.PersistentFlags()
	flags.String("root", ".", "Repository root containing config and data directories")
	flags.String("config", "", "Path to configuration entry file")
	flags.String("format", "yaml", "Output format (yaml|json)")
	flags.Bool("fail-fast", false, "Stop validation on first error")
	flags.Bool("yes", false, "Auto-confirm prompts")
	flags.Bool("verbose", false, "Enable verbose logging")

	cmd.AddCommand(
		newInitCommand(),
		newEntityCommand(),
		newListCommand(),
		newGetCommand(),
		newCreateCommand(),
		newUpdateCommand(),
		newDeleteCommand(),
		newExportCommand(),
		newValidateCommand(),
		newFmtCommand(),
		newConfigCommand(),
		newVersionCommand(),
		newGenERDCommand(),
	)

	return cmd
}

func contextFromCommand(cmd *cobra.Command) (*Context, error) {
	root, err := cmd.Flags().GetString("root")
	if err != nil {
		return nil, err
	}
	configPath, err := cmd.Flags().GetString("config")
	if err != nil {
		return nil, err
	}
	format, err := cmd.Flags().GetString("format")
	if err != nil {
		return nil, err
	}
	failFast, err := cmd.Flags().GetBool("fail-fast")
	if err != nil {
		return nil, err
	}
	yes, err := cmd.Flags().GetBool("yes")
	if err != nil {
		return nil, err
	}
	verbose, err := cmd.Flags().GetBool("verbose")
	if err != nil {
		return nil, err
	}

	ctx := &Context{
		Root:     root,
		Config:   configPath,
		Format:   strings.ToLower(format),
		FailFast: failFast,
		Yes:      yes,
		Verbose:  verbose,
		Stdout:   cmd.OutOrStdout(),
		Stderr:   cmd.ErrOrStderr(),
	}

	if ctx.Config == "" {
		ctx.Config = filepath.Join(ctx.Root, "mergeway.yaml")
	}

	return ctx, nil
}
