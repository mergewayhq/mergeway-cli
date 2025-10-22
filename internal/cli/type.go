package cli

import (
	"flag"
	"fmt"
	"sort"
)

func cmdType(ctx *Context, args []string) int {
	if len(args) == 0 {
		_, _ = fmt.Fprintln(ctx.Stderr, "type subcommand required (list|show)")
		return 1
	}

	switch args[0] {
	case "list":
		return cmdTypeList(ctx, args[1:])
	case "show":
		return cmdTypeShow(ctx, args[1:])
	default:
		_, _ = fmt.Fprintf(ctx.Stderr, "unknown type subcommand: %s\n", args[0])
		return 1
	}
}

func cmdTypeList(ctx *Context, args []string) int {
	fs := flag.NewFlagSet("type list", flag.ContinueOnError)
	fs.SetOutput(ctx.Stderr)
	if err := fs.Parse(args); err != nil {
		return 1
	}

	cfg, err := loadConfig(ctx)
	if err != nil {
		_, _ = fmt.Fprintf(ctx.Stderr, "type list: %v\n", err)
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

func cmdTypeShow(ctx *Context, args []string) int {
	fs := flag.NewFlagSet("type show", flag.ContinueOnError)
	fs.SetOutput(ctx.Stderr)
	if err := fs.Parse(args); err != nil {
		return 1
	}

	if fs.NArg() == 0 {
		_, _ = fmt.Fprintln(ctx.Stderr, "type show requires a type name")
		return 1
	}

	typeName := fs.Arg(0)

	cfg, err := loadConfig(ctx)
	if err != nil {
		_, _ = fmt.Fprintf(ctx.Stderr, "type show: %v\n", err)
		return 1
	}

	typeDef, ok := cfg.Types[typeName]
	if !ok {
		_, _ = fmt.Fprintf(ctx.Stderr, "unknown type %s\n", typeName)
		return 1
	}

	return writeFormatted(ctx, typeDef)
}
