package cli

import (
	"fmt"

	"github.com/mergewayhq/mergeway-cli/internal/config"
	"github.com/spf13/cobra"
)

func newConfigCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage configuration files",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, err := contextFromCommand(cmd)
			if err != nil {
				return err
			}
			_, _ = fmt.Fprintln(ctx.Stderr, "config subcommand required (lint|export)")
			return newExitError(1)
		},
	}

	cmd.AddCommand(
		newConfigLintCommand(),
		newConfigExportCommand(),
	)

	return cmd
}

func newConfigLintCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "lint",
		Short: "Validate configuration files",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, err := contextFromCommand(cmd)
			if err != nil {
				return err
			}

			if _, err := loadConfig(ctx); err != nil {
				_, _ = fmt.Fprintf(ctx.Stderr, "config lint: %v\n", err)
				return newExitError(1)
			}

			_, _ = fmt.Fprintln(ctx.Stdout, "configuration valid")
			return nil
		},
	}

	return cmd
}

func newConfigExportCommand() *cobra.Command {
	var typeName string

	cmd := &cobra.Command{
		Use:   "export",
		Short: "Export entity definition as JSON Schema",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, err := contextFromCommand(cmd)
			if err != nil {
				return err
			}

			if typeName == "" {
				_, _ = fmt.Fprintln(ctx.Stderr, "config export requires --type")
				return newExitError(1)
			}

			cfg, err := loadConfig(ctx)
			if err != nil {
				_, _ = fmt.Fprintf(ctx.Stderr, "config export: %v\n", err)
				return newExitError(1)
			}

			typeDef, ok := cfg.Types[typeName]
			if !ok {
				_, _ = fmt.Fprintf(ctx.Stderr, "unknown type %s\n", typeName)
				return newExitError(1)
			}

			schema, err := buildJSONSchema(typeDef)
			if err != nil {
				_, _ = fmt.Fprintf(ctx.Stderr, "config export: %v\n", err)
				return newExitError(1)
			}
			if code := writeFormatted(ctx, schema); code != 0 {
				return newExitError(code)
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&typeName, "type", "", "Type identifier")

	return cmd
}

func buildJSONSchema(typeDef *config.TypeDefinition) (map[string]any, error) {
	properties := make(map[string]any)
	required := make([]string, 0)

	seen := make(map[string]struct{})
	for _, name := range typeDef.FieldOrder {
		field := typeDef.Fields[name]
		if field == nil {
			continue
		}
		var err error
		properties[name], required, err = appendFieldSchema(properties[name], required, name, field)
		if err != nil {
			return nil, err
		}
		seen[name] = struct{}{}
	}

	for name, field := range typeDef.Fields {
		if _, ok := seen[name]; ok {
			continue
		}
		var err error
		properties[name], required, err = appendFieldSchema(properties[name], required, name, field)
		if err != nil {
			return nil, err
		}
	}

	schema := map[string]any{
		"$schema":    "https://json-schema.org/draft/2020-12/schema",
		"type":       "object",
		"properties": properties,
	}
	if typeDef.Description != "" {
		schema["description"] = typeDef.Description
	}
	if len(required) > 0 {
		schema["required"] = required
	}
	return schema, nil
}

func appendFieldSchema(existing any, required []string, name string, field *config.FieldDefinition) (map[string]any, []string, error) {
	if field == nil {
		return map[string]any{}, required, nil
	}

	prop := map[string]any{}

	switch field.Type {
	case "string", "integer", "number", "boolean":
		prop["type"] = field.Type
	case "enum":
		prop["type"] = "string"
		if len(field.Enum) > 0 {
			prop["enum"] = field.Enum
		}
	case "object":
		sub := &config.TypeDefinition{
			Fields:     field.Properties,
			FieldOrder: field.PropertyOrder,
		}
		var err error
		prop, err = buildJSONSchema(sub)
		if err != nil {
			return nil, required, err
		}
	default:
		if field.HasReferenceUnion() {
			return nil, required, fmt.Errorf("field %q uses reference union %q, which cannot be exported as JSON Schema", name, field.ReferenceLabel())
		}
		prop["type"] = "string"
		prop["x-reference-type"] = field.ReferenceLabel()
	}

	if field.Repeated {
		prop = map[string]any{
			"type":  "array",
			"items": prop,
		}
	}

	if field.Description != "" {
		prop["description"] = field.Description
	}

	if field.Required {
		required = append(required, name)
	}
	return prop, required, nil
}
