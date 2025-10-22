package cli

import (
	"flag"

	"github.com/mergewayhq/mergeway-cli/internal/version"
)

func cmdVersion(ctx *Context, args []string) int {
	fs := flag.NewFlagSet("version", flag.ContinueOnError)
	fs.SetOutput(ctx.Stderr)

	if err := fs.Parse(args); err != nil {
		return 1
	}

	return writeFormatted(ctx, version.Current())
}
