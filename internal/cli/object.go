package cli

import (
	"flag"
	"fmt"
)

func cmdList(ctx *Context, args []string) int {
	fs := flag.NewFlagSet("list", flag.ContinueOnError)
	fs.SetOutput(ctx.Stderr)
	typeName := fs.String("type", "", "Type identifier")
	filterExpr := fs.String("filter", "", "Simple filter expression key=value")
	if err := fs.Parse(args); err != nil {
		return 1
	}

	if *typeName == "" {
		_, _ = fmt.Fprintln(ctx.Stderr, "list requires --type")
		return 1
	}

	cfg, err := loadConfig(ctx)
	if err != nil {
		_, _ = fmt.Fprintf(ctx.Stderr, "list: %v\n", err)
		return 1
	}

	store, err := loadStore(ctx, cfg)
	if err != nil {
		_, _ = fmt.Fprintf(ctx.Stderr, "list: %v\n", err)
		return 1
	}

	objects, err := store.LoadAll(*typeName)
	if err != nil {
		_, _ = fmt.Fprintf(ctx.Stderr, "list: %v\n", err)
		return 1
	}

	key, value := parseFilter(*filterExpr)

	for _, obj := range objects {
		if key != "" {
			if val, ok := obj.Fields[key]; !ok || fmt.Sprint(val) != value {
				continue
			}
		}
		_, _ = fmt.Fprintln(ctx.Stdout, obj.ID)
	}

	return 0
}

func cmdGet(ctx *Context, args []string) int {
	fs := flag.NewFlagSet("get", flag.ContinueOnError)
	fs.SetOutput(ctx.Stderr)
	typeName := fs.String("type", "", "Type identifier")
	if err := fs.Parse(args); err != nil {
		return 1
	}

	if fs.NArg() == 0 {
		_, _ = fmt.Fprintln(ctx.Stderr, "get requires an identifier")
		return 1
	}
	id := fs.Arg(0)

	cfg, err := loadConfig(ctx)
	if err != nil {
		_, _ = fmt.Fprintf(ctx.Stderr, "get: %v\n", err)
		return 1
	}

	store, err := loadStore(ctx, cfg)
	if err != nil {
		_, _ = fmt.Fprintf(ctx.Stderr, "get: %v\n", err)
		return 1
	}

	typeID := *typeName
	if typeID == "" {
		_, _ = fmt.Fprintln(ctx.Stderr, "get requires --type")
		return 1
	}

	obj, err := store.Get(typeID, id)
	if err != nil {
		_, _ = fmt.Fprintf(ctx.Stderr, "get: %v\n", err)
		return 1
	}

	return writeFormatted(ctx, obj.Fields)
}

func cmdCreate(ctx *Context, args []string) int {
	fs := flag.NewFlagSet("create", flag.ContinueOnError)
	fs.SetOutput(ctx.Stderr)
	typeName := fs.String("type", "", "Type identifier")
	filePath := fs.String("file", "", "Path to payload file (defaults to STDIN)")
	idFlag := fs.String("id", "", "Override object identifier")
	if err := fs.Parse(args); err != nil {
		return 1
	}

	if *typeName == "" {
		_, _ = fmt.Fprintln(ctx.Stderr, "create requires --type")
		return 1
	}

	payload, err := readPayload(*filePath)
	if err != nil {
		_, _ = fmt.Fprintf(ctx.Stderr, "create: %v\n", err)
		return 1
	}

	if *idFlag != "" {
		payload["id"] = *idFlag
	}

	cfg, err := loadConfig(ctx)
	if err != nil {
		_, _ = fmt.Fprintf(ctx.Stderr, "create: %v\n", err)
		return 1
	}

	store, err := loadStore(ctx, cfg)
	if err != nil {
		_, _ = fmt.Fprintf(ctx.Stderr, "create: %v\n", err)
		return 1
	}

	obj, err := store.Create(*typeName, payload)
	if err != nil {
		_, _ = fmt.Fprintf(ctx.Stderr, "create: %v\n", err)
		return 1
	}

	_, _ = fmt.Fprintf(ctx.Stdout, "%s %s created\n", obj.Type, obj.ID)
	return 0
}

func cmdUpdate(ctx *Context, args []string) int {
	fs := flag.NewFlagSet("update", flag.ContinueOnError)
	fs.SetOutput(ctx.Stderr)
	typeName := fs.String("type", "", "Type identifier")
	filePath := fs.String("file", "", "Path to payload file (defaults to STDIN)")
	merge := fs.Bool("merge", false, "Merge fields instead of replacing")
	idFlag := fs.String("id", "", "Object identifier")
	if err := fs.Parse(args); err != nil {
		return 1
	}

	if *typeName == "" {
		_, _ = fmt.Fprintln(ctx.Stderr, "update requires --type")
		return 1
	}

	if *idFlag == "" {
		_, _ = fmt.Fprintln(ctx.Stderr, "update requires --id")
		return 1
	}

	payload, err := readPayload(*filePath)
	if err != nil {
		_, _ = fmt.Fprintf(ctx.Stderr, "update: %v\n", err)
		return 1
	}

	cfg, err := loadConfig(ctx)
	if err != nil {
		_, _ = fmt.Fprintf(ctx.Stderr, "update: %v\n", err)
		return 1
	}

	store, err := loadStore(ctx, cfg)
	if err != nil {
		_, _ = fmt.Fprintf(ctx.Stderr, "update: %v\n", err)
		return 1
	}

	obj, err := store.Update(*typeName, *idFlag, payload, *merge)
	if err != nil {
		_, _ = fmt.Fprintf(ctx.Stderr, "update: %v\n", err)
		return 1
	}

	_, _ = fmt.Fprintf(ctx.Stdout, "%s %s updated\n", obj.Type, obj.ID)
	return 0
}

func cmdDelete(ctx *Context, args []string) int {
	fs := flag.NewFlagSet("delete", flag.ContinueOnError)
	fs.SetOutput(ctx.Stderr)
	typeName := fs.String("type", "", "Type identifier")
	if err := fs.Parse(args); err != nil {
		return 1
	}

	if fs.NArg() == 0 {
		_, _ = fmt.Fprintln(ctx.Stderr, "delete requires an identifier")
		return 1
	}
	id := fs.Arg(0)

	if *typeName == "" {
		_, _ = fmt.Fprintln(ctx.Stderr, "delete requires --type")
		return 1
	}

	if !ctx.Yes {
		confirmed, err := confirm(ctx.Stdin(), ctx.Stderr, fmt.Sprintf("Delete %s %s? [y/N]: ", *typeName, id))
		if err != nil {
			_, _ = fmt.Fprintf(ctx.Stderr, "delete: %v\n", err)
			return 1
		}
		if !confirmed {
			_, _ = fmt.Fprintln(ctx.Stderr, "aborted")
			return 1
		}
	}

	cfg, err := loadConfig(ctx)
	if err != nil {
		_, _ = fmt.Fprintf(ctx.Stderr, "delete: %v\n", err)
		return 1
	}

	store, err := loadStore(ctx, cfg)
	if err != nil {
		_, _ = fmt.Fprintf(ctx.Stderr, "delete: %v\n", err)
		return 1
	}

	if err := store.Delete(*typeName, id); err != nil {
		_, _ = fmt.Fprintf(ctx.Stderr, "delete: %v\n", err)
		return 1
	}

	_, _ = fmt.Fprintf(ctx.Stdout, "%s %s deleted\n", *typeName, id)
	return 0
}
