package diff

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

type SnapshotDataFile struct {
	Path    string
	Exists  bool
	Content []byte
}

type SnapshotDataCorpus struct {
	Snapshot SnapshotRef
	Schema   *diffSnapshotSchema
	Files    []SnapshotDataFile
}

type DiffDataCorpora struct {
	Paths []string
	Left  SnapshotDataCorpus
	Right SnapshotDataCorpus
}

func loadDiffDataCorpora(root, configPath string, left, right SnapshotRef) (DiffDataCorpora, error) {
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return DiffDataCorpora{}, fmt.Errorf("diff: resolve root: %w", err)
	}

	absConfig, err := filepath.Abs(configPath)
	if err != nil {
		return DiffDataCorpora{}, fmt.Errorf("diff: resolve config path: %w", err)
	}

	leftSchema, err := loadSnapshotDiffSchema(absRoot, absConfig, left)
	if err != nil {
		return DiffDataCorpora{}, err
	}
	rightSchema, err := loadSnapshotDiffSchema(absRoot, absConfig, right)
	if err != nil {
		return DiffDataCorpora{}, err
	}

	paths, err := discoverDiffDataPaths(absRoot, absConfig, left, right)
	if err != nil {
		return DiffDataCorpora{}, err
	}

	leftCorpus, err := loadSnapshotDataCorpus(absRoot, left, leftSchema, paths)
	if err != nil {
		return DiffDataCorpora{}, err
	}
	rightCorpus, err := loadSnapshotDataCorpus(absRoot, right, rightSchema, paths)
	if err != nil {
		return DiffDataCorpora{}, err
	}

	return DiffDataCorpora{
		Paths: paths,
		Left:  leftCorpus,
		Right: rightCorpus,
	}, nil
}

func discoverDiffDataPaths(root, configPath string, left, right SnapshotRef) ([]string, error) {
	leftPatterns, err := loadSnapshotDataIncludePatterns(root, configPath, left)
	if err != nil {
		return nil, err
	}
	rightPatterns, err := loadSnapshotDataIncludePatterns(root, configPath, right)
	if err != nil {
		return nil, err
	}

	patternSet := make(map[string]struct{}, len(leftPatterns)+len(rightPatterns))
	for _, pattern := range leftPatterns {
		patternSet[pattern] = struct{}{}
	}
	for _, pattern := range rightPatterns {
		patternSet[pattern] = struct{}{}
	}

	patterns := sortedKeys(patternSet)
	paths := make(map[string]struct{})
	for _, pattern := range patterns {
		leftMatches, err := matchSnapshotPattern(root, left, pattern)
		if err != nil {
			return nil, err
		}
		for _, path := range leftMatches {
			paths[path] = struct{}{}
		}

		rightMatches, err := matchSnapshotPattern(root, right, pattern)
		if err != nil {
			return nil, err
		}
		for _, path := range rightMatches {
			paths[path] = struct{}{}
		}
	}

	return sortedKeys(paths), nil
}

func loadSnapshotDataCorpus(root string, snapshot SnapshotRef, schema *diffSnapshotSchema, paths []string) (SnapshotDataCorpus, error) {
	reader, err := newSnapshotReader(root, snapshot)
	if err != nil {
		return SnapshotDataCorpus{}, err
	}

	files := make([]SnapshotDataFile, 0, len(paths))
	for _, path := range paths {
		content, exists, err := reader.Read(path)
		if err != nil {
			return SnapshotDataCorpus{}, err
		}
		files = append(files, SnapshotDataFile{
			Path:    path,
			Exists:  exists,
			Content: content,
		})
	}

	return SnapshotDataCorpus{
		Snapshot: snapshot,
		Schema:   schema,
		Files:    files,
	}, nil
}

type snapshotReader struct {
	root     string
	snapshot SnapshotRef

	files     []string
	fileSet   map[string]struct{}
	unstaged  map[string]struct{}
	untracked map[string]struct{}
}

