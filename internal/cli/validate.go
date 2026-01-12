package cli

import (
	"fmt"

	"github.com/mergewayhq/mergeway-cli/internal/validation"
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

			cfg, err := loadConfig(ctx)
			if err != nil {
				_, _ = fmt.Fprintf(ctx.Stderr, "validate: %v\n", err)
				return newExitError(1)
			}

			opts := validation.Options{
				FailFast: ctx.FailFast,
				Phases:   phaseFlags.Values,
			}

			result, err := validation.Validate(ctx.Root, cfg, opts)
			if err != nil {
				_, _ = fmt.Fprintf(ctx.Stderr, "validate: %v\n", err)
				return newExitError(1)
			}

			if len(result.Errors) == 0 {
				_, _ = fmt.Fprintln(ctx.Stdout, "validation succeeded")
				return nil
			}

			if code := writeFormatted(ctx, result.Errors); code != 0 {
				return newExitError(code)
			}
			return nil
		},
	}

	cmd.Flags().Var(&phaseFlags, "phase", "Validation phase to run (format|schema|references), repeatable")

	return cmd
}
