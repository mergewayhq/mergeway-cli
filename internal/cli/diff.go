package cli

import (
	"errors"
	"fmt"

	diffpkg "github.com/mergewayhq/mergeway-cli/internal/diff"
	"github.com/spf13/cobra"
)

func newDiffCommand() *cobra.Command {
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "diff [<left>] [<right>]",
		Short: "Compare Mergeway-managed data between snapshots",
		Long: `Compare Mergeway-managed data between logical repository snapshots.

This command is a data-only diff. It compares logical Mergeway records and excludes configuration files entirely.

Snapshot modes:
  diff                compare HEAD data vs working tree data with unstaged changes only
  diff <left>         compare <left> vs current working tree data including unstaged changes
  diff <left> <right> compare <left> vs <right>

Flags:
  --json              emit machine-readable semantic diff output`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, err := contextFromCommand(cmd)
			if err != nil {
				return err
			}

			output, err := diffpkg.Run(diffpkg.Options{
				Root:   ctx.Root,
				Config: ctx.Config,
				Args:   args,
				JSON:   jsonOutput,
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

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Emit machine-readable semantic diff output")

	return cmd
}
