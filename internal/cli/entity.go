package cli

import (
	"fmt"
	"sort"

	"github.com/spf13/cobra"
)

func newEntityCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "entity",
		Short: "Manage entity schemas",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, err := contextFromCommand(cmd)
			if err != nil {
				return err
			}
			_, _ = fmt.Fprintln(ctx.Stderr, "entity subcommand required (list|show)")
			return newExitError(1)
		},
	}

	cmd.AddCommand(
		newEntityListCommand(),
		newEntityShowCommand(),
	)

	return cmd
}

func newEntityListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List known entities",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, err := contextFromCommand(cmd)
			if err != nil {
				return err
			}

			cfg, err := loadConfig(ctx)
			if err != nil {
				_, _ = fmt.Fprintf(ctx.Stderr, "entity list: %v\n", err)
				return newExitError(1)
			}

			names := make([]string, 0, len(cfg.Types))
			for name := range cfg.Types {
				names = append(names, name)
			}
			sort.Strings(names)
			for _, name := range names {
				_, _ = fmt.Fprintln(ctx.Stdout, name)
			}
			return nil
		},
	}

	return cmd
}

func newEntityShowCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <name>",
		Short: "Show schema for an entity",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, err := contextFromCommand(cmd)
			if err != nil {
				return err
			}

			if len(args) == 0 {
				_, _ = fmt.Fprintln(ctx.Stderr, "entity show requires an entity name")
				return newExitError(1)
			}

			typeName := args[0]

			cfg, err := loadConfig(ctx)
			if err != nil {
				_, _ = fmt.Fprintf(ctx.Stderr, "entity show: %v\n", err)
				return newExitError(1)
			}

			typeDef, ok := cfg.Types[typeName]
			if !ok {
				_, _ = fmt.Fprintf(ctx.Stderr, "unknown entity %s\n", typeName)
				return newExitError(1)
			}

			if code := writeFormatted(ctx, typeDef); code != 0 {
				return newExitError(code)
			}
			return nil
		},
	}

	return cmd
}
