package diff

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestDiffRegressionPathIndependentIdentityReportsRelocationNotModification(t *testing.T) {
	repo := newGitRepoFixture(t)

	from := filepath.Join(repo.Root, "data", "users", "user-bob.yaml")
	to := filepath.Join(repo.Root, "data", "users", "bob-renamed.yaml")
	if err := os.Rename(from, to); err != nil {
		t.Fatalf("rename data file: %v", err)
	}

	result := mustDiffFromSnapshots(t, repo.Root, headSnapshot(), workingTreeSnapshot(WorkingTreeViewFull))

	want := []string{"relocated:User:User-Bob"}
	if got := diffEntryKeys(result); !reflect.DeepEqual(got, want) {
		t.Fatalf("expected path-only move to stay relocation-only, got %v", got)
	}
}

func TestDiffRegressionSplitPreservesStableComparisonFacts(t *testing.T) {
	repo := newGitRepoFixture(t)

	if err := os.Remove(filepath.Join(repo.Root, "data", "posts", "posts.yaml")); err != nil {
		t.Fatalf("remove original posts file: %v", err)
	}
	repo.WriteDataChange(t, "data/posts/post-001.yaml", "id: Post-001\ntitle: First Post\nauthor: User-Alice\ntags:\n  - Tag-Writing\n  - Tag-Product\nbody: Hello world\n")
	repo.WriteDataChange(t, "data/posts/post-002.yaml", "id: Post-002\ntitle: Second Post\nauthor: User-Alice\ntags:\n  - Tag-Writing\nbody: Another post\n")

	result := mustDiffFromSnapshots(t, repo.Root, headSnapshot(), workingTreeSnapshot(WorkingTreeViewFull))

	want := []string{
		"relocated:Post:Post-001",
		"relocated:Post:Post-002",
	}
	if got := diffEntryKeys(result); !reflect.DeepEqual(got, want) {
		t.Fatalf("expected split to preserve relocation-only facts, got %v", got)
	}
	for _, entry := range result.Entries {
		if entry.Kind != DiffEntryKindRelocated {
			t.Fatalf("expected split comparison to avoid false modifications, got %+v", entry)
		}
	}
}

func TestDiffRegressionCombinePreservesStableComparisonFacts(t *testing.T) {
	repo := newGitRepoFixture(t)

	if err := os.Remove(filepath.Join(repo.Root, "data", "users", "user-alice.yaml")); err != nil {
		t.Fatalf("remove alice file: %v", err)
	}
	if err := os.Remove(filepath.Join(repo.Root, "data", "users", "user-bob.yaml")); err != nil {
		t.Fatalf("remove bob file: %v", err)
	}
	repo.WriteDataChange(t, "data/users/users.yaml", "items:\n  - id: User-Alice\n    name: Alice Example\n    email: alice@example.com\n    role: admin\n  - id: User-Bob\n    name: Bob Example\n    email: bob@example.com\n    role: editor\n")

	result := mustDiffFromSnapshots(t, repo.Root, headSnapshot(), workingTreeSnapshot(WorkingTreeViewFull))

	want := []string{
		"relocated:User:User-Alice",
		"relocated:User:User-Bob",
	}
	if got := diffEntryKeys(result); !reflect.DeepEqual(got, want) {
		t.Fatalf("expected combine to preserve relocation-only facts, got %v", got)
	}
	for _, entry := range result.Entries {
		if entry.Kind != DiffEntryKindRelocated {
			t.Fatalf("expected combine comparison to avoid false modifications, got %+v", entry)
		}
	}
}

func mustDiffFromSnapshots(t *testing.T, root string, left, right SnapshotRef) DiffResult {
	t.Helper()

	corpora := mustLoadDiffDataCorpora(t, root, left, right)
	leftDB := mustBuildLogicalDatabase(t, corpora.Left)
	rightDB := mustBuildLogicalDatabase(t, corpora.Right)

	return mustDiffLogicalDatabases(t, leftDB, rightDB)
}
