package cli

import (
	"fmt"

	"github.com/mergewayhq/mergeway-cli/internal/validation"
	"github.com/mergewayhq/mergeway-cli/internal/workspace"
	"github.com/spf13/cobra"
)

func newValidateCommand() *cobra.Command {
	phaseFlags := multiFlag{}

	cmd := &cobra.Command{
		Use:   "validate",
		Short: "Validate repository contents",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, err := contextFromCommand(cmd)
			if err != nil {
				return err
			}

			opts := validation.Options{
				FailFast: ctx.FailFast,
				Phases:   phaseFlags.Values,
			}

			report, err := workspace.Validate(ctx.Root, ctx.Config, opts)
			if err != nil {
				_, _ = fmt.Fprintf(ctx.Stderr, "validate: %v\n", err)
				return newExitError(1)
			}
			result := report.Result

			if len(result.Errors) == 0 {
				if code := writeFormatted(ctx, map[string]string{"status": "validation succeeded"}); code != 0 {
					return newExitError(code)
				}
				return nil
			}

			if code := writeFormatted(ctx, result.Errors); code != 0 {
				return newExitError(code)
			}
			return newExitError(1)
		},
	}

	cmd.Flags().Var(&phaseFlags, "phase", "Validation phase to run (format|schema|references), repeatable")

	return cmd
}
