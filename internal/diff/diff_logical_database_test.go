package diff

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestBuildLogicalDatabaseNormalizesMovedObjectIdentically(t *testing.T) {
	repo := newGitRepoFixture(t)

	from := filepath.Join(repo.Root, "data", "users", "user-bob.yaml")
	to := filepath.Join(repo.Root, "data", "users", "bob-renamed.yaml")
	if err := os.Rename(from, to); err != nil {
		t.Fatalf("rename data file: %v", err)
	}

	corpora := mustLoadDiffDataCorpora(t, repo.Root, headSnapshot(), workingTreeSnapshot(WorkingTreeViewFull))
	left := mustBuildLogicalDatabase(t, corpora.Left)
	right := mustBuildLogicalDatabase(t, corpora.Right)

	if !reflect.DeepEqual(logicalDatabaseSignature(left), logicalDatabaseSignature(right)) {
		t.Fatalf("expected renamed object layout to preserve logical database equivalence")
	}
}

func TestBuildLogicalDatabaseTracksRelocatedObjectAsSameLogicalRecord(t *testing.T) {
	repo := newGitRepoFixture(t)

	from := filepath.Join(repo.Root, "data", "users", "user-bob.yaml")
	to := filepath.Join(repo.Root, "data", "users", "bob-renamed.yaml")
	if err := os.Rename(from, to); err != nil {
		t.Fatalf("rename data file: %v", err)
	}

	corpora := mustLoadDiffDataCorpora(t, repo.Root, headSnapshot(), workingTreeSnapshot(WorkingTreeViewFull))
	left := mustBuildLogicalDatabase(t, corpora.Left)
	right := mustBuildLogicalDatabase(t, corpora.Right)

	leftBob := logicalObjectByKey(t, left, "User", "User-Bob")
	rightBob := logicalObjectByKey(t, right, "User", "User-Bob")

	if leftBob.Canonical != rightBob.Canonical {
		t.Fatalf("expected moved object to preserve canonical value")
	}
	if leftBob.Sources[0].Path == rightBob.Sources[0].Path {
		t.Fatalf("expected moved object source metadata to reflect different file paths")
	}
}

func TestBuildLogicalDatabasePreservesEquivalenceWhenSplittingFile(t *testing.T) {
	repo := newGitRepoFixture(t)

	if err := os.Remove(filepath.Join(repo.Root, "data", "posts", "posts.yaml")); err != nil {
		t.Fatalf("remove original posts file: %v", err)
	}
	repo.WriteDataChange(t, "data/posts/post-001.yaml", "id: Post-001\ntitle: First Post\nauthor: User-Alice\ntags:\n  - Tag-Writing\n  - Tag-Product\nbody: Hello world\n")
	repo.WriteDataChange(t, "data/posts/post-002.yaml", "id: Post-002\ntitle: Second Post\nauthor: User-Alice\ntags:\n  - Tag-Writing\nbody: Another post\n")

	corpora := mustLoadDiffDataCorpora(t, repo.Root, headSnapshot(), workingTreeSnapshot(WorkingTreeViewFull))
	left := mustBuildLogicalDatabase(t, corpora.Left)
	right := mustBuildLogicalDatabase(t, corpora.Right)

	if !reflect.DeepEqual(logicalDatabaseSignature(left), logicalDatabaseSignature(right)) {
		t.Fatalf("expected splitting one file into two to preserve logical equivalence")
	}
}

func TestBuildLogicalDatabasePreservesEquivalenceWhenCombiningFiles(t *testing.T) {
	repo := newGitRepoFixture(t)

	if err := os.Remove(filepath.Join(repo.Root, "data", "users", "user-alice.yaml")); err != nil {
		t.Fatalf("remove alice file: %v", err)
	}
	if err := os.Remove(filepath.Join(repo.Root, "data", "users", "user-bob.yaml")); err != nil {
		t.Fatalf("remove bob file: %v", err)
	}
	repo.WriteDataChange(t, "data/users/users.yaml", "items:\n  - id: User-Alice\n    name: Alice Example\n    email: alice@example.com\n    role: admin\n  - id: User-Bob\n    name: Bob Example\n    email: bob@example.com\n    role: editor\n")

	corpora := mustLoadDiffDataCorpora(t, repo.Root, headSnapshot(), workingTreeSnapshot(WorkingTreeViewFull))
	left := mustBuildLogicalDatabase(t, corpora.Left)
	right := mustBuildLogicalDatabase(t, corpora.Right)

	if !reflect.DeepEqual(logicalDatabaseSignature(left), logicalDatabaseSignature(right)) {
		t.Fatalf("expected combining files to preserve logical equivalence")
	}
}

