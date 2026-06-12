package cli

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/mergewayhq/mergeway-cli/internal/config"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

type listFileEntry struct {
	Type string `json:"type" yaml:"type"`
	File string `json:"file" yaml:"file"`
}

func newFilesCommand() *cobra.Command {
	var typeName string
	var group bool

	cmd := &cobra.Command{
		Use:   "files",
		Short: "List included YAML files and their entity types",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, err := contextFromCommand(cmd)
			if err != nil {
				return err
			}

			cfg, err := loadConfig(ctx)
			if err != nil {
				_, _ = fmt.Fprintf(ctx.Stderr, "files: %v\n", err)
				return newExitError(1)
			}
			if typeName != "" {
				if _, ok := cfg.Types[typeName]; !ok {
					_, _ = fmt.Fprintf(ctx.Stderr, "files: unknown type %s\n", typeName)
					return newExitError(1)
				}
			}

			entries, err := collectListFileEntries(ctx.Root, cfg, typeName, group)
			if err != nil {
				_, _ = fmt.Fprintf(ctx.Stderr, "files: %v\n", err)
				return newExitError(1)
			}

			sort.Slice(entries, func(i, j int) bool {
				if entries[i].Type != entries[j].Type {
					return entries[i].Type < entries[j].Type
				}
				return entries[i].File < entries[j].File
			})

			if code := writeFormatted(ctx, entries); code != 0 {
				return newExitError(code)
			}
			return nil
		},
	}

	cmd.Flags().BoolVar(&group, "group", false, "Group output by storage container")
	cmd.Flags().StringVar(&typeName, "type", "", "Type identifier")

	return cmd
}

func collectListFileEntries(root string, cfg *config.Config, typeName string, group bool) ([]listFileEntry, error) {
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return nil, fmt.Errorf("resolve root %s: %w", root, err)
	}

	files := &configuredFileSet{
		files: make(map[string]*config.TypeDefinition),
	}
	entries := make(map[string]listFileEntry)

	if cfg == nil {
		return nil, nil
	}

	for _, typeDef := range cfg.Types {
		if typeDef == nil {
			continue
		}
		if typeName != "" && typeDef.Name != typeName {
			continue
		}

		for _, include := range typeDef.Include {
			if include.Path == "" {
				continue
			}

			pattern := include.Path
			if !filepath.IsAbs(pattern) {
				pattern = filepath.Join(absRoot, filepath.Clean(pattern))
			}

			matches, err := filepath.Glob(pattern)
			if err != nil {
				return nil, fmt.Errorf("glob %s: %w", include.Path, err)
			}
			if len(matches) == 0 {
				continue
			}

			for _, match := range matches {
				info, err := os.Stat(match)
				if err != nil {
					if errors.Is(err, fs.ErrNotExist) {
						continue
					}
					return nil, fmt.Errorf("stat %s: %w", match, err)
				}
				if info.IsDir() {
					continue
				}

				absMatch := match
				if !filepath.IsAbs(absMatch) {
					absMatch, err = filepath.Abs(absMatch)
					if err != nil {
						return nil, fmt.Errorf("resolve match %s: %w", match, err)
					}
				}
				absMatch = filepath.Clean(absMatch)

				if !isYAMLPath(absMatch) {
					continue
				}
				if err := files.add(absMatch, typeDef); err != nil {
					return nil, err
				}

				entryPath := displayWorkspacePath(absRoot, absMatch)
				if group && include.Selector == "" && includeHasGlob(include.Path) {
					if multi, ok := yamlFileContainsItems(absMatch); ok && !multi {
						entryPath = displayWorkspacePath(absRoot, pattern)
					}
				}

				key := typeDef.Name + "\x00" + entryPath
				entries[key] = listFileEntry{
					Type: typeDef.Name,
					File: entryPath,
				}
			}
		}
	}

	result := make([]listFileEntry, 0, len(entries))
	for _, entry := range entries {
		result = append(result, entry)
	}
	return result, nil
}

func isYAMLPath(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	return ext == ".yaml" || ext == ".yml"
}

func includeHasGlob(path string) bool {
	return strings.ContainsAny(path, "*?[")
}

func yamlFileContainsItems(path string) (multi bool, ok bool) {
	body, err := os.ReadFile(path)
	if err != nil {
		return false, false
	}

	var doc map[string]any
	if err := yaml.Unmarshal(body, &doc); err != nil {
		return false, false
	}

	_, multi = doc["items"]
	return multi, true
}

func displayWorkspacePath(root, path string) string {
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return path
	}
	absPath, err := filepath.Abs(path)
	if err != nil {
		return path
	}
	rel, err := filepath.Rel(absRoot, absPath)
	if err != nil {
		return path
	}
	return filepath.ToSlash(rel)
}
