package cli

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

func cmdInit(ctx *Context, args []string) int {
	fs := flag.NewFlagSet("init", flag.ContinueOnError)
	fs.SetOutput(ctx.Stderr)

	if err := fs.Parse(args); err != nil {
		return 1
	}

	dirs := []string{
		filepath.Join(ctx.Root, "types"),
		filepath.Join(ctx.Root, "data"),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			_, _ = fmt.Fprintf(ctx.Stderr, "init: create %s: %v\n", dir, err)
			return 1
		}
	}

	configPath := ctx.Config
	if err := ensureFile(configPath, defaultConfigTemplate()); err != nil {
		_, _ = fmt.Fprintf(ctx.Stderr, "init: %v\n", err)
		return 1
	}

	_, _ = fmt.Fprintf(ctx.Stdout, "Initialized repository at %s\n", ctx.Root)
	return 0
}

func ensureFile(path, content string) error {
	if _, err := os.Stat(path); err == nil {
		return nil
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	return os.WriteFile(path, []byte(content), 0o644)
}

func defaultConfigTemplate() string {
	return "# mw configuration\nversion: 1\ntypes: {}\n"
}
