package diff

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestResolveDiffSnapshotsZeroArgs(t *testing.T) {
	repo := newGitRepoFixture(t)

	got, err := resolveDiffSnapshots(repo.Root, nil)
	if err != nil {
		t.Fatalf("resolve diff snapshots: %v", err)
	}

	if got.Left.Kind != SnapshotKindHead {
		t.Fatalf("expected left snapshot kind %q, got %q", SnapshotKindHead, got.Left.Kind)
	}
	if got.Right.Kind != SnapshotKindWorkingTree || got.Right.WorkingTreeView != WorkingTreeViewUnstaged {
		t.Fatalf("expected right snapshot to be unstaged working tree, got %+v", got.Right)
	}
}

func TestResolveDiffSnapshotsOneArg(t *testing.T) {
	repo := newGitRepoFixture(t)
	left := repo.Revision(t, "HEAD")

	got, err := resolveDiffSnapshots(repo.Root, []string{left})
	if err != nil {
		t.Fatalf("resolve diff snapshots: %v", err)
	}

	if got.Left.Kind != SnapshotKindRevision || got.Left.Revision != left {
		t.Fatalf("expected left revision %q, got %+v", left, got.Left)
	}
	if got.Right.Kind != SnapshotKindWorkingTree || got.Right.WorkingTreeView != WorkingTreeViewFull {
		t.Fatalf("expected full working tree right snapshot, got %+v", got.Right)
	}
}

func TestResolveDiffSnapshotsTwoArgs(t *testing.T) {
	repo := newGitRepoFixture(t)
	left := repo.Revision(t, "HEAD")
	right := repo.CommitDataChange(t, "data/users/user-bob.yaml", "id: User-Bob\nname: Bob Changed\nemail: bob@example.com\n")

	got, err := resolveDiffSnapshots(repo.Root, []string{left, right})
	if err != nil {
		t.Fatalf("resolve diff snapshots: %v", err)
	}

	if got.Left.Kind != SnapshotKindRevision || got.Left.Revision != left {
		t.Fatalf("expected left revision %q, got %+v", left, got.Left)
	}
	if got.Right.Kind != SnapshotKindRevision || got.Right.Revision != right {
		t.Fatalf("expected right revision %q, got %+v", right, got.Right)
	}
}

func TestResolveDiffSnapshotsInvalidRevision(t *testing.T) {
	repo := newGitRepoFixture(t)

	_, err := resolveDiffSnapshots(repo.Root, []string{"does-not-exist"})
	if err == nil {
		t.Fatalf("expected invalid revision to fail")
	}
	if !strings.Contains(err.Error(), `invalid revision "does-not-exist"`) {
		t.Fatalf("expected clean invalid revision error, got %v", err)
	}
}

func TestGitRepoFixtureSupportsCommittedStagedAndUnstagedChanges(t *testing.T) {
	repo := newGitRepoFixture(t)
	initial := repo.Revision(t, "HEAD")
	committed := repo.CommitDataChange(t, "data/users/user-alice.yaml", "id: User-Alice\nname: Alice Example\nemail: alice+committed@example.com\n")
	if committed == initial {
		t.Fatalf("expected committed change to advance HEAD")
	}
	if status := repo.StatusShort(t); status != "" {
		t.Fatalf("expected clean status after commit, got %q", status)
	}

	repo.StageDataChange(t, "data/users/user-bob.yaml", "id: User-Bob\nname: Bob Example\nemail: bob+staged@example.com\n")
	status := repo.StatusShort(t)
	if !strings.Contains(status, "M  data/users/user-bob.yaml") {
		t.Fatalf("expected staged change in status, got %q", status)
	}

	repo.WriteDataChange(t, "data/tags/tag-product.yaml", "id: Tag-Product\nlabel: Product Updated\n")
	status = repo.StatusShort(t)
	if !strings.Contains(status, " M data/tags/tag-product.yaml") {
		t.Fatalf("expected unstaged change in status, got %q", status)
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

func (r gitRepoFixture) StatusShort(t *testing.T) string {
	t.Helper()
	return strings.TrimRight(runGitCommand(t, r.Root, "status", "--short"), "\n")
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
