package workspace

import (
	"path/filepath"
	"reflect"
	"sort"
	"testing"

	"github.com/mergewayhq/mergeway-cli/internal/config"
	"github.com/mergewayhq/mergeway-cli/internal/validation"
)

func TestLoadBuildsWorkspaceIndex(t *testing.T) {
	root := filepath.Join("..", "data", "testdata", "repo")

	ws, err := Load(root, "")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if ws.Root == "" || ws.ConfigPath == "" {
		t.Fatalf("expected resolved root/config paths, got %+v", ws)
	}
	if ws.Config == nil {
		t.Fatalf("expected config to be loaded")
	}

	if got := len(ws.Objects("User")); got != 2 {
		t.Fatalf("expected 2 users, got %d", got)
	}
	if got := len(ws.Objects("Post")); got != 2 {
		t.Fatalf("expected 2 posts, got %d", got)
	}

	alice := ws.Find("User", "User-Alice")
	if len(alice) != 1 {
		t.Fatalf("expected exactly one User-Alice, got %d", len(alice))
	}
	if alice[0].Fields["email"] != "alice@example.com" {
		t.Fatalf("expected indexed user fields, got %+v", alice[0].Fields)
	}

	post := ws.Find("Post", "Post-001")
	if len(post) != 1 {
		t.Fatalf("expected exactly one Post-001, got %d", len(post))
	}
	expectedPostFile, err := filepath.Abs(filepath.Join(root, "data", "posts", "posts.yaml"))
	if err != nil {
		t.Fatalf("abs post path: %v", err)
	}
	if post[0].File != expectedPostFile {
		t.Fatalf("expected indexed post file, got %q", post[0].File)
	}
}

func TestValidateMatchesValidationPackage(t *testing.T) {
	cases := []string{
		filepath.Join("..", "validation", "testdata", "valid"),
		filepath.Join("..", "validation", "testdata", "schema_error"),
		filepath.Join("..", "validation", "testdata", "reference_error"),
	}

	for _, root := range cases {
		root := root
		t.Run(filepath.Base(root), func(t *testing.T) {
			cfgPath := filepath.Join(root, "mergeway.yaml")
			cfg, err := config.Load(cfgPath)
			if err != nil {
				t.Fatalf("load config: %v", err)
			}

			expected, err := validation.Validate(root, cfg, validation.Options{})
			if err != nil {
				t.Fatalf("validation.Validate: %v", err)
			}

			report, err := Validate(root, "", validation.Options{})
			if err != nil {
				t.Fatalf("workspace.Validate: %v", err)
			}

			if !reflect.DeepEqual(sortedErrors(report.Result.Errors), sortedErrors(expected.Errors)) {
				t.Fatalf("expected parity with validation package\nexpected: %+v\ngot: %+v", expected, report.Result)
			}

			if filepath.Base(root) == "valid" {
				if report.Workspace == nil {
					t.Fatalf("expected loaded workspace for valid fixture, got load error %v", report.WorkspaceLoadError)
				}
			}
		})
	}
}

func sortedErrors(errs []validation.Error) []validation.Error {
	cloned := append([]validation.Error(nil), errs...)
	sort.Slice(cloned, func(i, j int) bool {
		left := cloned[i]
		right := cloned[j]
		if left.Phase != right.Phase {
			return left.Phase < right.Phase
		}
		if left.Type != right.Type {
			return left.Type < right.Type
		}
		if left.ID != right.ID {
			return left.ID < right.ID
		}
		if left.File != right.File {
			return left.File < right.File
		}
		return left.Message < right.Message
	})
	return cloned
}
