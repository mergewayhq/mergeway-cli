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

func TestPersistentFlagsAfterSubcommand(t *testing.T) {
	repo := copyFixture(t)
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	code := Run([]string{"entity", "list", "--root", repo}, stdout, stderr)
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
	if strings.Contains(stdout.String(), "\"$path\"") {
		t.Fatalf("expected no implicit path metadata, got %s", stdout.String())
	}
}

func TestListAndGetParentTypeIncludeDescendants(t *testing.T) {
	repo := cliInheritanceRepo(t)
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	code := Run([]string{"--root", repo, "list", "--type", "Animal"}, stdout, stderr)
	if code != 0 {
		t.Fatalf("list exit %d stderr %s", code, stderr.String())
	}
	lines := strings.Fields(stdout.String())
	expected := []string{"animal-1", "dog-1"}
	if !reflect.DeepEqual(lines, expected) {
		t.Fatalf("expected IDs %v, got %v", expected, lines)
	}

	stdout.Reset()
	stderr.Reset()
	code = Run([]string{"--root", repo, "--format", "json", "get", "--type", "Animal", "dog-1"}, stdout, stderr)
	if code != 0 {
		t.Fatalf("get exit %d stderr %s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "\"breed\": \"collie\"") {
		t.Fatalf("expected child fields in parent get output, got %s", stdout.String())
	}
	if !strings.Contains(stdout.String(), "\"name\": \"Fido\"") {
		t.Fatalf("expected inherited fields in parent get output, got %s", stdout.String())
	}
}

func TestListFilterSupportsDeclaredPathDerivedFields(t *testing.T) {
	repo := pathSegmentsRepo(t)
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	code := Run([]string{"--root", repo, "list", "--type", "Page", "--filter", "section=guides"}, stdout, stderr)
	if code != 0 {
		t.Fatalf("exit code %d, stderr %s", code, stderr.String())
	}

	lines := strings.Fields(stdout.String())
	expected := []string{"guide-install", "guide-validate"}
	if !reflect.DeepEqual(lines, expected) {
		t.Fatalf("expected filtered IDs %v, got %v", expected, lines)
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

func TestFilesCommand(t *testing.T) {
	repo := copyFixture(t)
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	code := Run([]string{"--root", repo, "files"}, stdout, stderr)
	if code != 0 {
		t.Fatalf("files exit %d stderr %s", code, stderr.String())
	}

	var entries []map[string]string
	if err := yaml.Unmarshal(stdout.Bytes(), &entries); err != nil {
		t.Fatalf("expected yaml output, got parse error: %v\nbody:\n%s", err, stdout.String())
	}

	expected := []map[string]string{
		{"type": "Post", "file": "data/posts/posts.yaml"},
		{"type": "Tag", "file": "data/tags/tag-product.yaml"},
		{"type": "Tag", "file": "data/tags/tag-writing.yaml"},
		{"type": "User", "file": "data/users/user-alice.yaml"},
		{"type": "User", "file": "data/users/user-bob.yaml"},
	}
	if !reflect.DeepEqual(entries, expected) {
		t.Fatalf("expected %v, got %v", expected, entries)
	}
}

func TestFilesCommandFiltersByType(t *testing.T) {
	repo := copyFixture(t)
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	code := Run([]string{"--root", repo, "files", "--type", "Tag"}, stdout, stderr)
	if code != 0 {
		t.Fatalf("files --type exit %d stderr %s", code, stderr.String())
	}

	var entries []map[string]string
	if err := yaml.Unmarshal(stdout.Bytes(), &entries); err != nil {
		t.Fatalf("expected yaml output, got parse error: %v\nbody:\n%s", err, stdout.String())
	}

	expected := []map[string]string{
		{"type": "Tag", "file": "data/tags/tag-product.yaml"},
		{"type": "Tag", "file": "data/tags/tag-writing.yaml"},
	}
	if !reflect.DeepEqual(entries, expected) {
		t.Fatalf("expected %v, got %v", expected, entries)
	}
}

func TestFilesCommandFormatsJSON(t *testing.T) {
	repo := copyFixture(t)
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	code := Run([]string{"--root", repo, "--format", "json", "files"}, stdout, stderr)
	if code != 0 {
		t.Fatalf("files --format json exit %d stderr %s", code, stderr.String())
	}

	var entries []map[string]string
	if err := json.Unmarshal(stdout.Bytes(), &entries); err != nil {
		t.Fatalf("expected json output, got parse error: %v\nbody:\n%s", err, stdout.String())
	}
	if len(entries) != 5 {
		t.Fatalf("expected 5 entries, got %d (%v)", len(entries), entries)
	}
	if entries[0]["type"] != "Post" || entries[0]["file"] != "data/posts/posts.yaml" {
		t.Fatalf("unexpected first entry: %v", entries[0])
	}
}

func TestFilesCommandGroupsContainers(t *testing.T) {
	repo := copyFixture(t)
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	code := Run([]string{"--root", repo, "--format", "json", "files", "--group"}, stdout, stderr)
	if code != 0 {
		t.Fatalf("files --group exit %d stderr %s", code, stderr.String())
	}

	var entries []map[string]string
	if err := json.Unmarshal(stdout.Bytes(), &entries); err != nil {
		t.Fatalf("expected json output, got parse error: %v\nbody:\n%s", err, stdout.String())
	}

	expected := []map[string]string{
		{"type": "Post", "file": "data/posts/posts.yaml"},
		{"type": "Tag", "file": "data/tags/*.yaml"},
		{"type": "User", "file": "data/users/*.yaml"},
	}
	if !reflect.DeepEqual(entries, expected) {
		t.Fatalf("expected %v, got %v", expected, entries)
	}
}

func TestFilesCommandGroupsContainersWithRelativeRoot(t *testing.T) {
	repo := copyFixture(t)
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	relativeRoot, err := filepath.Rel(cwd, repo)
	if err != nil {
		t.Fatalf("relative root: %v", err)
	}

	withWorkingDir(t, cwd, func() {
		stdout := &bytes.Buffer{}
		stderr := &bytes.Buffer{}

		code := Run([]string{"--root", relativeRoot, "--format", "json", "files", "--group"}, stdout, stderr)
		if code != 0 {
			t.Fatalf("files --group with relative root exit %d stderr %s", code, stderr.String())
		}

		var entries []map[string]string
		if err := json.Unmarshal(stdout.Bytes(), &entries); err != nil {
			t.Fatalf("expected json output, got parse error: %v\nbody:\n%s", err, stdout.String())
		}

		expected := []map[string]string{
			{"type": "Post", "file": "data/posts/posts.yaml"},
			{"type": "Tag", "file": "data/tags/*.yaml"},
			{"type": "User", "file": "data/users/*.yaml"},
		}
		if !reflect.DeepEqual(entries, expected) {
			t.Fatalf("expected %v, got %v", expected, entries)
		}
	})
}

func TestFilesCommandRejectsUnknownType(t *testing.T) {
	repo := copyFixture(t)
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	code := Run([]string{"--root", repo, "files", "--type", "Missing"}, stdout, stderr)
	if code == 0 {
		t.Fatalf("expected files unknown type to fail, stdout %s stderr %s", stdout.String(), stderr.String())
	}
	if !strings.Contains(stderr.String(), "unknown type Missing") {
		t.Fatalf("expected unknown type error, got %s", stderr.String())
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

	if stdout.String() != "status: validation succeeded\n" {
		t.Fatalf("unexpected output: %s", stdout.String())
	}
}

func TestValidateCommandWithRelativeRoot(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd: %v", err)
	}
	repoRoot, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatalf("Abs(repo root): %v", err)
	}
	if err := os.Chdir(repoRoot); err != nil {
		t.Fatalf("Chdir(repo root): %v", err)
	}
	t.Cleanup(func() {
		if chdirErr := os.Chdir(cwd); chdirErr != nil {
			t.Fatalf("restore cwd: %v", chdirErr)
		}
	})

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	code := Run([]string{"--root", "examples/full", "validate"}, stdout, stderr)
	if code != 0 {
		t.Fatalf("validate with relative root exit %d stderr %s", code, stderr.String())
	}

	if stdout.String() != "status: validation succeeded\n" {
		t.Fatalf("unexpected output: %s", stdout.String())
	}
}

func TestValidateCommandFormatsSuccessAsJSON(t *testing.T) {
	repo := copyFixture(t)
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	code := Run([]string{"--root", repo, "--format", "json", "validate"}, stdout, stderr)
	if code != 0 {
		t.Fatalf("validate --format json exit %d stderr %s", code, stderr.String())
	}

	var payload map[string]string
	if err := json.Unmarshal(stdout.Bytes(), &payload); err != nil {
		t.Fatalf("expected json output, got parse error: %v\nbody:\n%s", err, stdout.String())
	}
	if payload["status"] != "validation succeeded" {
		t.Fatalf("expected success status, got %q", payload["status"])
	}
}

func TestValidateCommandReturnsNonZeroOnValidationErrors(t *testing.T) {
	repo := copyFixture(t)
	target := filepath.Join(repo, "data", "posts", "one.yaml")
	content := `type: Post
items:
  - id: Post-1
    title: First Post
    author: User-Missing
    tags:
      - Tag-One
`
	if err := os.WriteFile(target, []byte(content), 0o644); err != nil {
		t.Fatalf("write invalid post: %v", err)
	}

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	code := Run([]string{"--root", repo, "validate"}, stdout, stderr)
	if code == 0 {
		t.Fatalf("expected validate to fail, stdout %s stderr %s", stdout.String(), stderr.String())
	}

	if !strings.Contains(stdout.String(), "references missing User") {
		t.Fatalf("expected validation error in stdout, got %s", stdout.String())
	}
}

func TestValidateCommandFormatsErrorsAsJSON(t *testing.T) {
	repo := copyFixture(t)
	target := filepath.Join(repo, "data", "posts", "one.yaml")
	content := `type: Post
items:
  - id: Post-1
    title: First Post
    author: User-Missing
    tags:
      - Tag-One
`
	if err := os.WriteFile(target, []byte(content), 0o644); err != nil {
		t.Fatalf("write invalid post: %v", err)
	}

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	code := Run([]string{"--root", repo, "--format", "json", "validate"}, stdout, stderr)
	if code == 0 {
		t.Fatalf("expected validate --format json to fail, stdout %s stderr %s", stdout.String(), stderr.String())
	}

	var errs []map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &errs); err != nil {
		t.Fatalf("expected json output, got parse error: %v\nbody:\n%s", err, stdout.String())
	}
	if len(errs) == 0 {
		t.Fatalf("expected validation errors, got %s", stdout.String())
	}
	if errs[0]["Phase"] != "references" {
		t.Fatalf("expected reference phase error, got %v", errs[0]["Phase"])
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

func TestConfigExportRejectsReferenceUnion(t *testing.T) {
	repo := t.TempDir()
	cfg := []byte(`mergeway:
  version: 1

entities:
  User:
    identifier: id
    data:
      - id: user-1
    fields:
      id: string
  Team:
    identifier: id
    data:
      - id: team-1
    fields:
      id: string
  Activity:
    identifier: id
    data:
      - id: activity-1
        owner: user-1
    fields:
      id: string
      owner:
        type: User | Team
`)
	if err := os.WriteFile(filepath.Join(repo, "mergeway.yaml"), cfg, 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	code := Run([]string{"--root", repo, "config", "export", "--type", "Activity"}, stdout, stderr)
	if code == 0 {
		t.Fatalf("expected export to fail")
	}
	if !strings.Contains(stderr.String(), "cannot be exported as JSON Schema") {
		t.Fatalf("expected union export error, got %s", stderr.String())
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

func TestPathIdentifierLifecycleCommands(t *testing.T) {
	repo := pathIdentifierRepo(t)

	listOut := &bytes.Buffer{}
	listErr := &bytes.Buffer{}
	code := Run([]string{"--root", repo, "list", "--type", "Note"}, listOut, listErr)
	if code != 0 {
		t.Fatalf("list exit %d stderr %s", code, listErr.String())
	}
	if !strings.Contains(listOut.String(), "data/notes/alpha.yaml") {
		t.Fatalf("expected path identifier in list output, got %s", listOut.String())
	}

	getOut := &bytes.Buffer{}
	getErr := &bytes.Buffer{}
	code = Run([]string{"--root", repo, "get", "--type", "Note", "data/notes/alpha.yaml"}, getOut, getErr)
	if code != 0 {
		t.Fatalf("get exit %d stderr %s", code, getErr.String())
	}
	if !strings.Contains(getOut.String(), "title: Alpha") {
		t.Fatalf("expected note payload, got %s", getOut.String())
	}
	if strings.Contains(getOut.String(), "$path:") || strings.Contains(getOut.String(), "$path_segment") {
		t.Fatalf("expected no implicit path metadata, got %s", getOut.String())
	}

	payload := filepath.Join(t.TempDir(), "note.yaml")
	if err := os.WriteFile(payload, []byte("title: Beta\n"), 0o644); err != nil {
		t.Fatalf("write payload: %v", err)
	}

	createOut := &bytes.Buffer{}
	createErr := &bytes.Buffer{}
	code = Run([]string{"--root", repo, "create", "--type", "Note", "--file", payload, "--id", "data/notes/beta.yaml"}, createOut, createErr)
	if code != 0 {
		t.Fatalf("create exit %d stderr %s", code, createErr.String())
	}
	createdBody, err := os.ReadFile(filepath.Join(repo, "data", "notes", "beta.yaml"))
	if err != nil {
		t.Fatalf("read created note: %v", err)
	}
	if strings.Contains(string(createdBody), "$path") {
		t.Fatalf("expected created note to omit $path, got %s", string(createdBody))
	}

	updatePayload := filepath.Join(t.TempDir(), "note-update.yaml")
	if err := os.WriteFile(updatePayload, []byte("title: Alpha Updated\n"), 0o644); err != nil {
		t.Fatalf("write update payload: %v", err)
	}

	updateOut := &bytes.Buffer{}
	updateErr := &bytes.Buffer{}
	code = Run([]string{"--root", repo, "update", "--type", "Note", "--id", "data/notes/alpha.yaml", "--file", updatePayload, "--merge"}, updateOut, updateErr)
	if code != 0 {
		t.Fatalf("update exit %d stderr %s", code, updateErr.String())
	}
	updatedBody, err := os.ReadFile(filepath.Join(repo, "data", "notes", "alpha.yaml"))
	if err != nil {
		t.Fatalf("read updated note: %v", err)
	}
	if !strings.Contains(string(updatedBody), "Alpha Updated") {
		t.Fatalf("expected updated note body, got %s", string(updatedBody))
	}

	deleteOut := &bytes.Buffer{}
	deleteErr := &bytes.Buffer{}
	code = Run([]string{"--root", repo, "--yes", "delete", "--type", "Note", "data/notes/beta.yaml"}, deleteOut, deleteErr)
	if code != 0 {
		t.Fatalf("delete exit %d stderr %s", code, deleteErr.String())
	}
	if _, err := os.Stat(filepath.Join(repo, "data", "notes", "beta.yaml")); !os.IsNotExist(err) {
		t.Fatalf("expected note to be deleted, err=%v", err)
	}
}

func TestCreatePathIdentifierRequiresID(t *testing.T) {
	repo := pathIdentifierRepo(t)
	payload := filepath.Join(t.TempDir(), "note.yaml")
	if err := os.WriteFile(payload, []byte("title: Beta\n"), 0o644); err != nil {
		t.Fatalf("write payload: %v", err)
	}

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	code := Run([]string{"--root", repo, "create", "--type", "Note", "--file", payload}, stdout, stderr)
	if code == 0 {
		t.Fatalf("expected create to require --id for path identifiers")
	}
	if !strings.Contains(stderr.String(), "--id is required") {
		t.Fatalf("expected missing --id error, got %s", stderr.String())
	}
}

func TestExternalPathIdentifierReadCommands(t *testing.T) {
	repo := externalRootPathRepo(t)

	listOut := &bytes.Buffer{}
	listErr := &bytes.Buffer{}
	code := Run([]string{"--root", repo, "list", "--type", "Product"}, listOut, listErr)
	if code != 0 {
		t.Fatalf("list exit %d stderr %s", code, listErr.String())
	}
	if !strings.Contains(listOut.String(), "../secondary/products/gadget.yaml") || !strings.Contains(listOut.String(), "../secondary/products/widget.yaml") {
		t.Fatalf("expected external path identifiers in list output, got %s", listOut.String())
	}

	getOut := &bytes.Buffer{}
	getErr := &bytes.Buffer{}
	code = Run([]string{"--root", repo, "get", "--type", "Product", "../secondary/products/widget.yaml"}, getOut, getErr)
	if code != 0 {
		t.Fatalf("get exit %d stderr %s", code, getErr.String())
	}
	if !strings.Contains(getOut.String(), "name: Widget") {
		t.Fatalf("expected product payload, got %s", getOut.String())
	}

	exportOut := &bytes.Buffer{}
	exportErr := &bytes.Buffer{}
	code = Run([]string{"--root", repo, "--format", "json", "export", "Product"}, exportOut, exportErr)
	if code != 0 {
		t.Fatalf("export exit %d stderr %s", code, exportErr.String())
	}

	var payload map[string][]map[string]any
	if err := json.Unmarshal(exportOut.Bytes(), &payload); err != nil {
		t.Fatalf("parse export: %v", err)
	}
	products := payload["Product"]
	if len(products) != 2 {
		t.Fatalf("expected 2 exported products, got %d", len(products))
	}
	if products[0]["$path"] != nil {
		t.Fatalf("expected exported products to omit implicit path metadata, got %v", products[0])
	}
}

func TestDeclaredPathDerivedFieldsAppearInGetAndExport(t *testing.T) {
	repo := pathSegmentsRepo(t)

	getOut := &bytes.Buffer{}
	getErr := &bytes.Buffer{}
	code := Run([]string{"--root", repo, "get", "--type", "Page", "guide-install"}, getOut, getErr)
	if code != 0 {
		t.Fatalf("get exit %d stderr %s", code, getErr.String())
	}
	if !strings.Contains(getOut.String(), "section: guides") || !strings.Contains(getOut.String(), "filename: install.yaml") || !strings.Contains(getOut.String(), "relative_path: data/library/guides/install.yaml") {
		t.Fatalf("expected declared derived fields in get output, got %s", getOut.String())
	}
	if strings.Contains(getOut.String(), "$path_segment") {
		t.Fatalf("expected no implicit path segment fields, got %s", getOut.String())
	}

	exportOut := &bytes.Buffer{}
	exportErr := &bytes.Buffer{}
	code = Run([]string{"--root", repo, "--format", "json", "export", "Page"}, exportOut, exportErr)
	if code != 0 {
		t.Fatalf("export exit %d stderr %s", code, exportErr.String())
	}

	var payload map[string][]map[string]any
	if err := json.Unmarshal(exportOut.Bytes(), &payload); err != nil {
		t.Fatalf("parse export: %v", err)
	}
	pages := payload["Page"]
	if len(pages) == 0 {
		t.Fatalf("expected exported pages, got %v", payload)
	}
	if pages[0]["section"] == nil || pages[0]["filename"] == nil || pages[0]["relative_path"] == nil {
		t.Fatalf("expected declared derived fields in export, got %v", pages[0])
	}
	if pages[0]["$path"] != nil {
		t.Fatalf("expected no implicit path metadata in export, got %v", pages[0])
	}
}

func TestExternalPathIdentifierWriteCommandsRejected(t *testing.T) {
	repo := externalRootPathRepo(t)

	payload := filepath.Join(t.TempDir(), "product.yaml")
	if err := os.WriteFile(payload, []byte("name: New Product\n"), 0o644); err != nil {
		t.Fatalf("write payload: %v", err)
	}

	createOut := &bytes.Buffer{}
	createErr := &bytes.Buffer{}
	code := Run([]string{"--root", repo, "create", "--type", "Product", "--file", payload, "--id", "../secondary/products/new.yaml"}, createOut, createErr)
	if code == 0 {
		t.Fatalf("expected create to reject external path")
	}
	if !strings.Contains(createErr.String(), "must stay within the workspace root") {
		t.Fatalf("expected external create error, got %s", createErr.String())
	}

	updateOut := &bytes.Buffer{}
	updateErr := &bytes.Buffer{}
	code = Run([]string{"--root", repo, "update", "--type", "Product", "--id", "../secondary/products/widget.yaml", "--file", payload, "--merge"}, updateOut, updateErr)
	if code == 0 {
		t.Fatalf("expected update to reject external path")
	}
	if !strings.Contains(updateErr.String(), "outside the workspace root and cannot be modified") {
		t.Fatalf("expected external update error, got %s", updateErr.String())
	}

	deleteOut := &bytes.Buffer{}
	deleteErr := &bytes.Buffer{}
	code = Run([]string{"--root", repo, "--yes", "delete", "--type", "Product", "../secondary/products/widget.yaml"}, deleteOut, deleteErr)
	if code == 0 {
		t.Fatalf("expected delete to reject external path")
	}
	if !strings.Contains(deleteErr.String(), "outside the workspace root and cannot be modified") {
		t.Fatalf("expected external delete error, got %s", deleteErr.String())
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

func TestEntityShowInheritedSchema(t *testing.T) {
	repo := cliInheritanceRepo(t)
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	code := Run([]string{"--root", repo, "entity", "show", "Dog"}, stdout, stderr)
	if code != 0 {
		t.Fatalf("entity show exit %d stderr %s", code, stderr.String())
	}
	out := stdout.String()
	if !strings.Contains(out, "extends: Animal") {
		t.Fatalf("expected extends metadata, got %s", out)
	}
	if !strings.Contains(out, "name:") || !strings.Contains(out, "breed:") {
		t.Fatalf("expected inherited and local fields, got %s", out)
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

func cliInheritanceRepo(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	cfg := `mergeway:
  version: 1

entities:
  Animal:
    identifier: id
    include:
      - data/animals/*.yaml
    fields:
      id: string
      name: string
  Dog:
    extends: Animal
    include:
      - data/dogs/*.yaml
    fields:
      breed: string
`
	if err := os.WriteFile(filepath.Join(root, "mergeway.yaml"), []byte(cfg), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(root, "data", "animals"), 0o755); err != nil {
		t.Fatalf("mkdir animals: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(root, "data", "dogs"), 0o755); err != nil {
		t.Fatalf("mkdir dogs: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "data", "animals", "animal.yaml"), []byte("id: animal-1\nname: Generic\n"), 0o644); err != nil {
		t.Fatalf("write animal: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "data", "dogs", "dog.yaml"), []byte("id: dog-1\nname: Fido\nbreed: collie\n"), 0o644); err != nil {
		t.Fatalf("write dog: %v", err)
	}
	return root
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

func pathIdentifierRepo(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	config := `mergeway:
  version: 1

entities:
  Note:
    identifier: $path
    include:
      - data/notes/*.yaml
    fields:
      title:
        type: string
`
	if err := os.WriteFile(filepath.Join(root, "mergeway.yaml"), []byte(config), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(root, "data", "notes"), 0o755); err != nil {
		t.Fatalf("create data dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "data", "notes", "alpha.yaml"), []byte("title: Alpha\n"), 0o644); err != nil {
		t.Fatalf("write seed note: %v", err)
	}
	return root
}

func pathSegmentsRepo(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	config := `mergeway:
  version: 1

entities:
  Page:
    identifier: slug
    include:
      - data/library/*/*.yaml
    fields:
      slug:
        type: string
      title:
        type: string
      kind:
        type: string
      section:
        type: string
        source:
          path_segment: 2
      filename:
        type: string
        source:
          path_segment_rev: 0
      relative_path:
        type: string
        source:
          path: true
`
	if err := os.WriteFile(filepath.Join(root, "mergeway.yaml"), []byte(config), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(root, "data", "library", "guides"), 0o755); err != nil {
		t.Fatalf("create guides dir: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(root, "data", "library", "reference"), 0o755); err != nil {
		t.Fatalf("create reference dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "data", "library", "guides", "install.yaml"), []byte("slug: guide-install\ntitle: Install Mergeway\nkind: guide\n"), 0o644); err != nil {
		t.Fatalf("write install page: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "data", "library", "guides", "validate.yaml"), []byte("slug: guide-validate\ntitle: Validate a Workspace\nkind: guide\n"), 0o644); err != nil {
		t.Fatalf("write validate page: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "data", "library", "reference", "config.yaml"), []byte("slug: ref-config\ntitle: Configuration Reference\nkind: reference\n"), 0o644); err != nil {
		t.Fatalf("write config page: %v", err)
	}
	return root
}

func externalRootPathRepo(t *testing.T) string {
	t.Helper()
	base := t.TempDir()
	primary := filepath.Join(base, "primary")
	secondary := filepath.Join(base, "secondary")

	if err := os.MkdirAll(filepath.Join(primary, "data", "order-lines"), 0o755); err != nil {
		t.Fatalf("create primary data dir: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(secondary, "products"), 0o755); err != nil {
		t.Fatalf("create secondary data dir: %v", err)
	}

	config := `mergeway:
  version: 1

entities:
  OrderLine:
    identifier: id
    include:
      - data/order-lines/*.yaml
    fields:
      id:
        type: integer
        required: true
      product_id:
        type: Product
        required: true
      quantity:
        type: integer
        required: true

  Product:
    identifier: $path
    include:
      - ../secondary/products/*.yaml
    fields:
      name:
        type: string
        required: true
`
	if err := os.WriteFile(filepath.Join(primary, "mergeway.yaml"), []byte(config), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}
	if err := os.WriteFile(filepath.Join(primary, "data", "order-lines", "line-1001.yaml"), []byte("id: 1001\nproduct_id: ../secondary/products/widget.yaml\nquantity: 2\n"), 0o644); err != nil {
		t.Fatalf("write order line: %v", err)
	}
	if err := os.WriteFile(filepath.Join(secondary, "products", "widget.yaml"), []byte("name: Widget\n"), 0o644); err != nil {
		t.Fatalf("write widget: %v", err)
	}
	if err := os.WriteFile(filepath.Join(secondary, "products", "gadget.yaml"), []byte("name: Gadget\n"), 0o644); err != nil {
		t.Fatalf("write gadget: %v", err)
	}

	return primary
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
