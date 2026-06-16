package workspace

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/mergewayhq/mergeway-cli/internal/config"
)

var configFileNames = []string{"mergeway.yaml", "mergeway.yml"}

// RootIndex captures one detected mergeway root and its isolated file/index state.
type RootIndex struct {
	Root        string
	ConfigPath  string
	ConfigFiles map[string]struct{}
	DataFiles   map[string][]string
	Workspace   *Workspace
}

// RootSet captures all roots detected during one initialization pass.
type RootSet struct {
	Roots        []*RootIndex
	MissingRoots []string
}

// OpenRoots detects mergeway roots under the provided directories and loads one
// isolated workspace index per discovered root.
func OpenRoots(candidates []string) (*RootSet, error) {
	result := &RootSet{}
	seenRoots := make(map[string]struct{})

	for _, candidate := range candidates {
		resolvedRoot, err := normalizeRootCandidate(candidate)
		if err != nil {
			return nil, err
		}
		if resolvedRoot == "" {
			continue
		}
		if _, ok := seenRoots[resolvedRoot]; ok {
			continue
		}
		seenRoots[resolvedRoot] = struct{}{}

		configPath, found, err := DetectConfigPath(resolvedRoot)
		if err != nil {
			return nil, err
		}
		if !found {
			result.MissingRoots = append(result.MissingRoots, resolvedRoot)
			continue
		}

		cfg, err := config.Load(configPath)
		if err != nil {
			return nil, err
		}
		ws, err := LoadWithConfig(resolvedRoot, configPath, cfg)
		if err != nil {
			return nil, err
		}

		rootIndex, err := buildRootIndex(resolvedRoot, configPath, cfg, ws)
		if err != nil {
			return nil, err
		}
		result.Roots = append(result.Roots, rootIndex)
	}

	sort.Slice(result.Roots, func(i, j int) bool {
		return result.Roots[i].Root < result.Roots[j].Root
	})
	sort.Strings(result.MissingRoots)

	return result, nil
}

// DetectConfigPath resolves the mergeway config entry file for a root, checking
// both mergeway.yaml and mergeway.yml.
func DetectConfigPath(root string) (string, bool, error) {
	if strings.TrimSpace(root) == "" {
		return "", false, errors.New("workspace: root is required")
	}

	resolvedRoot, err := filepath.Abs(root)
	if err != nil {
		return "", false, fmt.Errorf("workspace: resolve root: %w", err)
	}

	for _, name := range configFileNames {
		candidate := filepath.Join(resolvedRoot, name)
		info, err := os.Stat(candidate)
		switch {
		case err == nil && !info.IsDir():
			return candidate, true, nil
		case errors.Is(err, fs.ErrNotExist):
			continue
		case err != nil:
			return "", false, fmt.Errorf("workspace: stat %s: %w", candidate, err)
		}
	}

	return "", false, nil
}

// OwnsPath reports whether the file belongs to this root as either a config
// file or a configured data include target.
func (r *RootIndex) OwnsPath(path string) bool {
	resolved, ok := normalizeOwnedPath(path)
	if !ok {
		return false
	}
	if _, exists := r.ConfigFiles[resolved]; exists {
		return true
	}
	_, exists := r.DataFiles[resolved]
	return exists
}

// TypesForFile reports the configured types that include the given file.
func (r *RootIndex) TypesForFile(path string) []string {
	resolved, ok := normalizeOwnedPath(path)
	if !ok {
		return nil
	}
	return append([]string(nil), r.DataFiles[resolved]...)
}

func buildRootIndex(root, configPath string, cfg *config.Config, ws *Workspace) (*RootIndex, error) {
	dataFiles, err := collectOwnedDataFiles(root, cfg)
	if err != nil {
		return nil, err
	}

	configFiles := make(map[string]struct{})
	for _, source := range append([]string{configPath}, cfg.Sources()...) {
		resolved, ok := normalizeOwnedPath(source)
		if !ok {
			continue
		}
		configFiles[resolved] = struct{}{}
	}

	return &RootIndex{
		Root:        root,
		ConfigPath:  configPath,
		ConfigFiles: configFiles,
		DataFiles:   dataFiles,
		Workspace:   ws,
	}, nil
}

func collectOwnedDataFiles(root string, cfg *config.Config) (map[string][]string, error) {
	files := make(map[string][]string)
	if cfg == nil {
		return files, nil
	}

	for _, typeName := range sortedTypeNames(cfg.Types) {
		typeDef := cfg.Types[typeName]
		if typeDef == nil {
			continue
		}

		for _, include := range typeDef.Include {
			if include.Path == "" {
				continue
			}

			pattern := include.Path
			if !filepath.IsAbs(pattern) {
				pattern = filepath.Join(root, filepath.Clean(pattern))
			}

			matches, err := filepath.Glob(pattern)
			if err != nil {
				return nil, fmt.Errorf("workspace: glob %s: %w", include.Path, err)
			}

			sort.Strings(matches)
			for _, match := range matches {
				info, err := os.Stat(match)
				if err != nil {
					if errors.Is(err, fs.ErrNotExist) {
						continue
					}
					return nil, fmt.Errorf("workspace: stat %s: %w", match, err)
				}
				if info.IsDir() {
					continue
				}

				resolved, ok := normalizeOwnedPath(match)
				if !ok {
					continue
				}
				files[resolved] = appendIfMissing(files[resolved], typeDef.Name)
			}
		}
	}

	return files, nil
}

func normalizeRootCandidate(candidate string) (string, error) {
	if strings.TrimSpace(candidate) == "" {
		return "", nil
	}

	resolved, err := filepath.Abs(candidate)
	if err != nil {
		return "", fmt.Errorf("workspace: resolve candidate %s: %w", candidate, err)
	}

	info, err := os.Stat(resolved)
	if err != nil {
		return "", fmt.Errorf("workspace: stat %s: %w", resolved, err)
	}
	if info.IsDir() {
		return filepath.Clean(resolved), nil
	}

	base := strings.ToLower(filepath.Base(resolved))
	for _, name := range configFileNames {
		if base == name {
			return filepath.Dir(resolved), nil
		}
	}

	return filepath.Dir(resolved), nil
}

func normalizeOwnedPath(path string) (string, bool) {
	if strings.TrimSpace(path) == "" {
		return "", false
	}
	resolved, err := filepath.Abs(path)
	if err != nil {
		return "", false
	}
	return filepath.Clean(resolved), true
}

func appendIfMissing(values []string, value string) []string {
	for _, existing := range values {
		if existing == value {
			return values
		}
	}
	return append(values, value)
}
