package main

import (
	"os"
	"testing"
)

func TestMainExecutesCLI(t *testing.T) {
	origArgs := os.Args
	defer func() { os.Args = origArgs }()

	os.Args = []string{"mw", "version"}
	main()
}
