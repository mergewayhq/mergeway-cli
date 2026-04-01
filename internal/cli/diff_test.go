package cli

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestDiffAcceptsZeroArgs(t *testing.T) {
	repo := newGitRepoFixture(t)
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	code := Run([]string{"--root", repo.Root, "diff"}, stdout, stderr)
	if code != 0 {
		t.Fatalf("expected diff with zero args to succeed, exit %d stderr %s", code, stderr.String())
	}
	if stdout.String() != "No changes.\n" {
		t.Fatalf("expected empty diff output, got %q", stdout.String())
	}
}

func TestDiffAcceptsOneArg(t *testing.T) {
	repo := newGitRepoFixture(t)
	left := repo.Revision(t, "HEAD")
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	code := Run([]string{"--root", repo.Root, "diff", left}, stdout, stderr)
	if code != 0 {
		t.Fatalf("expected diff with one arg to succeed, exit %d stderr %s", code, stderr.String())
	}
	if stdout.String() != "No changes.\n" {
		t.Fatalf("expected empty diff output, got %q", stdout.String())
	}
}

func TestDiffAcceptsTwoArgs(t *testing.T) {
	repo := newGitRepoFixture(t)
	left := repo.Revision(t, "HEAD")
	right := repo.CommitDataChange(t, "data/users/user-bob.yaml", "id: User-Bob\nname: Bob Changed\nemail: bob@example.com\n")
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	code := Run([]string{"--root", repo.Root, "diff", left, right}, stdout, stderr)
	if code != 0 {
		t.Fatalf("expected diff with two args to succeed, exit %d stderr %s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "MODIFIED User[User-Bob]") {
		t.Fatalf("expected revision diff output, got %s", stdout.String())
	}
	if !strings.Contains(stdout.String(), "  name: \"Bob Example\" -> \"Bob Changed\"") {
		t.Fatalf("expected modified name line in diff output, got %s", stdout.String())
	}
	if !strings.Contains(stdout.String(), "  role: \"editor\" -> null") {
		t.Fatalf("expected modified fields in diff output, got %s", stdout.String())
	}
}

func TestDiffRejectsThreeArgs(t *testing.T) {
	repo := newGitRepoFixture(t)
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	code := Run([]string{"--root", repo.Root, "diff", "a", "b", "c"}, stdout, stderr)
	if code == 0 {
		t.Fatalf("expected diff with three args to fail")
	}
	if !strings.Contains(stdout.String(), "Usage:") {
		t.Fatalf("expected usage text in stdout, got %s", stdout.String())
	}
	if !strings.Contains(stderr.String(), "accepts at most 2 snapshot arguments") {
		t.Fatalf("expected arity error in stderr, got %s", stderr.String())
	}
	if !strings.Contains(stderr.String(), "input error") {
		t.Fatalf("expected input error classification, got %s", stderr.String())
	}
}

func TestDiffRejectsInvalidRevision(t *testing.T) {
	repo := newGitRepoFixture(t)
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	code := Run([]string{"--root", repo.Root, "diff", "does-not-exist"}, stdout, stderr)
	if code == 0 {
		t.Fatalf("expected diff with invalid revision to fail")
	}
	if !strings.Contains(stderr.String(), `invalid revision "does-not-exist"`) {
		t.Fatalf("expected clean invalid revision error, got %s", stderr.String())
	}
	if !strings.Contains(stderr.String(), "input error") {
		t.Fatalf("expected input error classification, got %s", stderr.String())
	}
}

func TestDiffHelpMentionsDataOnlyDiffing(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	code := Run([]string{"diff", "--help"}, stdout, stderr)
	if code != 0 {
		t.Fatalf("expected diff help to succeed, exit %d stderr %s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "data-only diff") {
		t.Fatalf("expected help text to mention data-only diffing, got %s", stdout.String())
	}
}

func TestDiffHelpMentionsConfigurationIsExcluded(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	code := Run([]string{"diff", "--help"}, stdout, stderr)
	if code != 0 {
		t.Fatalf("expected diff help to succeed, exit %d stderr %s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "excludes configuration files entirely") {
		t.Fatalf("expected help text to mention configuration exclusion, got %s", stdout.String())
	}
}

type gitRepoFixture struct {
	Root string
}

func newGitRepoFixture(t *testing.T) gitRepoFixture {
	t.Helper()
	root := copyDiffFixture(t)

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
	target := filepath.Join(r.Root, filepath.FromSlash(relativePath))
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		t.Fatalf("create parent dir: %v", err)
	}
	if err := os.WriteFile(target, []byte(content), 0o644); err != nil {
		t.Fatalf("write data change: %v", err)
	}
	runGitCommand(t, r.Root, "add", relativePath)
	runGitCommand(t, r.Root, "commit", "-m", "update "+relativePath)
	return r.Revision(t, "HEAD")
}

func copyDiffFixture(t *testing.T) string {
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
