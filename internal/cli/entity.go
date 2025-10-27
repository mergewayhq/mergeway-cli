package cli

import (
	"flag"
	"fmt"
	"sort"
)

func cmdEntity(ctx *Context, args []string) int {
	if len(args) == 0 {
		_, _ = fmt.Fprintln(ctx.Stderr, "entity subcommand required (list|show)")
		return 1
	}

	switch args[0] {
	case "list":
		return cmdEntityList(ctx, args[1:])
	case "show":
		return cmdEntityShow(ctx, args[1:])
	default:
		_, _ = fmt.Fprintf(ctx.Stderr, "unknown entity subcommand: %s\n", args[0])
		return 1
	}
}

func cmdEntityList(ctx *Context, args []string) int {
	fs := flag.NewFlagSet("entity list", flag.ContinueOnError)
	fs.SetOutput(ctx.Stderr)
	if err := fs.Parse(args); err != nil {
		return 1
	}

	cfg, err := loadConfig(ctx)
	if err != nil {
		_, _ = fmt.Fprintf(ctx.Stderr, "entity list: %v\n", err)
		return 1
	}

	names := make([]string, 0, len(cfg.Types))
	for name := range cfg.Types {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		_, _ = fmt.Fprintln(ctx.Stdout, name)
	}
	return 0
}

func cmdEntityShow(ctx *Context, args []string) int {
	fs := flag.NewFlagSet("entity show", flag.ContinueOnError)
	fs.SetOutput(ctx.Stderr)
	if err := fs.Parse(args); err != nil {
		return 1
	}

	if fs.NArg() == 0 {
		_, _ = fmt.Fprintln(ctx.Stderr, "entity show requires an entity name")
		return 1
	}

	typeName := fs.Arg(0)

	cfg, err := loadConfig(ctx)
	if err != nil {
		_, _ = fmt.Fprintf(ctx.Stderr, "entity show: %v\n", err)
		return 1
	}

	typeDef, ok := cfg.Types[typeName]
	if !ok {
		_, _ = fmt.Fprintf(ctx.Stderr, "unknown entity %s\n", typeName)
		return 1
	}

	return writeFormatted(ctx, typeDef)
}
