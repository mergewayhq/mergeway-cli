package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
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

func TestListSortsIdentifiers(t *testing.T) {
	repo := copyFixture(t)
	payload := filepath.Join(repo, "data", "users", "user-new.yaml")
	if err := os.WriteFile(payload, []byte("id: User-Zeta\nname: Z\nemail: z@example.com\n"), 0o644); err != nil {
		t.Fatalf("write payload: %v", err)
	}

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	code := Run([]string{"--root", repo, "list", "--type", "User"}, stdout, stderr)
	if code != 0 {
		t.Fatalf("list exit %d stderr %s", code, stderr.String())
	}
	// Should use store.List ordering, so the lexicographically last id prints last.
	out := strings.Fields(stdout.String())
	if len(out) == 0 || out[len(out)-1] != "User-Zeta" {
		t.Fatalf("expected deterministic ordering with new id last, got %v", out)
	}
}

func TestListFilterSortsResults(t *testing.T) {
	repo := copyFixture(t)
	extra := `type: Post
items:
  - id: Post-Z
    title: Late
    author: User-Bob
    tags: []
  - id: Post-A
    title: Early
    author: User-Bob
    tags: []
`
	if err := os.WriteFile(filepath.Join(repo, "data", "posts", "extra.yaml"), []byte(extra), 0o644); err != nil {
		t.Fatalf("write extra posts: %v", err)
	}
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	code := Run([]string{"--root", repo, "list", "--type", "Post", "--filter", "author=User-Bob"}, stdout, stderr)
	if code != 0 {
		t.Fatalf("list --filter exit %d stderr %s", code, stderr.String())
	}
	// Filtered output should still appear sorted even though we had to load objects.
	lines := strings.Fields(stdout.String())
	expected := []string{"Post-A", "Post-Z"}
	if !reflect.DeepEqual(lines, expected) {
		t.Fatalf("expected %v, got %v", expected, lines)
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

func TestFmtCommandInPlace(t *testing.T) {
	repo := copyFixture(t)
	target := filepath.Join(repo, "data", "posts", "posts.yaml")
	content := `type: Post
items:
  - id: post-b
    title: Beta
  - id: post-a
    title: Alpha
`
	if err := os.WriteFile(target, []byte(content), 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	code := Run([]string{"--root", repo, "fmt", "data/posts/posts.yaml"}, stdout, stderr)
	if code != 0 {
		t.Fatalf("fmt exit %d stderr %s", code, stderr.String())
	}
	if stdout.String() != "Formatted data/posts/posts.yaml\n" {
		t.Fatalf("expected formatted file notice, got %q", stdout.String())
	}

	body, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("read formatted file: %v", err)
	}
	formatted := string(body)
	idxA := strings.Index(formatted, "post-a")
	idxB := strings.Index(formatted, "post-b")
	if idxA == -1 || idxB == -1 || idxA > idxB {
		t.Fatalf("expected post-a before post-b, got %s", formatted)
	}
}

func TestFmtCommandStdout(t *testing.T) {
	repo := copyFixture(t)
	target := filepath.Join(repo, "data", "posts", "posts.yaml")
	content := `items:
  - id: post-b
    title: Beta
  - id: post-a
    title: Alpha
`
	if err := os.WriteFile(target, []byte(content), 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	code := Run([]string{"--root", repo, "fmt", "--stdout", "data/posts/posts.yaml"}, stdout, stderr)
	if code != 0 {
		t.Fatalf("fmt exit %d stderr %s", code, stderr.String())
	}
	body := stdout.String()
	idxA := strings.Index(body, "post-a")
	idxB := strings.Index(body, "post-b")
	if idxA == -1 || idxB == -1 || idxA > idxB {
		t.Fatalf("expected post-a before post-b, got %s", body)
	}

	original, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("read original file: %v", err)
	}
	if string(original) != content {
		t.Fatalf("expected file to remain unchanged")
	}
}

func TestFmtCommandLintDefaultsToConfig(t *testing.T) {
	repo := copyFixture(t)
	target := filepath.Join(repo, "data", "posts", "posts.yaml")
	content := `items:
  - id: post-b
    title: Beta
  - id: post-a
    title: Alpha
`
	if err := os.WriteFile(target, []byte(content), 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	code := Run([]string{"--root", repo, "fmt", "--lint"}, stdout, stderr)
	if code != 1 {
		t.Fatalf("fmt --lint exit %d stderr %s", code, stderr.String())
	}
	if stdout.String() != "data/posts/posts.yaml\n" {
		t.Fatalf("expected lint output path, got %q", stdout.String())
	}
}

func TestFmtCommandLintClean(t *testing.T) {
	repo := copyFixture(t)
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	code := Run([]string{"--root", repo, "fmt", "--lint", "data/posts/posts.yaml"}, stdout, stderr)
	if code != 0 {
		t.Fatalf("fmt --lint exit %d stderr %s", code, stderr.String())
	}
	if stdout.Len() != 0 {
		t.Fatalf("expected no lint output, got %s", stdout.String())
	}
}

func TestFmtCommandLintDetectsChanges(t *testing.T) {
	repo := copyFixture(t)
	target := filepath.Join(repo, "data", "posts", "posts.yaml")
	content := `items:
  - id: post-b
    title: Beta
  - id: post-a
    title: Alpha
`
	if err := os.WriteFile(target, []byte(content), 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	code := Run([]string{"--root", repo, "fmt", "--lint", "data/posts/posts.yaml"}, stdout, stderr)
	if code != 1 {
		t.Fatalf("fmt --lint exit %d stderr %s", code, stderr.String())
	}
	if stdout.String() != "data/posts/posts.yaml\n" {
		t.Fatalf("expected lint output path, got %q", stdout.String())
	}

	body, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("read formatted file: %v", err)
	}
	if string(body) != content {
		t.Fatalf("expected file to remain unchanged")
	}
}

func TestFmtCommandOrdersFields(t *testing.T) {
	repo := copyFixture(t)
	target := filepath.Join(repo, "data", "posts", "posts.yaml")
	content := `type: Post
items:
  - title: Beta
    author: User-Alice
    id: post-a
`
	if err := os.WriteFile(target, []byte(content), 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	code := Run([]string{"--root", repo, "fmt", "--in-place", "data/posts/posts.yaml"}, stdout, stderr)
	if code != 0 {
		t.Fatalf("fmt exit %d stderr %s", code, stderr.String())
	}

	body, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("read formatted file: %v", err)
	}
	formatted := string(body)
	idxID := strings.Index(formatted, "id: post-a")
	idxTitle := strings.Index(formatted, "title: Beta")
	idxAuthor := strings.Index(formatted, "author: User-Alice")
	if idxID == -1 || idxTitle == -1 || idxAuthor == -1 || idxID >= idxTitle || idxTitle >= idxAuthor {
		t.Fatalf("expected id -> title -> author order, got:\n%s", formatted)
	}
}

func TestFmtCommandRejectsUntrackedFile(t *testing.T) {
	repo := copyFixture(t)
	target := filepath.Join(repo, "extra.yaml")
	if err := os.WriteFile(target, []byte("id: extra\n"), 0o644); err != nil {
		t.Fatalf("write extra file: %v", err)
	}

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	code := Run([]string{"--root", repo, "fmt", target}, stdout, stderr)
	if code != 1 {
		t.Fatalf("expected fmt to reject untracked file, exit %d stdout %s stderr %s", code, stdout.String(), stderr.String())
	}
	if !strings.Contains(stderr.String(), "not part of the configured data set") {
		t.Fatalf("expected rejection message, got %s", stderr.String())
	}
}

func TestFmtCommandLintInPlaceConflict(t *testing.T) {
	repo := copyFixture(t)
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	code := Run([]string{"--root", repo, "fmt", "--lint", "--in-place", "data/posts/posts.yaml"}, stdout, stderr)
	if code != 1 {
		t.Fatalf("fmt --lint --in-place exit %d stdout %s stderr %s", code, stdout.String(), stderr.String())
	}
	if !strings.Contains(stderr.String(), "--lint cannot be combined with --in-place") {
		t.Fatalf("expected conflict message, got %s", stderr.String())
	}
}

func TestCreateRespectsCustomIdentifier(t *testing.T) {
	repo := customIdentifierRepo(t)
	payload := filepath.Join(t.TempDir(), "payload.yaml")
	if err := os.WriteFile(payload, []byte("name: Gadget\n"), 0o644); err != nil {
		t.Fatalf("write payload: %v", err)
	}

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	code := Run([]string{"--root", repo, "create", "--type", "Gadget", "--file", payload, "--id", "gadget-42"}, stdout, stderr)
	if code != 0 {
		t.Fatalf("create exit %d stderr %s", code, stderr.String())
	}
	expected := filepath.Join(repo, "data", "gadgets", "gadget-42.yaml")
	body, err := os.ReadFile(expected)
	if err != nil {
		t.Fatalf("read created file: %v", err)
	}
	content := string(body)
	if !strings.Contains(content, "slug: gadget-42") {
		t.Fatalf("expected slug field, got %s", content)
	}
}

func TestConfigLintCommand(t *testing.T) {
	repo := copyFixture(t)
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	code := Run([]string{"--root", repo, "config", "lint"}, stdout, stderr)
	if code != 0 {
		t.Fatalf("config lint exit %d stderr %s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "configuration valid") {
		t.Fatalf("expected success message, got %s", stdout.String())
	}
}

func TestConfigLintCommandMissingConfig(t *testing.T) {
	root := t.TempDir()
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	code := Run([]string{"--root", root, "config", "lint"}, stdout, stderr)
	if code == 0 {
		t.Fatalf("expected config lint to fail without config file")
	}
	if !strings.Contains(stderr.String(), "config lint") {
		t.Fatalf("expected lint error message, got %s", stderr.String())
	}
}

func TestEntityShowCommand(t *testing.T) {
	repo := copyFixture(t)
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	code := Run([]string{"--root", repo, "entity", "show", "User"}, stdout, stderr)
	if code != 0 {
		t.Fatalf("entity show exit %d stderr %s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "fields:") {
		t.Fatalf("expected schema output, got %s", stdout.String())
	}
}

func TestEntityShowUnknown(t *testing.T) {
	repo := copyFixture(t)
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	code := Run([]string{"--root", repo, "entity", "show", "Missing"}, stdout, stderr)
	if code == 0 {
		t.Fatalf("expected entity show to fail for unknown type")
	}
	if !strings.Contains(stderr.String(), "unknown entity") {
		t.Fatalf("expected unknown entity message, got %s", stderr.String())
	}
}

func TestInitCommandScaffoldsConfig(t *testing.T) {
	root := t.TempDir()
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	code := Run([]string{"--root", root, "init"}, stdout, stderr)
	if code != 0 {
		t.Fatalf("init exit %d stderr %s", code, stderr.String())
	}

	configPath := filepath.Join(root, "mergeway.yaml")
	body, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("read config: %v", err)
	}
	if !strings.Contains(string(body), "mergeway:") {
		t.Fatalf("expected default config contents, got %s", string(body))
	}
}

func TestInitCommandRejectsArgs(t *testing.T) {
	root := t.TempDir()
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	code := Run([]string{"--root", root, "init", "extra"}, stdout, stderr)
	if code == 0 {
		t.Fatalf("expected init to reject positional args")
	}
	if !strings.Contains(stderr.String(), "no arguments") {
		t.Fatalf("expected rejection message, got %s", stderr.String())
	}
}

func TestUpdateCommandMerge(t *testing.T) {
	repo := copyFixture(t)
	payload := filepath.Join(t.TempDir(), "update.yaml")
	if err := os.WriteFile(payload, []byte("role: maintainer\n"), 0o644); err != nil {
		t.Fatalf("write payload: %v", err)
	}
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	code := Run([]string{"--root", repo, "update", "--type", "User", "--id", "User-Alice", "--file", payload, "--merge"}, stdout, stderr)
	if code != 0 {
		t.Fatalf("update --merge exit %d stderr %s", code, stderr.String())
	}

	body, err := os.ReadFile(filepath.Join(repo, "data", "users", "user-alice.yaml"))
	if err != nil {
		t.Fatalf("read updated user: %v", err)
	}
	content := string(body)
	if !strings.Contains(content, "role: maintainer") {
		t.Fatalf("expected merged field, got %s", content)
	}
	if !strings.Contains(content, "name: Alice Example") {
		t.Fatalf("expected existing fields to remain, got %s", content)
	}
}

func TestUpdateCommandReplace(t *testing.T) {
	repo := copyFixture(t)
	payload := filepath.Join(t.TempDir(), "replace.yaml")
	content := "name: Alicia Example\nemail: alicia@example.com\n"
	if err := os.WriteFile(payload, []byte(content), 0o644); err != nil {
		t.Fatalf("write payload: %v", err)
	}
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	code := Run([]string{"--root", repo, "update", "--type", "User", "--id", "User-Bob", "--file", payload}, stdout, stderr)
	if code != 0 {
		t.Fatalf("update replace exit %d stderr %s", code, stderr.String())
	}

	body, err := os.ReadFile(filepath.Join(repo, "data", "users", "user-bob.yaml"))
	if err != nil {
		t.Fatalf("read updated user: %v", err)
	}
	text := string(body)
	if strings.Contains(text, "role:") {
		t.Fatalf("expected previous fields to be replaced, got %s", text)
	}
	if !strings.Contains(text, "name: Alicia Example") || !strings.Contains(text, "email: alicia@example.com") {
		t.Fatalf("expected replacement fields, got %s", text)
	}
}

func TestDeleteCommandYesSkipsPrompt(t *testing.T) {
	repo := copyFixture(t)
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	code := Run([]string{"--root", repo, "--yes", "delete", "--type", "Tag", "Tag-Writing"}, stdout, stderr)
	if code != 0 {
		t.Fatalf("delete --yes exit %d stderr %s", code, stderr.String())
	}
	if _, err := os.Stat(filepath.Join(repo, "data", "tags", "tag-writing.yaml")); !os.IsNotExist(err) {
		t.Fatalf("expected tag file to be removed, err=%v", err)
	}
}

func TestDeleteCommandAbortWithoutConfirmation(t *testing.T) {
	repo := copyFixture(t)
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	withStdin(t, "n\n", func() {
		code := Run([]string{"--root", repo, "delete", "--type", "Tag", "Tag-Product"}, stdout, stderr)
		if code == 0 {
			t.Fatalf("expected delete to abort without confirmation")
		}
	})

	if !strings.Contains(stderr.String(), "aborted") {
		t.Fatalf("expected abort message, got %s", stderr.String())
	}
	if _, err := os.Stat(filepath.Join(repo, "data", "tags", "tag-product.yaml")); err != nil {
		t.Fatalf("expected tag to remain after abort: %v", err)
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

func customIdentifierRepo(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	config := `mergeway:
  version: 1

entities:
  Gadget:
    identifier:
      field: slug
    include:
      - data/gadgets/*.yaml
    fields:
      slug:
        type: string
        required: true
      name:
        type: string
`
	if err := os.WriteFile(filepath.Join(root, "mergeway.yaml"), []byte(config), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(root, "data", "gadgets"), 0o755); err != nil {
		t.Fatalf("create data dir: %v", err)
	}
	return root
}

func withStdin(t *testing.T, input string, fn func()) {
	t.Helper()
	orig := os.Stdin
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	if _, err := w.Write([]byte(input)); err != nil {
		t.Fatalf("write stdin: %v", err)
	}
	if err := w.Close(); err != nil {
		t.Fatalf("close stdin writer: %v", err)
	}
	os.Stdin = r
	defer func() {
		_ = r.Close()
		os.Stdin = orig
	}()
	fn()
}