func newSnapshotReader(root string, snapshot SnapshotRef) (*snapshotReader, error) {
	r := &snapshotReader{
		root:     root,
		snapshot: snapshot,
	}

	if snapshot.Kind == SnapshotKindWorkingTree && snapshot.WorkingTreeView == WorkingTreeViewUnstaged {
		unstaged, err := listGitNames(root, "diff", "--name-only", "--no-renames")
		if err != nil {
			return nil, err
		}
		r.unstaged = sliceToSet(unstaged)

		untracked, err := listGitNames(root, "ls-files", "--others", "--exclude-standard")
		if err != nil {
			return nil, err
		}
		r.untracked = sliceToSet(untracked)
	}

	return r, nil
}

func (r *snapshotReader) Read(path string) ([]byte, bool, error) {
	switch r.snapshot.Kind {
	case SnapshotKindHead:
		return readGitRevisionFile(r.root, "HEAD", path)
	case SnapshotKindRevision:
		return readGitRevisionFile(r.root, r.snapshot.Revision, path)
	case SnapshotKindWorkingTree:
		if r.snapshot.WorkingTreeView == WorkingTreeViewUnstaged {
			if _, ok := r.untracked[path]; ok {
				return readWorkingTreeFile(r.root, path)
			}
			if _, ok := r.unstaged[path]; ok {
				return readWorkingTreeFile(r.root, path)
			}
			return readGitRevisionFile(r.root, "HEAD", path)
		}
		return readWorkingTreeFile(r.root, path)
	default:
		return nil, false, fmt.Errorf("diff: unsupported snapshot kind %q", r.snapshot.Kind)
	}
}

func (r *snapshotReader) ListFiles() ([]string, error) {
	if r.files != nil {
		return append([]string(nil), r.files...), nil
	}

	var files []string
	switch r.snapshot.Kind {
	case SnapshotKindHead:
		var err error
		files, err = listGitTreeFiles(r.root, "HEAD")
		if err != nil {
			return nil, err
		}
	case SnapshotKindRevision:
		var err error
		files, err = listGitTreeFiles(r.root, r.snapshot.Revision)
		if err != nil {
			return nil, err
		}
	case SnapshotKindWorkingTree:
		var err error
		files, err = listWorkingTreeFiles(r.root)
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("diff: unsupported snapshot kind %q", r.snapshot.Kind)
	}

	r.files = files
	r.fileSet = sliceToSet(files)
	return append([]string(nil), files...), nil
}

func loadSnapshotDataIncludePatterns(root, configPath string, snapshot SnapshotRef) ([]string, error) {
	schema, err := loadSnapshotDiffSchema(root, configPath, snapshot)
	if err != nil {
		return nil, err
	}
	return schema.includePatterns(), nil
}

type diffConfigInclude struct {
	Path     string `yaml:"path"`
	Selector string `yaml:"selector"`
}

func (d *diffConfigInclude) UnmarshalYAML(node *yaml.Node) error {
	if node == nil {
		return nil
	}

	switch node.Kind {
	case yaml.ScalarNode:
		var path string
		if err := node.Decode(&path); err != nil {
			return err
		}
		path = strings.TrimSpace(path)
		if path == "" {
			return fmt.Errorf("include path must be a non-empty string")
		}
		d.Path = path
		d.Selector = ""
		return nil
	case yaml.MappingNode, yaml.AliasNode:
		type alias diffConfigInclude
		var tmp alias
		if err := node.Decode(&tmp); err != nil {
			return err
		}
		tmp.Path = strings.TrimSpace(tmp.Path)
		tmp.Selector = strings.TrimSpace(tmp.Selector)
		if tmp.Path == "" {
			return fmt.Errorf("include path must be a non-empty string")
		}
		*d = diffConfigInclude(tmp)
		return nil
	default:
		return fmt.Errorf("include entry must be a string or mapping, got %s", node.ShortTag())
	}
}

