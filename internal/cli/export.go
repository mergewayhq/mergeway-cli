package cli

import (
	"flag"
	"fmt"
	"os"
	"sort"
)

func cmdExport(ctx *Context, args []string) int {
	fs := flag.NewFlagSet("export", flag.ContinueOnError)
	fs.SetOutput(ctx.Stderr)
	outputPath := fs.String("output", "", "Path to output file (defaults to STDOUT)")
	if err := fs.Parse(args); err != nil {
		return 1
	}

	include := fs.Args()

	cfg, err := loadConfig(ctx)
	if err != nil {
		_, _ = fmt.Fprintf(ctx.Stderr, "export: %v\n", err)
		return 1
	}

	types := make([]string, 0, len(cfg.Types))
	if len(include) == 0 {
		for name := range cfg.Types {
			types = append(types, name)
		}
		sort.Strings(types)
	} else {
		seen := make(map[string]struct{}, len(include))
		for _, name := range include {
			if name == "" {
				continue
			}
			if _, ok := cfg.Types[name]; !ok {
				_, _ = fmt.Fprintf(ctx.Stderr, "export: unknown entity %s\n", name)
				return 1
			}
			if _, dup := seen[name]; dup {
				continue
			}
			seen[name] = struct{}{}
			types = append(types, name)
		}
	}

	store, err := loadStore(ctx, cfg)
	if err != nil {
		_, _ = fmt.Fprintf(ctx.Stderr, "export: %v\n", err)
		return 1
	}

	result := make(map[string]any, len(types))
	for _, typeName := range types {
		objects, err := store.LoadAll(typeName)
		if err != nil {
			_, _ = fmt.Fprintf(ctx.Stderr, "export: %v\n", err)
			return 1
		}
		if len(objects) > 1 {
			sort.Slice(objects, func(i, j int) bool {
				return objects[i].ID < objects[j].ID
			})
		}

		records := make([]map[string]any, len(objects))
		for i, obj := range objects {
			records[i] = obj.Fields
		}
		result[typeName] = records
	}

	var writer = ctx.Stdout
	if *outputPath != "" {
		f, err := os.Create(*outputPath)
		if err != nil {
			_, _ = fmt.Fprintf(ctx.Stderr, "export: %v\n", err)
			return 1
		}
		defer func() {
			_ = f.Close()
		}()
		writer = f
	}

	originalStdout := ctx.Stdout
	ctx.Stdout = writer
	defer func() {
		ctx.Stdout = originalStdout
	}()

	return writeFormatted(ctx, result)
}
