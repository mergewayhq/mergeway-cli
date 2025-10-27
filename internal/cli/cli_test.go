package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestEntityList(t *testing.T) {
	repo := copyFixture(t)
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	code := Run([]string{"--root", repo, "entity", "list"}, stdout, stderr)
	if code != 0 {
		t.Fatalf("exit code %d, stderr %s", code, stderr.String())
	}

	out := stdout.String()
	if !strings.Contains(out, "Post") || !strings.Contains(out, "User") {
		t.Fatalf("unexpected output: %s", out)
	}
}

func TestGet(t *testing.T) {
	repo := copyFixture(t)
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	code := Run([]string{"--root", repo, "--format", "json", "get", "--type", "User", "User-Alice"}, stdout, stderr)
	if code != 0 {
		t.Fatalf("exit code %d, stderr %s", code, stderr.String())
	}

	if !strings.Contains(stdout.String(), "\"name\": \"Alice Example\"") {
		t.Fatalf("unexpected output: %s", stdout.String())
	}
}

func TestCreateAndList(t *testing.T) {
	repo := copyFixture(t)

	payload := filepath.Join(t.TempDir(), "tag.yaml")
	if err := os.WriteFile(payload, []byte("id: Tag-New\nlabel: New Tag\n"), 0o644); err != nil {
		t.Fatalf("write payload: %v", err)
	}

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	code := Run([]string{"--root", repo, "create", "--type", "Tag", "--file", payload}, stdout, stderr)
	if code != 0 {
		t.Fatalf("create exit %d stderr %s", code, stderr.String())
	}

	stdout.Reset()
	stderr.Reset()
	code = Run([]string{"--root", repo, "list", "--type", "Tag"}, stdout, stderr)
	if code != 0 {
		t.Fatalf("list exit %d stderr %s", code, stderr.String())
	}

	if !strings.Contains(stdout.String(), "Tag-New") {
		t.Fatalf("expected new tag in list, got %s", stdout.String())
	}
}

func TestValidateCommand(t *testing.T) {
	repo := copyFixture(t)
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	code := Run([]string{"--root", repo, "validate"}, stdout, stderr)
	if code != 0 {
		t.Fatalf("validate exit %d stderr %s", code, stderr.String())
	}

	if !strings.Contains(stdout.String(), "validation succeeded") {
		t.Fatalf("unexpected output: %s", stdout.String())
	}
}

func TestConfigExport(t *testing.T) {
	repo := copyFixture(t)
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	code := Run([]string{"--root", repo, "--format", "json", "config", "export", "--type", "Post"}, stdout, stderr)
	if code != 0 {
		t.Fatalf("export exit %d stderr %s", code, stderr.String())
	}

	if !strings.Contains(stdout.String(), "\"properties\"") {
		t.Fatalf("unexpected output: %s", stdout.String())
	}
}

func TestExportToStdout(t *testing.T) {
	repo := copyFixture(t)
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	code := Run([]string{"--root", repo, "export"}, stdout, stderr)
	if code != 0 {
		t.Fatalf("export exit %d stderr %s", code, stderr.String())
	}

	data := make(map[string]any)
	if err := yaml.Unmarshal(stdout.Bytes(), &data); err != nil {
		t.Fatalf("unexpected yaml output: %v (body=%s)", err, stdout.String())
	}

	users, ok := data["User"].([]any)
	if !ok || len(users) == 0 {
		t.Fatalf("expected user records, got %v", data["User"])
	}

	if _, ok := data["Post"].([]any); !ok {
		t.Fatalf("expected post records, got %v", data["Post"])
	}
}

func TestExportWithOutputAndFilters(t *testing.T) {
	repo := copyFixture(t)
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	outputPath := filepath.Join(t.TempDir(), "export.json")

	code := Run([]string{"--root", repo, "--format", "json", "export", "--output", outputPath, "Post"}, stdout, stderr)
	if code != 0 {
		t.Fatalf("export exit %d stderr %s", code, stderr.String())
	}

	if stdout.Len() != 0 {
		t.Fatalf("expected no stdout, got %s", stdout.String())
	}

	body, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("read export file: %v", err)
	}

	var payload map[string]any
	if err := json.Unmarshal(body, &payload); err != nil {
		t.Fatalf("unexpected json output: %v (body=%s)", err, string(body))
	}

	if len(payload) != 1 {
		t.Fatalf("expected only one entity, got %v", payload)
	}
	posts, ok := payload["Post"].([]any)
	if !ok || len(posts) == 0 {
		t.Fatalf("expected post records, got %v", payload["Post"])
	}
}

func TestVersionCommand(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	code := Run([]string{"version"}, stdout, stderr)
	if code != 0 {
		t.Fatalf("version exit %d stderr %s", code, stderr.String())
	}

	info := make(map[string]any)
	if err := yaml.Unmarshal(stdout.Bytes(), &info); err != nil {
		t.Fatalf("unexpected yaml output: %v (body=%s)", err, stdout.String())
	}

	versionVal, ok := info["version"].(string)
	if !ok || versionVal == "" {
		t.Fatalf("expected version field, got %v", info)
	}
}

func copyFixture(t *testing.T) string {
	t.Helper()
	src := filepath.Join("..", "data", "testdata", "repo")
	dest := t.TempDir()

	if err := filepath.WalkDir(src, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dest, rel)
		if d.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(target, data, 0o644)
	}); err != nil {
		t.Fatalf("copy fixture: %v", err)
	}

	return dest
}
