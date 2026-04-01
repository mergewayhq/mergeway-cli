package diff

import (
	"errors"
	"fmt"
	"os/exec"
	"strings"
)

var ErrTooManyArgs = errors.New("diff accepts at most 2 snapshot arguments")

type SnapshotKind string

const (
	SnapshotKindHead        SnapshotKind = "head"
	SnapshotKindRevision    SnapshotKind = "revision"
	SnapshotKindWorkingTree SnapshotKind = "working_tree"
)

type WorkingTreeView string

const (
	WorkingTreeViewFull        WorkingTreeView = "full"
	WorkingTreeViewUnstaged    WorkingTreeView = "unstaged"
	WorkingTreeViewUnspecified WorkingTreeView = ""
)

type SnapshotRef struct {
	Kind            SnapshotKind
	Revision        string
	WorkingTreeView WorkingTreeView
}

type DiffSnapshots struct {
	Left  SnapshotRef
	Right SnapshotRef
}

func (s SnapshotRef) String() string {
	switch s.Kind {
	case SnapshotKindHead:
		return "HEAD"
	case SnapshotKindRevision:
		return s.Revision
	case SnapshotKindWorkingTree:
		switch s.WorkingTreeView {
		case WorkingTreeViewUnstaged:
			return "WORKTREE_UNSTAGED"
		default:
			return "WORKTREE"
		}
	default:
		return "UNKNOWN"
	}
}

// resolveDiffSnapshots keeps the working-tree modes explicit so staged vs
// unstaged semantics remain visible in code and protected by tests.
func resolveDiffSnapshots(root string, args []string) (DiffSnapshots, error) {
	switch len(args) {
	case 0:
		if err := validateGitRevision(root, "HEAD"); err != nil {
			return DiffSnapshots{}, err
		}
		return DiffSnapshots{
			Left: SnapshotRef{
				Kind: SnapshotKindHead,
			},
			Right: SnapshotRef{
				Kind:            SnapshotKindWorkingTree,
				WorkingTreeView: WorkingTreeViewUnstaged,
			},
		}, nil
	case 1:
		if err := validateGitRevision(root, args[0]); err != nil {
			return DiffSnapshots{}, err
		}
		return DiffSnapshots{
			Left: SnapshotRef{
				Kind:     SnapshotKindRevision,
				Revision: args[0],
			},
			Right: SnapshotRef{
				Kind:            SnapshotKindWorkingTree,
				WorkingTreeView: WorkingTreeViewFull,
			},
		}, nil
	case 2:
		if err := validateGitRevision(root, args[0]); err != nil {
			return DiffSnapshots{}, err
		}
		if err := validateGitRevision(root, args[1]); err != nil {
			return DiffSnapshots{}, err
		}
		return DiffSnapshots{
			Left: SnapshotRef{
				Kind:     SnapshotKindRevision,
				Revision: args[0],
			},
			Right: SnapshotRef{
				Kind:     SnapshotKindRevision,
				Revision: args[1],
			},
		}, nil
	default:
		return DiffSnapshots{}, ErrTooManyArgs
	}
}

func validateGitRevision(root, revision string) error {
	cmd := exec.Command("git", "-C", root, "rev-parse", "--verify", "--quiet", revision+"^{commit}")
	output, err := cmd.CombinedOutput()
	if err == nil {
		return nil
	}

	message := strings.TrimSpace(string(output))
	if message == "" {
		return fmt.Errorf("invalid revision %q", revision)
	}
	return fmt.Errorf("resolve revision %q: %s", revision, message)
}
