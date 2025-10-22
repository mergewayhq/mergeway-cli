package cli

import (
	"flag"
	"fmt"

	"github.com/mergewayhq/mergeway-cli/internal/validation"
)

func cmdValidate(ctx *Context, args []string) int {
	fs := flag.NewFlagSet("validate", flag.ContinueOnError)
	fs.SetOutput(ctx.Stderr)
	phaseFlags := multiFlag{}
	fs.Var(&phaseFlags, "phase", "Validation phase to run (format|schema|references), repeatable")
	failFast := fs.Bool("fail-fast", ctx.FailFast, "Stop on first error")
	if err := fs.Parse(args); err != nil {
		return 1
	}

	cfg, err := loadConfig(ctx)
	if err != nil {
		_, _ = fmt.Fprintf(ctx.Stderr, "validate: %v\n", err)
		return 1
	}

	opts := validation.Options{
		FailFast: *failFast,
		Phases:   phaseFlags.Values,
	}

	result, err := validation.Validate(ctx.Root, cfg, opts)
	if err != nil {
		_, _ = fmt.Fprintf(ctx.Stderr, "validate: %v\n", err)
		return 1
	}

	if len(result.Errors) == 0 {
		_, _ = fmt.Fprintln(ctx.Stdout, "validation succeeded")
		return 0
	}

	return writeFormatted(ctx, result.Errors)
}
