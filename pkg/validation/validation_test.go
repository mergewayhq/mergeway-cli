package validation_test

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	pkgconfig "github.com/mergewayhq/mergeway-cli/pkg/config"
	"github.com/mergewayhq/mergeway-cli/pkg/validation"
)

func TestValidateFixture(t *testing.T) {
	repo := copyFixture(t, "repo")
	cfg, err := pkgconfig.Load(filepath.Join(repo, "mergeway.yaml"))
	if err != nil {
		t.Fatalf("load config: %v", err)
	}

	result, err := validation.Validate(repo, cfg, validation.Options{})
	if err != nil {
		t.Fatalf("Validate: %v", err)
	}
	if len(result.Errors) != 0 {
		t.Fatalf("expected no validation errors, got %v", result.Errors)
	}
}

func copyFixture(t *testing.T, name string) string {
	t.Helper()
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	base := filepath.Dir(filename)
	src := filepath.Join(base, "..", "..", "internal", "data", "testdata", name)
	dest := t.TempDir()

	if err := filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dest, rel)

		if info.IsDir() {
			return os.MkdirAll(target, info.Mode())
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(target, data, info.Mode())
	}); err != nil {
		t.Fatalf("copy fixture: %v", err)
	}
	return dest
}
