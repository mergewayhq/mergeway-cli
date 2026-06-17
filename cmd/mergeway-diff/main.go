package main

import (
	"os"

	"github.com/mergewayhq/mergeway-cli/internal/diffcmd"
)

func main() {
	code := diffcmd.Run(os.Args[1:], os.Stdout, os.Stderr)
	if code != 0 {
		os.Exit(code)
	}
}
