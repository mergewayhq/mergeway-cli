package diff

import (
	"bytes"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"
)

func TestLoadDiffDataCorporaIgnoresConfigurationOnlyChanges(t *testing.T) {
	repo := newGitRepoFixture(t)
	appendLine(t, filepath.Join(repo.Root, "types", "User.yaml"), "      nickname:\n        type: string\n")

	corpora := mustLoadDiffDataCorpora(t, repo.Root, headSnapshot(), workingTreeSnapshot(WorkingTreeViewFull))

	if changed := changedCorpusPaths(corpora); len(changed) != 0 {
		t.Fatalf("expected no data changes for config-only edit, got %v", changed)
	}
	for _, path := range corpora.Paths {
		if strings.HasPrefix(path, "types/") || path == "mergeway.yaml" {
			t.Fatalf("expected config paths to be excluded, got %s", path)
		}
	}
}

func TestLoadSnapshotDataIncludePatterns(t *testing.T) {
	repo := newGitRepoFixture(t)

	patterns, err := loadSnapshotDataIncludePatterns(repo.Root, filepath.Join(repo.Root, "mergeway.yaml"), headSnapshot())
	if err != nil {
		t.Fatalf("load snapshot data include patterns: %v", err)
	}

	expected := []string{
		"data/posts/*.yaml",
		"data/tags/*.yaml",
		"data/users/*.yaml",
	}
	if !reflect.DeepEqual(patterns, expected) {
		t.Fatalf("expected patterns %v, got %v", expected, patterns)
	}
}

func TestLoadDiffDataCorporaLoadsDataOnlyChanges(t *testing.T) {
	repo := newGitRepoFixture(t)
	repo.WriteDataChange(t, "data/users/user-alice.yaml", "id: User-Alice\nname: Alice Changed\nemail: alice@example.com\n")

	corpora := mustLoadDiffDataCorpora(t, repo.Root, headSnapshot(), workingTreeSnapshot(WorkingTreeViewFull))
	changed := changedCorpusPaths(corpora)

	expected := []string{"data/users/user-alice.yaml"}
	if !reflect.DeepEqual(changed, expected) {
		t.Fatalf("expected changed paths %v, got %v", expected, changed)
	}

	rightFile := corpusFile(t, corpora.Right, "data/users/user-alice.yaml")
	if !rightFile.Exists || !bytes.Contains(rightFile.Content, []byte("Alice Changed")) {
		t.Fatalf("expected changed working tree content, got %+v", rightFile)
	}
}

func TestLoadDiffDataCorporaMixedConfigAndDataOnlyIncludesData(t *testing.T) {
	repo := newGitRepoFixture(t)
	appendLine(t, filepath.Join(repo.Root, "types", "User.yaml"), "      nickname:\n        type: string\n")
	repo.WriteDataChange(t, "data/users/user-bob.yaml", "id: User-Bob\nname: Bob Changed\nemail: bob@example.com\n")

	corpora := mustLoadDiffDataCorpora(t, repo.Root, headSnapshot(), workingTreeSnapshot(WorkingTreeViewFull))
	changed := changedCorpusPaths(corpora)

	expected := []string{"data/users/user-bob.yaml"}
	if !reflect.DeepEqual(changed, expected) {
		t.Fatalf("expected only data path to change, got %v", changed)
	}
}

func TestLoadDiffDataCorporaRepresentsDeletedDataFiles(t *testing.T) {
	repo := newGitRepoFixture(t)
	target := filepath.Join(repo.Root, "data", "tags", "tag-product.yaml")
	if err := os.Remove(target); err != nil {
		t.Fatalf("remove data file: %v", err)
	}

	corpora := mustLoadDiffDataCorpora(t, repo.Root, headSnapshot(), workingTreeSnapshot(WorkingTreeViewFull))
	leftFile := corpusFile(t, corpora.Left, "data/tags/tag-product.yaml")
	rightFile := corpusFile(t, corpora.Right, "data/tags/tag-product.yaml")

	if !leftFile.Exists {
		t.Fatalf("expected deleted file to exist on left snapshot")
	}
	if rightFile.Exists {
		t.Fatalf("expected deleted file to be absent on right snapshot")
	}
}

func TestLoadDiffDataCorporaRepresentsNewDataFiles(t *testing.T) {
	repo := newGitRepoFixture(t)
	repo.WriteDataChange(t, "data/tags/tag-new.yaml", "id: Tag-New\nlabel: New Tag\n")

	corpora := mustLoadDiffDataCorpora(t, repo.Root, headSnapshot(), workingTreeSnapshot(WorkingTreeViewFull))
	leftFile := corpusFile(t, corpora.Left, "data/tags/tag-new.yaml")
	rightFile := corpusFile(t, corpora.Right, "data/tags/tag-new.yaml")

	if leftFile.Exists {
		t.Fatalf("expected new file to be absent on left snapshot")
	}
	if !rightFile.Exists || !bytes.Contains(rightFile.Content, []byte("Tag-New")) {
		t.Fatalf("expected new file content on right snapshot, got %+v", rightFile)
	}
}

