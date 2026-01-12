package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newListCommand() *cobra.Command {
	var typeName string
	var filterExpr string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List object identifiers",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, err := contextFromCommand(cmd)
			if err != nil {
				return err
			}

			if typeName == "" {
				_, _ = fmt.Fprintln(ctx.Stderr, "list requires --type")
				return newExitError(1)
			}

			cfg, err := loadConfig(ctx)
			if err != nil {
				_, _ = fmt.Fprintf(ctx.Stderr, "list: %v\n", err)
				return newExitError(1)
			}

			store, err := loadStore(ctx, cfg)
			if err != nil {
				_, _ = fmt.Fprintf(ctx.Stderr, "list: %v\n", err)
				return newExitError(1)
			}

			if err := emitList(ctx, store, typeName, filterExpr); err != nil {
				_, _ = fmt.Fprintf(ctx.Stderr, "list: %v\n", err)
				return newExitError(1)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&typeName, "type", "", "Type identifier")
	cmd.Flags().StringVar(&filterExpr, "filter", "", "Simple filter expression key=value")

	return cmd
}

func newGetCommand() *cobra.Command {
	var typeName string

	cmd := &cobra.Command{
		Use:   "get <id>",
		Short: "Get an object",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, err := contextFromCommand(cmd)
			if err != nil {
				return err
			}

			if len(args) == 0 {
				_, _ = fmt.Fprintln(ctx.Stderr, "get requires an identifier")
				return newExitError(1)
			}
			id := args[0]

			cfg, err := loadConfig(ctx)
			if err != nil {
				_, _ = fmt.Fprintf(ctx.Stderr, "get: %v\n", err)
				return newExitError(1)
			}

			store, err := loadStore(ctx, cfg)
			if err != nil {
				_, _ = fmt.Fprintf(ctx.Stderr, "get: %v\n", err)
				return newExitError(1)
			}

			if typeName == "" {
				_, _ = fmt.Fprintln(ctx.Stderr, "get requires --type")
				return newExitError(1)
			}

			obj, err := store.Get(typeName, id)
			if err != nil {
				_, _ = fmt.Fprintf(ctx.Stderr, "get: %v\n", err)
				return newExitError(1)
			}

			if code := writeFormatted(ctx, obj.Fields); code != 0 {
				return newExitError(code)
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&typeName, "type", "", "Type identifier")

	return cmd
}

func newCreateCommand() *cobra.Command {
	var typeName string
	var filePath string
	var idFlag string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create an object",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, err := contextFromCommand(cmd)
			if err != nil {
				return err
			}

			if typeName == "" {
				_, _ = fmt.Fprintln(ctx.Stderr, "create requires --type")
				return newExitError(1)
			}

			payload, err := readPayload(filePath)
			if err != nil {
				_, _ = fmt.Fprintf(ctx.Stderr, "create: %v\n", err)
				return newExitError(1)
			}

			cfg, err := loadConfig(ctx)
			if err != nil {
				_, _ = fmt.Fprintf(ctx.Stderr, "create: %v\n", err)
				return newExitError(1)
			}

			typeDef, ok := cfg.Types[typeName]
			if !ok {
				_, _ = fmt.Fprintf(ctx.Stderr, "create: unknown type %s\n", typeName)
				return newExitError(1)
			}

			if idFlag != "" {
				idField := typeDef.Identifier.Field
				if idField == "" {
					idField = "id"
				}
				payload[idField] = idFlag
			}

			store, err := loadStore(ctx, cfg)
			if err != nil {
				_, _ = fmt.Fprintf(ctx.Stderr, "create: %v\n", err)
				return newExitError(1)
			}

			obj, err := store.Create(typeName, payload)
			if err != nil {
				_, _ = fmt.Fprintf(ctx.Stderr, "create: %v\n", err)
				return newExitError(1)
			}

			_, _ = fmt.Fprintf(ctx.Stdout, "%s %s created\n", obj.Type, obj.ID)
			return nil
		},
	}

	cmd.Flags().StringVar(&typeName, "type", "", "Type identifier")
	cmd.Flags().StringVar(&filePath, "file", "", "Path to payload file (defaults to STDIN)")
	cmd.Flags().StringVar(&idFlag, "id", "", "Override object identifier")

	return cmd
}

