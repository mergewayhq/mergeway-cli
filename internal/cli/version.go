package cli

import (
	"github.com/mergewayhq/mergeway-cli/internal/version"
	"github.com/spf13/cobra"
)

func newVersionCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Display CLI build information",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, err := contextFromCommand(cmd)
			if err != nil {
				return err
			}

			if code := writeFormatted(ctx, version.Current()); code != 0 {
				return newExitError(code)
			}
			return nil
		},
	}

	return cmd
}
