package validation

import (
	"os"
	"path/filepath"
	"sort"
	"testing"

	"github.com/mergewayhq/mergeway-cli/internal/config"
)

func TestValidateExamplesDataset(t *testing.T) {
	for _, root := range exampleRoots(t) {
		root := root
		t.Run(filepath.Base(root), func(t *testing.T) {
			t.Helper()
			cfgPath := filepath.Join(root, "mergeway.yaml")
			cfg, err := config.Load(cfgPath)
			if err != nil {
				t.Fatalf("load config: %v", err)
			}

			res, err := Validate(root, cfg, Options{})
			if err != nil {
				t.Fatalf("validate: %v", err)
			}

			if len(res.Errors) != 0 {
				t.Fatalf("expected no validation errors, got %v", res.Errors)
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
		t.Fatalf("readdir: %v", err)
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