func newUpdateCommand() *cobra.Command {
	var typeName string
	var filePath string
	var merge bool
	var idFlag string

	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update an object",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, err := contextFromCommand(cmd)
			if err != nil {
				return err
			}

			if typeName == "" {
				_, _ = fmt.Fprintln(ctx.Stderr, "update requires --type")
				return newExitError(1)
			}

			if idFlag == "" {
				_, _ = fmt.Fprintln(ctx.Stderr, "update requires --id")
				return newExitError(1)
			}

			payload, err := readPayload(filePath)
			if err != nil {
				_, _ = fmt.Fprintf(ctx.Stderr, "update: %v\n", err)
				return newExitError(1)
			}

			cfg, err := loadConfig(ctx)
			if err != nil {
				_, _ = fmt.Fprintf(ctx.Stderr, "update: %v\n", err)
				return newExitError(1)
			}

			store, err := loadStore(ctx, cfg)
			if err != nil {
				_, _ = fmt.Fprintf(ctx.Stderr, "update: %v\n", err)
				return newExitError(1)
			}

			obj, err := store.Update(typeName, idFlag, payload, merge)
			if err != nil {
				_, _ = fmt.Fprintf(ctx.Stderr, "update: %v\n", err)
				return newExitError(1)
			}

			_, _ = fmt.Fprintf(ctx.Stdout, "%s %s updated\n", obj.Type, obj.ID)
			return nil
		},
	}

	cmd.Flags().StringVar(&typeName, "type", "", "Type identifier")
	cmd.Flags().StringVar(&filePath, "file", "", "Path to payload file (defaults to STDIN)")
	cmd.Flags().BoolVar(&merge, "merge", false, "Merge fields instead of replacing")
	cmd.Flags().StringVar(&idFlag, "id", "", "Object identifier")

	return cmd
}

func newDeleteCommand() *cobra.Command {
	var typeName string

	cmd := &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete an object",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, err := contextFromCommand(cmd)
			if err != nil {
				return err
			}

			if len(args) == 0 {
				_, _ = fmt.Fprintln(ctx.Stderr, "delete requires an identifier")
				return newExitError(1)
			}
			id := args[0]

			if typeName == "" {
				_, _ = fmt.Fprintln(ctx.Stderr, "delete requires --type")
				return newExitError(1)
			}

			if !ctx.Yes {
				confirmed, err := confirm(ctx.Stdin(), ctx.Stderr, fmt.Sprintf("Delete %s %s? [y/N]: ", typeName, id))
				if err != nil {
					_, _ = fmt.Fprintf(ctx.Stderr, "delete: %v\n", err)
					return newExitError(1)
				}
				if !confirmed {
					_, _ = fmt.Fprintln(ctx.Stderr, "aborted")
					return newExitError(1)
				}
			}

			cfg, err := loadConfig(ctx)
			if err != nil {
				_, _ = fmt.Fprintf(ctx.Stderr, "delete: %v\n", err)
				return newExitError(1)
			}

			store, err := loadStore(ctx, cfg)
			if err != nil {
				_, _ = fmt.Fprintf(ctx.Stderr, "delete: %v\n", err)
				return newExitError(1)
			}

			if err := store.Delete(typeName, id); err != nil {
				_, _ = fmt.Fprintf(ctx.Stderr, "delete: %v\n", err)
				return newExitError(1)
			}

			_, _ = fmt.Fprintf(ctx.Stdout, "%s %s deleted\n", typeName, id)
			return nil
		},
	}

	cmd.Flags().StringVar(&typeName, "type", "", "Type identifier")

	return cmd
}