func TestLoadDiffDataCorporaCanLoadMovedDataFromBothSides(t *testing.T) {
	repo := newGitRepoFixture(t)
	from := filepath.Join(repo.Root, "data", "users", "user-bob.yaml")
	to := filepath.Join(repo.Root, "data", "users", "bob-renamed.yaml")
	if err := os.Rename(from, to); err != nil {
		t.Fatalf("rename data file: %v", err)
	}

	corpora := mustLoadDiffDataCorpora(t, repo.Root, headSnapshot(), workingTreeSnapshot(WorkingTreeViewFull))
	leftOld := corpusFile(t, corpora.Left, "data/users/user-bob.yaml")
	rightOld := corpusFile(t, corpora.Right, "data/users/user-bob.yaml")
	leftNew := corpusFile(t, corpora.Left, "data/users/bob-renamed.yaml")
	rightNew := corpusFile(t, corpora.Right, "data/users/bob-renamed.yaml")

	if !leftOld.Exists || rightOld.Exists {
		t.Fatalf("expected old path only on left snapshot, left=%+v right=%+v", leftOld, rightOld)
	}
	if leftNew.Exists || !rightNew.Exists {
		t.Fatalf("expected new path only on right snapshot, left=%+v right=%+v", leftNew, rightNew)
	}
	if !bytes.Equal(leftOld.Content, rightNew.Content) {
		t.Fatalf("expected moved file content to be preserved")
	}
}

func TestLoadDiffDataCorporaSupportsCommittedRevisionSnapshots(t *testing.T) {
	repo := newGitRepoFixture(t)
	left := repo.Revision(t, "HEAD")
	right := repo.CommitDataChange(t, "data/users/user-alice.yaml", "id: User-Alice\nname: Alice Committed\nemail: alice@example.com\n")

	corpora := mustLoadDiffDataCorpora(t, repo.Root, revisionSnapshot(left), revisionSnapshot(right))
	changed := changedCorpusPaths(corpora)

	expected := []string{"data/users/user-alice.yaml"}
	if !reflect.DeepEqual(changed, expected) {
		t.Fatalf("expected committed revision change %v, got %v", expected, changed)
	}
}

func TestLoadDiffDataCorporaSupportsUnstagedWorkingTreeView(t *testing.T) {
	repo := newGitRepoFixture(t)
	repo.StageDataChange(t, "data/users/user-bob.yaml", "id: User-Bob\nname: Bob Staged\nemail: bob@example.com\n")
	repo.WriteDataChange(t, "data/tags/tag-product.yaml", "id: Tag-Product\nlabel: Product Unstaged\n")

	corpora := mustLoadDiffDataCorpora(t, repo.Root, headSnapshot(), workingTreeSnapshot(WorkingTreeViewUnstaged))
	changed := changedCorpusPaths(corpora)

	expected := []string{"data/tags/tag-product.yaml"}
	if !reflect.DeepEqual(changed, expected) {
		t.Fatalf("expected only unstaged change %v, got %v", expected, changed)
	}

	bob := corpusFile(t, corpora.Right, "data/users/user-bob.yaml")
	if bytes.Contains(bob.Content, []byte("Bob Staged")) {
		t.Fatalf("expected staged-only change to be excluded from unstaged view")
	}
}

func TestLoadSnapshotDataCorpusSupportsWorkingTreeFullView(t *testing.T) {
	repo := newGitRepoFixture(t)
	repo.StageDataChange(t, "data/users/user-bob.yaml", "id: User-Bob\nname: Bob Staged\nemail: bob@example.com\n")

	corpora := mustLoadDiffDataCorpora(t, repo.Root, headSnapshot(), workingTreeSnapshot(WorkingTreeViewFull))
	changed := changedCorpusPaths(corpora)

	expected := []string{"data/users/user-bob.yaml"}
	if !reflect.DeepEqual(changed, expected) {
		t.Fatalf("expected full working tree to include staged change, got %v", changed)
	}
}

func mustLoadDiffDataCorpora(t *testing.T, root string, left, right SnapshotRef) DiffDataCorpora {
	t.Helper()
	corpora, err := loadDiffDataCorpora(root, filepath.Join(root, "mergeway.yaml"), left, right)
	if err != nil {
		t.Fatalf("load diff data corpora: %v", err)
	}
	return corpora
}

func changedCorpusPaths(corpora DiffDataCorpora) []string {
	var changed []string
	leftByPath := corpusFilesByPath(corpora.Left)
	rightByPath := corpusFilesByPath(corpora.Right)

	paths := append([]string(nil), corpora.Paths...)
	sort.Strings(paths)
	for _, path := range paths {
		left := leftByPath[path]
		right := rightByPath[path]
		if left.Exists != right.Exists || !bytes.Equal(left.Content, right.Content) {
			changed = append(changed, path)
		}
	}

	return changed
}

func corpusFilesByPath(corpus SnapshotDataCorpus) map[string]SnapshotDataFile {
	files := make(map[string]SnapshotDataFile, len(corpus.Files))
	for _, file := range corpus.Files {
		files[file.Path] = file
	}
	return files
}

func corpusFile(t *testing.T, corpus SnapshotDataCorpus, path string) SnapshotDataFile {
	t.Helper()
	for _, file := range corpus.Files {
		if file.Path == path {
			return file
		}
	}
	t.Fatalf("missing corpus file %s", path)
	return SnapshotDataFile{}
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

func headSnapshot() SnapshotRef {
	return SnapshotRef{Kind: SnapshotKindHead}
}

func revisionSnapshot(revision string) SnapshotRef {
	return SnapshotRef{
		Kind:     SnapshotKindRevision,
		Revision: revision,
	}
}

func workingTreeSnapshot(view WorkingTreeView) SnapshotRef {
	return SnapshotRef{
		Kind:            SnapshotKindWorkingTree,
		WorkingTreeView: view,
	}
}
