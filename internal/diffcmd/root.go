package diffcmd

import (
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	diffpkg "github.com/mergewayhq/mergeway-cli/internal/diff"
	"github.com/spf13/cobra"
)

type context struct {
	Root   string
	Config string
	Format string
	Stdout io.Writer
	Stderr io.Writer
}

// Run executes the mergeway-diff CLI. It returns an exit code.
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
		Use:           "mergeway-diff [<left>] [<right>]",
		Short:         "Compare Mergeway-managed data between snapshots",
		SilenceUsage:  true,
		SilenceErrors: true,
		Long: `Compare Mergeway-managed data between logical repository snapshots.

This command is a data-only diff. It compares logical Mergeway records and excludes configuration files entirely.

Snapshot modes:
  mergeway-diff                compare HEAD data vs working tree data with unstaged changes only
  mergeway-diff <left>         compare <left> vs current working tree data including unstaged changes
  mergeway-diff <left> <right> compare <left> vs <right>

Use --format json to emit machine-readable semantic diff output.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, err := contextFromCommand(cmd)
			if err != nil {
				return err
			}

			output, err := diffpkg.Run(diffpkg.Options{
				Root:   ctx.Root,
				Config: ctx.Config,
				Args:   args,
				JSON:   ctx.Format == "json",
			})
			if err != nil {
				if errors.Is(err, diffpkg.ErrTooManyArgs) {
					_ = cmd.Help()
				}
				_, _ = fmt.Fprintln(ctx.Stderr, diffpkg.FormatCommandError(err))
				return newExitError(1)
			}

			_, _ = fmt.Fprint(ctx.Stdout, output)
			return nil
		},
	}
	cmd.SetOut(stdout)
	cmd.SetErr(stderr)

	flags := cmd.Flags()
	flags.String("root", ".", "Repository root containing config and data directories")
	flags.String("config", "", "Path to configuration entry file")
	flags.String("format", "yaml", "Output format (yaml|json)")

	return cmd
}

func contextFromCommand(cmd *cobra.Command) (*context, error) {
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

	ctx := &context{
		Root:   root,
		Config: configPath,
		Format: strings.ToLower(format),
		Stdout: cmd.OutOrStdout(),
		Stderr: cmd.ErrOrStderr(),
	}

	if ctx.Config == "" {
		ctx.Config = filepath.Join(ctx.Root, "mergeway.yaml")
	}

	return ctx, nil
}
