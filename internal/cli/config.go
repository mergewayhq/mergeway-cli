package cli

import (
	"flag"
	"fmt"

	"github.com/mergewayhq/mergeway-cli/internal/config"
)

func cmdConfig(ctx *Context, args []string) int {
	if len(args) == 0 {
		_, _ = fmt.Fprintln(ctx.Stderr, "config subcommand required (lint|export)")
		return 1
	}

	switch args[0] {
	case "lint":
		return cmdConfigLint(ctx, args[1:])
	case "export":
		return cmdConfigExport(ctx, args[1:])
	default:
		_, _ = fmt.Fprintf(ctx.Stderr, "unknown config subcommand: %s\n", args[0])
		return 1
	}
}

func cmdConfigLint(ctx *Context, args []string) int {
	fs := flag.NewFlagSet("config lint", flag.ContinueOnError)
	fs.SetOutput(ctx.Stderr)
	if err := fs.Parse(args); err != nil {
		return 1
	}

	if _, err := loadConfig(ctx); err != nil {
		_, _ = fmt.Fprintf(ctx.Stderr, "config lint: %v\n", err)
		return 1
	}

	_, _ = fmt.Fprintln(ctx.Stdout, "configuration valid")
	return 0
}

func cmdConfigExport(ctx *Context, args []string) int {
	fs := flag.NewFlagSet("config export", flag.ContinueOnError)
	fs.SetOutput(ctx.Stderr)
	typeName := fs.String("type", "", "Type identifier")
	if err := fs.Parse(args); err != nil {
		return 1
	}

	if *typeName == "" {
		_, _ = fmt.Fprintln(ctx.Stderr, "config export requires --type")
		return 1
	}

	cfg, err := loadConfig(ctx)
	if err != nil {
		_, _ = fmt.Fprintf(ctx.Stderr, "config export: %v\n", err)
		return 1
	}

	typeDef, ok := cfg.Types[*typeName]
	if !ok {
		_, _ = fmt.Fprintf(ctx.Stderr, "unknown type %s\n", *typeName)
		return 1
	}

	schema := buildJSONSchema(typeDef)
	return writeFormatted(ctx, schema)
}

func buildJSONSchema(typeDef *config.TypeDefinition) map[string]any {
	properties := make(map[string]any)
	required := make([]string, 0)

	for name, field := range typeDef.Fields {
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
			sub := &config.TypeDefinition{Fields: field.Properties}
			prop = buildJSONSchema(sub)
		default:
			prop["type"] = "string"
			prop["x-reference-type"] = field.Type
		}

		if field.Repeated {
			prop = map[string]any{
				"type":  "array",
				"items": prop,
			}
		}

		properties[name] = prop
		if field.Required {
			required = append(required, name)
		}
	}

	schema := map[string]any{
		"$schema":    "https://json-schema.org/draft/2020-12/schema",
		"type":       "object",
		"properties": properties,
	}
	if len(required) > 0 {
		schema["required"] = required
	}
	return schema
}