func normalizeSnapshotPattern(baseDir, pattern string) (string, error) {
	pattern = strings.TrimSpace(pattern)
	if pattern == "" {
		return "", errors.New("empty include path")
	}

	if filepath.IsAbs(pattern) {
		return "", fmt.Errorf("absolute include paths are not supported for snapshot loading: %s", pattern)
	}

	joined := filepath.Clean(filepath.Join(baseDir, pattern))
	if joined == ".." || strings.HasPrefix(joined, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("include path resolves outside repository root: %s", pattern)
	}

	return joined, nil
}

func matchSnapshotPattern(root string, snapshot SnapshotRef, pattern string) ([]string, error) {
	reader, err := newSnapshotReader(root, snapshot)
	if err != nil {
		return nil, err
	}
	files, err := reader.ListFiles()
	if err != nil {
		return nil, err
	}

	var matches []string
	for _, path := range files {
		ok, err := filepath.Match(pattern, path)
		if err != nil {
			return nil, fmt.Errorf("diff: match pattern %s: %w", pattern, err)
		}
		if ok {
			matches = append(matches, path)
		}
	}

	return matches, nil
}

func listWorkingTreeFiles(root string) ([]string, error) {
	var files []string
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if path == filepath.Join(root, ".git") {
				return filepath.SkipDir
			}
			return nil
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		files = append(files, filepath.Clean(rel))
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("diff: walk working tree: %w", err)
	}
	sort.Strings(files)
	return files, nil
}

func listGitTreeFiles(root, revision string) ([]string, error) {
	lines, err := listGitNames(root, "ls-tree", "-r", "--name-only", revision)
	if err != nil {
		return nil, err
	}
	sort.Strings(lines)
	return lines, nil
}

func listGitNames(root string, args ...string) ([]string, error) {
	output, err := runGit(root, args...)
	if err != nil {
		return nil, err
	}
	output = bytes.TrimSpace(output)
	if len(output) == 0 {
		return nil, nil
	}

	lines := strings.Split(string(output), "\n")
	sort.Strings(lines)
	return lines, nil
}

func readWorkingTreeFile(root, path string) ([]byte, bool, error) {
	target := filepath.Join(root, filepath.Clean(path))
	content, err := os.ReadFile(target)
	if err == nil {
		return content, true, nil
	}
	if errors.Is(err, fs.ErrNotExist) {
		return nil, false, nil
	}
	return nil, false, fmt.Errorf("diff: read working tree file %s: %w", path, err)
}

func readGitRevisionFile(root, revision, path string) ([]byte, bool, error) {
	spec := revision + ":" + path
	cmd := exec.Command("git", "-C", root, "cat-file", "-e", spec)
	if err := cmd.Run(); err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return nil, false, nil
		}
		return nil, false, fmt.Errorf("diff: inspect %s: %w", spec, err)
	}

	content, err := runGit(root, "show", spec)
	if err != nil {
		return nil, false, err
	}
	return content, true, nil
}

func runGit(root string, args ...string) ([]byte, error) {
	cmd := exec.Command("git", append([]string{"-C", root}, args...)...)
	output, err := cmd.CombinedOutput()
	if err == nil {
		return output, nil
	}
	message := strings.TrimSpace(string(output))
	if message == "" {
		message = err.Error()
	}
	return nil, fmt.Errorf("diff: git %s: %s", strings.Join(args, " "), message)
}

func rootRelativePath(root, path string) (string, error) {
	rel, err := filepath.Rel(root, path)
	if err != nil {
		return "", err
	}
	rel = filepath.Clean(rel)
	if rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("path resolves outside repository root")
	}
	return rel, nil
}

func sliceToSet(values []string) map[string]struct{} {
	if len(values) == 0 {
		return nil
	}
	set := make(map[string]struct{}, len(values))
	for _, value := range values {
		set[value] = struct{}{}
	}
	return set
}

func sortedKeys(set map[string]struct{}) []string {
	if len(set) == 0 {
		return nil
	}
	keys := make([]string, 0, len(set))
	for key := range set {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}
