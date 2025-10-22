package validation

import (
	"path/filepath"
	"testing"

	"github.com/mergewayhq/mergeway-cli/internal/config"
)

func TestValidateExamplesDataset(t *testing.T) {
	root := rootExamplesDir(t)
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
}

func rootExamplesDir(t *testing.T) string {
	path := filepath.Join("..", "..", "examples", "full")
	abs, err := filepath.Abs(path)
	if err != nil {
		t.Fatalf("abs path: %v", err)
	}
	return abs
}
