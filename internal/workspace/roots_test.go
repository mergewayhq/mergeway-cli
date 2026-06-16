package workspace

import (
	"path/filepath"
	"testing"
)

func TestDetectConfigPathSupportsYAMLAndYML(t *testing.T) {
	base := filepath.Join("testdata", "phase4")

	ymlPath, found, err := DetectConfigPath(filepath.Join(base, "valid-basic"))
	if err != nil {
		t.Fatalf("DetectConfigPath(valid-basic): %v", err)
	}
	if !found || filepath.Base(ymlPath) != "mergeway.yml" {
		t.Fatalf("expected mergeway.yml to be detected, got found=%v path=%q", found, ymlPath)
	}

	yamlPath, found, err := DetectConfigPath(filepath.Join(base, "duplicate-id"))
	if err != nil {
		t.Fatalf("DetectConfigPath(duplicate-id): %v", err)
	}
	if !found || filepath.Base(yamlPath) != "mergeway.yaml" {
		t.Fatalf("expected mergeway.yaml to be detected, got found=%v path=%q", found, yamlPath)
	}
}

func TestOpenRootsDetectsMissingAndLoadsIndexes(t *testing.T) {
	base := filepath.Join("testdata", "phase4")

	set, err := OpenRoots([]string{
		filepath.Join(base, "valid-basic"),
		filepath.Join(base, "duplicate-id"),
		filepath.Join(base, "unknown-reference"),
		filepath.Join(base, "no-config"),
	})
	if err != nil {
		t.Fatalf("OpenRoots: %v", err)
	}

	if got := len(set.Roots); got != 3 {
		t.Fatalf("expected 3 detected roots, got %d", got)
	}
	if got := len(set.MissingRoots); got != 1 {
		t.Fatalf("expected 1 missing root, got %d", got)
	}
	if filepath.Base(set.MissingRoots[0]) != "no-config" {
		t.Fatalf("expected no-config to be reported missing, got %q", set.MissingRoots[0])
	}

	duplicateRoot := rootByBase(t, set, "duplicate-id")
	if got := len(duplicateRoot.Workspace.Find("User", "user-1")); got != 2 {
		t.Fatalf("expected duplicate-id root to keep two user-1 entries, got %d", got)
	}

	unknownRefRoot := rootByBase(t, set, "unknown-reference")
	if got := len(unknownRefRoot.Workspace.Find("Post", "post-1")); got != 1 {
		t.Fatalf("expected unknown-reference root to index post-1, got %d", got)
	}
	if got := len(unknownRefRoot.Workspace.Find("User", "missing-user")); got != 0 {
		t.Fatalf("expected unknown-reference root not to synthesize missing users, got %d", got)
	}
}

func TestOpenRootsKeepsMultiRootIndexesIsolated(t *testing.T) {
	base := filepath.Join("testdata", "phase4", "multi-root")

	set, err := OpenRoots([]string{
		filepath.Join(base, "root-a"),
		filepath.Join(base, "root-b"),
	})
	if err != nil {
		t.Fatalf("OpenRoots(multi-root): %v", err)
	}

	if got := len(set.Roots); got != 2 {
		t.Fatalf("expected 2 detected roots, got %d", got)
	}

	rootA := rootByBase(t, set, "root-a")
	rootB := rootByBase(t, set, "root-b")

	aUser := rootA.Workspace.Find("User", "shared-user")
	bUser := rootB.Workspace.Find("User", "shared-user")
	if len(aUser) != 1 || len(bUser) != 1 {
		t.Fatalf("expected one shared-user per root, got %d and %d", len(aUser), len(bUser))
	}
	if aUser[0].Fields["name"] == bUser[0].Fields["name"] {
		t.Fatalf("expected root isolation to preserve distinct records, got matching names %q", aUser[0].Fields["name"])
	}
}

func TestRootIndexOwnsOnlyConfiguredFiles(t *testing.T) {
	base := filepath.Join("testdata", "phase4", "valid-basic")

	set, err := OpenRoots([]string{base})
	if err != nil {
		t.Fatalf("OpenRoots(valid-basic): %v", err)
	}
	root := rootByBase(t, set, "valid-basic")

	inScope := filepath.Join(base, "data", "users", "alice.yaml")
	outOfScope := filepath.Join(base, "notes", "ignored.yaml")
	configFile := filepath.Join(base, "mergeway.yml")

	if !root.OwnsPath(inScope) {
		t.Fatalf("expected in-scope file to be owned: %s", inScope)
	}
	if root.OwnsPath(outOfScope) {
		t.Fatalf("expected out-of-scope file to be ignored: %s", outOfScope)
	}
	if !root.OwnsPath(configFile) {
		t.Fatalf("expected config file to be tracked: %s", configFile)
	}

	types := root.TypesForFile(inScope)
	if len(types) != 1 || types[0] != "User" {
		t.Fatalf("expected User ownership for %s, got %v", inScope, types)
	}
}

func rootByBase(t *testing.T, set *RootSet, base string) *RootIndex {
	t.Helper()
	for _, root := range set.Roots {
		if filepath.Base(root.Root) == base {
			return root
		}
	}
	t.Fatalf("root %q not found", base)
	return nil
}