func TestBuildLogicalDatabaseIgnoresSerializationOnlyReordering(t *testing.T) {
	repo := newGitRepoFixture(t)

	repo.WriteDataChange(t, "data/posts/posts.yaml", "items:\n  - body: Another post\n    tags:\n      - Tag-Writing\n    author: User-Alice\n    title: Second Post\n    id: Post-002\n  - body: Hello world\n    title: First Post\n    tags:\n      - Tag-Writing\n      - Tag-Product\n    id: Post-001\n    author: User-Alice\n")

	corpora := mustLoadDiffDataCorpora(t, repo.Root, headSnapshot(), workingTreeSnapshot(WorkingTreeViewFull))
	left := mustBuildLogicalDatabase(t, corpora.Left)
	right := mustBuildLogicalDatabase(t, corpora.Right)

	if !reflect.DeepEqual(logicalDatabaseSignature(left), logicalDatabaseSignature(right)) {
		t.Fatalf("expected reorder-only serialization change to preserve logical equivalence")
	}
}

func TestBuildLogicalDatabaseUsesDeterministicNormalizationOrder(t *testing.T) {
	repo := newGitRepoFixture(t)
	repo.WriteDataChange(t, "data/tags/aaa.yaml", "id: Tag-AAA\nlabel: A\n")

	corpora := mustLoadDiffDataCorpora(t, repo.Root, headSnapshot(), workingTreeSnapshot(WorkingTreeViewFull))
	db := mustBuildLogicalDatabase(t, corpora.Right)

	got := logicalObjectKeys(db)
	want := []string{
		"Post:Post-001",
		"Post:Post-002",
		"Tag:Tag-AAA",
		"Tag:Tag-Product",
		"Tag:Tag-Writing",
		"User:User-Alice",
		"User:User-Bob",
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("expected deterministic logical object order %v, got %v", want, got)
	}
}

func TestBuildLogicalDatabaseRejectsIdentityCollisionsClearly(t *testing.T) {
	repo := newGitRepoFixture(t)
	repo.WriteDataChange(t, "data/users/user-bob-copy.yaml", "id: User-Bob\nname: Bob Duplicate\nemail: bob-duplicate@example.com\nrole: editor\n")

	corpora := mustLoadDiffDataCorpora(t, repo.Root, headSnapshot(), workingTreeSnapshot(WorkingTreeViewFull))
	_, err := buildLogicalDatabase(corpora.Right)
	if err == nil {
		t.Fatalf("expected logical database build to fail on duplicate identity")
	}

	var buildErr *LogicalDatabaseBuildError
	if !errors.As(err, &buildErr) {
		t.Fatalf("expected structured build error, got %T", err)
	}
	if buildErr.Kind != LogicalDatabaseErrorIdentityCollision {
		t.Fatalf("expected collision error kind, got %s", buildErr.Kind)
	}
	if buildErr.TypeName != "User" || buildErr.ObjectID != "User-Bob" {
		t.Fatalf("expected duplicate User/User-Bob details, got %+v", buildErr)
	}
	if !strings.Contains(err.Error(), "already defined") {
		t.Fatalf("expected collision error message, got %v", err)
	}
}

func TestBuildLogicalDatabaseReturnsStructuredErrorsForInvalidData(t *testing.T) {
	repo := newGitRepoFixture(t)
	repo.WriteDataChange(t, "data/users/user-bob.yaml", "id: User-Bob\nname: [\n")

	corpora := mustLoadDiffDataCorpora(t, repo.Root, headSnapshot(), workingTreeSnapshot(WorkingTreeViewFull))
	_, err := buildLogicalDatabase(corpora.Right)
	if err == nil {
		t.Fatalf("expected logical database build to fail on invalid data")
	}

	var buildErr *LogicalDatabaseBuildError
	if !errors.As(err, &buildErr) {
		t.Fatalf("expected structured build error, got %T", err)
	}
	if buildErr.Kind != LogicalDatabaseErrorParse {
		t.Fatalf("expected parse error kind, got %s", buildErr.Kind)
	}
	if buildErr.Path != "data/users/user-bob.yaml" || buildErr.TypeName != "User" {
		t.Fatalf("expected invalid path/type in error, got %+v", buildErr)
	}
}

func mustBuildLogicalDatabase(t *testing.T, corpus SnapshotDataCorpus) LogicalDatabase {
	t.Helper()
	db, err := buildLogicalDatabase(corpus)
	if err != nil {
		t.Fatalf("build logical database: %v", err)
	}
	return db
}

func logicalDatabaseSignature(db LogicalDatabase) []string {
	signature := make([]string, 0, len(db.Objects))
	for _, obj := range db.Objects {
		signature = append(signature, obj.Type+"\x00"+obj.ID+"\x00"+obj.Canonical)
	}
	return signature
}

func logicalObjectKeys(db LogicalDatabase) []string {
	keys := make([]string, 0, len(db.Objects))
	for _, obj := range db.Objects {
		keys = append(keys, obj.Type+":"+obj.ID)
	}
	return keys
}

func logicalObjectByKey(t *testing.T, db LogicalDatabase, typeName, id string) LogicalObject {
	t.Helper()
	for _, obj := range db.Objects {
		if obj.Type == typeName && obj.ID == id {
			return obj
		}
	}
	t.Fatalf("missing logical object %s %s", typeName, id)
	return LogicalObject{}
}
