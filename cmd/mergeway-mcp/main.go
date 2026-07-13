package main

import (
	"context"
	"os"

	mergewaymcp "github.com/mergewayhq/mergeway-cli/internal/mcp"
	"github.com/mergewayhq/mergeway-cli/internal/mcpcmd"
)

func main() {
	os.Exit(mcpcmd.Run(context.Background(), os.Args[1:], os.Stdin, os.Stdout, os.Stderr, mcpcmd.Options{
		Start: func(ctx context.Context, inv mcpcmd.Invocation) error {
			service, err := mergewaymcp.NewService(inv.Root, inv.Entities)
			if err != nil {
				return err
			}
			return mergewaymcp.Run(ctx, mergewaymcp.RunOptions{
				Service:      service,
				Transport:    inv.Transport,
				Stdin:        inv.Stdin,
				Stdout:       inv.Stdout,
				HTTPListen:   inv.HTTPListen,
				HTTPBasePath: inv.HTTPBasePath,
			})
		},
	}))
}
