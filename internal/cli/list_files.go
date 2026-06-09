package cli

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/spf13/cobra"
)

type listFileEntry struct {
	Type string `json:"type" yaml:"type"`
	File string `json:"file" yaml:"file"`
}

func newFilesCommand() *cobra.Command {
	var typeName string

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

			files, err := collectConfiguredFiles(ctx.Root, cfg)
			if err != nil {
				_, _ = fmt.Fprintf(ctx.Stderr, "files: %v\n", err)
				return newExitError(1)
			}

			entries := make([]listFileEntry, 0, len(files.files))
			for _, path := range files.Slice() {
				typeDef := files.TypeFor(path)
				if typeDef == nil {
					continue
				}
				if typeName != "" && typeDef.Name != typeName {
					continue
				}

				ext := strings.ToLower(filepath.Ext(path))
				if ext != ".yaml" && ext != ".yml" {
					continue
				}

				entries = append(entries, listFileEntry{
					Type: typeDef.Name,
					File: displayWorkspacePath(ctx.Root, path),
				})
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

	cmd.Flags().StringVar(&typeName, "type", "", "Type identifier")

	return cmd
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
