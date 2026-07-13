package main

import (
	"context"
	"fmt"
	"os"

	"github.com/mergewayhq/mergeway-cli/internal/mcpcmd"
)

func main() {
	os.Exit(mcpcmd.Run(context.Background(), os.Args[1:], os.Stdin, os.Stdout, os.Stderr, mcpcmd.Options{
		Start: func(_ context.Context, _ mcpcmd.Invocation) error {
			return fmt.Errorf("server implementation not yet available")
		},
	}))
}
