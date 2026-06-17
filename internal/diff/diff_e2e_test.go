package diff_test

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	diffcmd "github.com/mergewayhq/mergeway-cli/internal/diffcmd"
)

func TestDiffFailsOutsideGitRepository(t *testing.T) {
	root := t.TempDir()

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	code := diffcmd.Run([]string{"--root", root}, stdout, stderr)
	if code == 0 {
		t.Fatalf("expected diff to fail outside a git repository")
	}
	if !strings.Contains(stderr.String(), "repository state error") {
		t.Fatalf("expected repository state classification, got %s", stderr.String())
	}
	if !strings.Contains(stderr.String(), "not a git repository") {
		t.Fatalf("expected git repository error, got %s", stderr.String())
	}
	if stdout.Len() != 0 {
		t.Fatalf("expected no stdout on repository error, got %q", stdout.String())
	}
}

func TestDiffShowsUnstagedDataChange(t *testing.T) {
	repo := newGitRepoFixture(t)
	repo.WriteDataChange(t, "data/users/user-alice.yaml", "id: User-Alice\nname: Alice Changed\nemail: alice@example.com\nrole: admin\n")

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	code := diffcmd.Run([]string{"--root", repo.Root}, stdout, stderr)
	if code != 0 {
		t.Fatalf("diff exit %d stderr %s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "MODIFIED User[User-Alice]") {
		t.Fatalf("expected unstaged user modification, got %s", stdout.String())
	}
	if !strings.Contains(stdout.String(), "  name: \"Alice Example\" -> \"Alice Changed\"") {
		t.Fatalf("expected changed field in output, got %s", stdout.String())
	}
}

func TestDiffIgnoresConfigOnlyUnstagedChange(t *testing.T) {
	repo := newGitRepoFixture(t)
	appendLine(t, filepath.Join(repo.Root, "types", "User.yaml"), "      nickname:\n        type: string\n")

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	code := diffcmd.Run([]string{"--root", repo.Root}, stdout, stderr)
	if code != 0 {
		t.Fatalf("diff exit %d stderr %s", code, stderr.String())
	}
	if stdout.String() != "No changes.\n" {
		t.Fatalf("expected config-only change to be ignored, got %q", stdout.String())
	}
}

func TestDiffIgnoresStagedOnlyDataChange(t *testing.T) {
	repo := newGitRepoFixture(t)
	repo.StageDataChange(t, "data/users/user-bob.yaml", "id: User-Bob\nname: Bob Staged\nemail: bob@example.com\nrole: editor\n")

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	code := diffcmd.Run([]string{"--root", repo.Root}, stdout, stderr)
	if code != 0 {
		t.Fatalf("diff exit %d stderr %s", code, stderr.String())
	}
	if stdout.String() != "No changes.\n" {
		t.Fatalf("expected staged-only change to be ignored, got %q", stdout.String())
	}
}

func TestDiffRevisionIncludesUnstagedDataChange(t *testing.T) {
	repo := newGitRepoFixture(t)
	left := repo.Revision(t, "HEAD")
	repo.WriteDataChange(t, "data/users/user-alice.yaml", "id: User-Alice\nname: Alice Changed\nemail: alice@example.com\nrole: admin\n")

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	code := diffcmd.Run([]string{"--root", repo.Root, left}, stdout, stderr)
	if code != 0 {
		t.Fatalf("diff exit %d stderr %s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "MODIFIED User[User-Alice]") {
		t.Fatalf("expected revision diff to include unstaged change, got %s", stdout.String())
	}
}

func TestDiffRevisionIncludesStagedAndUnstagedCurrentState(t *testing.T) {
	repo := newGitRepoFixture(t)
	left := repo.Revision(t, "HEAD")
	repo.StageDataChange(t, "data/users/user-bob.yaml", "id: User-Bob\nname: Bob Staged\nemail: bob@example.com\nrole: editor\n")
	repo.WriteDataChange(t, "data/tags/tag-product.yaml", "id: Tag-Product\nlabel: Product Unstaged\n")

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	code := diffcmd.Run([]string{"--root", repo.Root, left}, stdout, stderr)
	if code != 0 {
		t.Fatalf("diff exit %d stderr %s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "MODIFIED User[User-Bob]") {
		t.Fatalf("expected revision diff to include staged current-state change, got %s", stdout.String())
	}
	if !strings.Contains(stdout.String(), "MODIFIED Tag[Tag-Product]") {
		t.Fatalf("expected revision diff to include unstaged current-state change, got %s", stdout.String())
	}
}

func TestDiffRevisionToRevisionIgnoresWorkingTreeChanges(t *testing.T) {
	repo := newGitRepoFixture(t)
	left := repo.Revision(t, "HEAD")
	right := repo.CommitDataChange(t, "data/users/user-alice.yaml", "id: User-Alice\nname: Alice Committed\nemail: alice@example.com\nrole: admin\n")
	repo.WriteDataChange(t, "data/tags/tag-product.yaml", "id: Tag-Product\nlabel: Product Dirty Worktree\n")

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	code := diffcmd.Run([]string{"--root", repo.Root, left, right}, stdout, stderr)
	if code != 0 {
		t.Fatalf("diff exit %d stderr %s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "MODIFIED User[User-Alice]") {
		t.Fatalf("expected committed revision diff to include committed change, got %s", stdout.String())
	}
	if strings.Contains(stdout.String(), "Tag[Tag-Product]") {
		t.Fatalf("expected committed revision diff to ignore working tree changes, got %s", stdout.String())
	}
}

func TestDiffReturnsReadableErrorForUnparseableData(t *testing.T) {
	repo := newGitRepoFixture(t)
	repo.WriteDataChange(t, "data/users/user-bob.yaml", "id: User-Bob\nname: [\n")

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	code := diffcmd.Run([]string{"--root", repo.Root}, stdout, stderr)
	if code == 0 {
		t.Fatalf("expected diff to fail for invalid data")
	}
	if !strings.Contains(stderr.String(), "data error") {
		t.Fatalf("expected data error classification, got %s", stderr.String())
	}
	if !strings.Contains(stderr.String(), "data/users/user-bob.yaml") {
		t.Fatalf("expected path in parse error, got %s", stderr.String())
	}
	if !strings.Contains(stderr.String(), "parse") {
		t.Fatalf("expected parse error text, got %s", stderr.String())
	}
	if stdout.Len() != 0 {
		t.Fatalf("expected no stdout on parse error, got %q", stdout.String())
	}
}

func TestDiffFailsCleanlyForDuplicateLogicalIDs(t *testing.T) {
	repo := newGitRepoFixture(t)
	repo.WriteDataChange(t, "data/users/user-bob-copy.yaml", "id: User-Bob\nname: Bob Duplicate\nemail: bob-duplicate@example.com\nrole: editor\n")

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	code := diffcmd.Run([]string{"--root", repo.Root}, stdout, stderr)
	if code == 0 {
		t.Fatalf("expected diff to fail for duplicate logical ids")
	}
	if !strings.Contains(stderr.String(), "data error") {
		t.Fatalf("expected data error classification, got %s", stderr.String())
	}
	if !strings.Contains(stderr.String(), `id "User-Bob"`) {
		t.Fatalf("expected duplicate object id in error, got %s", stderr.String())
	}
	if stdout.Len() != 0 {
		t.Fatalf("expected no stdout on duplicate id error, got %q", stdout.String())
	}
}

func TestDiffReturnsClearErrorWhenConfigIsMissing(t *testing.T) {
	repo := newGitRepoWithFiles(t, map[string]string{
		"README.md": "fixture without mergeway config\n",
	})

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	code := diffcmd.Run([]string{"--root", repo.Root}, stdout, stderr)
	if code == 0 {
		t.Fatalf("expected diff to fail when config is missing")
	}
	if !strings.Contains(stderr.String(), "repository state error") {
		t.Fatalf("expected repository state classification, got %s", stderr.String())
	}
	if !strings.Contains(stderr.String(), "config file mergeway.yaml not found") {
		t.Fatalf("expected missing config message, got %s", stderr.String())
	}
}

func TestDiffRepositoryWithNoDiffableDataIsEmpty(t *testing.T) {
	repo := newGitRepoWithFiles(t, map[string]string{
		"mergeway.yaml": `mergeway:
  version: 1

entities:
  Note:
    identifier: id
    fields:
      id:
        type: string
    data:
      - id: note-1
`,
	})

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	code := diffcmd.Run([]string{"--root", repo.Root}, stdout, stderr)
	if code != 0 {
		t.Fatalf("diff exit %d stderr %s", code, stderr.String())
	}
	if stdout.String() != "No changes.\n" {
		t.Fatalf("expected no diffable data to produce empty output, got %q", stdout.String())
	}
}

func TestDiffSurfacesWorkingTreeReadFailuresWithoutPanic(t *testing.T) {
	repo := newGitRepoFixture(t)
	target := filepath.Join(repo.Root, "data", "users", "user-alice.yaml")
	if err := os.Remove(target); err != nil {
		t.Fatalf("remove original file: %v", err)
	}
	if err := os.MkdirAll(target, 0o755); err != nil {
		t.Fatalf("replace file with directory: %v", err)
	}

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	code := diffcmd.Run([]string{"--root", repo.Root}, stdout, stderr)
	if code == 0 {
		t.Fatalf("expected diff to fail for unreadable working tree entry")
	}
	if !strings.Contains(stderr.String(), "repository state error") {
		t.Fatalf("expected repository state classification, got %s", stderr.String())
	}
	if !strings.Contains(stderr.String(), "read working tree file data/users/user-alice.yaml") {
		t.Fatalf("expected read failure path in error, got %s", stderr.String())
	}
	if stdout.Len() != 0 {
		t.Fatalf("expected no stdout on read failure, got %q", stdout.String())
	}
}

func TestDiffEmptySemanticDiffUsesClearOutput(t *testing.T) {
	repo := newGitRepoFixture(t)

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	code := diffcmd.Run([]string{"--root", repo.Root, repo.Revision(t, "HEAD")}, stdout, stderr)
	if code != 0 {
		t.Fatalf("diff exit %d stderr %s", code, stderr.String())
	}
	if stdout.String() != "No changes.\n" {
		t.Fatalf("expected clear empty output, got %q", stdout.String())
	}
}

func TestDiffJSONOutputIsStable(t *testing.T) {
	repo := newGitRepoFixture(t)
	repo.WriteDataChange(t, "data/users/user-alice.yaml", "id: User-Alice\nname: Alice Changed\nemail: alice@example.com\nrole: admin\n")

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	code := diffcmd.Run([]string{"--root", repo.Root, "--format", "json"}, stdout, stderr)
	if code != 0 {
		t.Fatalf("diff --format json exit %d stderr %s", code, stderr.String())
	}

	want := `{
  "version": 1,
  "entries": [
    {
      "kind": "modified",
      "type": "User",
      "object_id": "User-Alice",
      "old_value": {
        "email": "alice@example.com",
        "id": "User-Alice",
        "name": "Alice Example",
        "role": "admin"
      },
      "new_value": {
        "email": "alice@example.com",
        "id": "User-Alice",
        "name": "Alice Changed",
        "role": "admin"
      },
      "changes": [
        {
          "path": "name",
          "before": "Alice Example",
          "after": "Alice Changed"
        }
      ],
      "old_sources": [
        {
          "path": "data/users/user-alice.yaml"
        }
      ],
      "new_sources": [
        {
          "path": "data/users/user-alice.yaml"
        }
      ]
    }
  ]
}
`
	if stdout.String() != want {
		t.Fatalf("unexpected diff --format json output\nwant:\n%s\ngot:\n%s", want, stdout.String())
	}
}

type gitRepoFixture struct {
	Root string
}

func newGitRepoFixture(t *testing.T) gitRepoFixture {
	t.Helper()
	root := copyFixture(t)

	return initGitRepoFixture(t, root)
}

func newGitRepoWithFiles(t *testing.T, files map[string]string) gitRepoFixture {
	t.Helper()
	root := t.TempDir()
	for relativePath, content := range files {
		target := filepath.Join(root, filepath.FromSlash(relativePath))
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			t.Fatalf("create parent dir: %v", err)
		}
		if err := os.WriteFile(target, []byte(content), 0o644); err != nil {
			t.Fatalf("write fixture file %s: %v", relativePath, err)
		}
	}

	return initGitRepoFixture(t, root)
}

func initGitRepoFixture(t *testing.T, root string) gitRepoFixture {
	t.Helper()
	runGitCommand(t, root, "init")
	runGitCommand(t, root, "config", "user.name", "Mergeway Tests")
	runGitCommand(t, root, "config", "user.email", "mergeway-tests@example.com")
	runGitCommand(t, root, "add", ".")
	runGitCommand(t, root, "commit", "-m", "initial fixture")

	return gitRepoFixture{Root: root}
}

func (r gitRepoFixture) Revision(t *testing.T, spec string) string {
	t.Helper()
	return strings.TrimSpace(runGitCommand(t, r.Root, "rev-parse", spec))
}

func (r gitRepoFixture) CommitDataChange(t *testing.T, relativePath, content string) string {
	t.Helper()
	r.WriteDataChange(t, relativePath, content)
	runGitCommand(t, r.Root, "add", relativePath)
	runGitCommand(t, r.Root, "commit", "-m", "update "+relativePath)
	return r.Revision(t, "HEAD")
}

func (r gitRepoFixture) StageDataChange(t *testing.T, relativePath, content string) {
	t.Helper()
	r.WriteDataChange(t, relativePath, content)
	runGitCommand(t, r.Root, "add", relativePath)
}

func (r gitRepoFixture) WriteDataChange(t *testing.T, relativePath, content string) {
	t.Helper()
	target := filepath.Join(r.Root, filepath.FromSlash(relativePath))
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		t.Fatalf("create parent dir: %v", err)
	}
	if err := os.WriteFile(target, []byte(content), 0o644); err != nil {
		t.Fatalf("write data change: %v", err)
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

func runGitCommand(t *testing.T, root string, args ...string) string {
	t.Helper()

	cmd := exec.Command("git", append([]string{"-C", root}, args...)...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %s failed: %v\n%s", strings.Join(args, " "), err, string(output))
	}
	return string(output)
}

func appendLine(t *testing.T, path, suffix string) {
	t.Helper()
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read file: %v", err)
	}
	content = append(content, []byte(suffix)...)
	if err := os.WriteFile(path, content, 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}
}
