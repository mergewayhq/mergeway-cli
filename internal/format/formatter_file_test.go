package format

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFormatFileReportsChanges(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "items.yaml")
	body := `items:
  - id: b
    name: Second
  - id: a
    name: First
`
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	res, err := FormatFile(path, nil)
	if err != nil {
		t.Fatalf("FormatFile: %v", err)
	}
	if !res.Changed {
		t.Fatalf("expected change to be detected")
	}
	formatted := string(res.Content)
	if strings.Index(formatted, "id: a") > strings.Index(formatted, "id: b") {
		t.Fatalf("expected sorted output, got:\n%s", formatted)
	}
}

func TestFormatFileMissingPath(t *testing.T) {
	if _, err := FormatFile("does-not-exist.yaml", nil); err == nil {
		t.Fatalf("expected error for missing file")
	}
}
