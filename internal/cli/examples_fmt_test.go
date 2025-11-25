package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"sort"
	"testing"
)

func TestFmtExamplesLint(t *testing.T) {
	for _, root := range exampleRoots(t) {
		root := root
		t.Run(filepath.Base(root), func(t *testing.T) {
			t.Helper()
			stdout := &bytes.Buffer{}
			stderr := &bytes.Buffer{}

			code := Run([]string{"--root", root, "fmt", "--lint"}, stdout, stderr)
			if code != 0 {
				t.Fatalf("fmt --lint exit %d stdout %s stderr %s", code, stdout.String(), stderr.String())
			}
			if stdout.Len() != 0 {
				t.Fatalf("expected no formatting diffs, got %s", stdout.String())
			}
		})
	}
}

func exampleRoots(t *testing.T) []string {
	t.Helper()
	base := filepath.Join("..", "..", "examples")
	abs, err := filepath.Abs(base)
	if err != nil {
		t.Fatalf("abs path: %v", err)
	}
	entries, err := os.ReadDir(abs)
	if err != nil {
		t.Fatalf("readdir %s: %v", abs, err)
	}
	var roots []string
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		dir := filepath.Join(abs, entry.Name())
		cfgPath := filepath.Join(dir, "mergeway.yaml")
		if _, err := os.Stat(cfgPath); err != nil {
			continue
		}
		roots = append(roots, dir)
	}
	if len(roots) == 0 {
		t.Fatalf("no example directories with mergeway.yaml under %s", abs)
	}
	sort.Strings(roots)
	return roots
}
