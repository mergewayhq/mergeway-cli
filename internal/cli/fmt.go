package cli

import (
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/mergewayhq/mergeway-cli/internal/config"
	"github.com/mergewayhq/mergeway-cli/internal/format"
)

func cmdFmt(ctx *Context, args []string) int {
	fs := flag.NewFlagSet("fmt", flag.ContinueOnError)
	fs.SetOutput(ctx.Stderr)
	inPlace := fs.Bool("in-place", false, "Rewrite files in place")
	lint := fs.Bool("lint", false, "Fail if formatting differs from the canonical form")
	if err := fs.Parse(args); err != nil {
		return 1
	}

	if *lint && *inPlace {
		_, _ = fmt.Fprintln(ctx.Stderr, "fmt: --lint cannot be combined with --in-place")
		return 1
	}

	absRoot, err := filepath.Abs(ctx.Root)
	if err != nil {
		_, _ = fmt.Fprintf(ctx.Stderr, "fmt: resolve root: %v\n", err)
		return 1
	}

	cfg, err := loadConfig(ctx)
	if err != nil {
		_, _ = fmt.Fprintf(ctx.Stderr, "fmt: %v\n", err)
		return 1
	}

	tracked, err := collectConfiguredFiles(absRoot, cfg)
	if err != nil {
		_, _ = fmt.Fprintf(ctx.Stderr, "fmt: %v\n", err)
		return 1
	}

	paths := fs.Args()
	defaultTargets := len(paths) == 0
	var targets []string
	if defaultTargets {
		targets = tracked.Slice()
	} else {
		targets, err = expandFmtTargets(absRoot, paths)
		if err != nil {
			_, _ = fmt.Fprintf(ctx.Stderr, "fmt: %v\n", err)
			return 1
		}
		for _, path := range targets {
			if !tracked.Has(path) {
				_, _ = fmt.Fprintf(ctx.Stderr, "fmt: %s is not part of the configured data set\n", displayPath(absRoot, path))
				return 1
			}
		}
	}

	if len(targets) == 0 {
		if defaultTargets {
			_, _ = fmt.Fprintln(ctx.Stderr, "fmt: configuration does not reference any data files; nothing to format")
			return 0
		}
		_, _ = fmt.Fprintln(ctx.Stderr, "fmt: no files matched the provided paths")
		return 1
	}

	schemaCache := make(map[string]*format.Schema)
	schemaFor := func(path string) *format.Schema {
		typeDef := tracked.TypeFor(path)
		if typeDef == nil {
			return nil
		}
		if cached, ok := schemaCache[typeDef.Name]; ok {
			return cached
		}
		schema := buildFormatSchema(typeDef)
		schemaCache[typeDef.Name] = schema
		return schema
	}

	if *lint {
		return fmtLint(ctx, absRoot, targets, schemaFor)
	}
	if *inPlace {
		return fmtWriteInPlace(ctx, absRoot, targets, schemaFor)
	}
	return fmtWriteStdout(ctx, targets, schemaFor)
}

type configuredFileSet struct {
	files map[string]*config.TypeDefinition
}

func (c *configuredFileSet) add(path string, typeDef *config.TypeDefinition) error {
	if path == "" {
		return nil
	}
	if c.files == nil {
		c.files = make(map[string]*config.TypeDefinition)
	}
	if existing, ok := c.files[path]; ok {
		if existing != typeDef && existing.Name != typeDef.Name {
			return fmt.Errorf("%s referenced by multiple entity types (%s, %s)", path, existing.Name, typeDef.Name)
		}
		return nil
	}
	c.files[path] = typeDef
	return nil
}

func (c *configuredFileSet) Has(path string) bool {
	return c.TypeFor(path) != nil
}

func (c *configuredFileSet) TypeFor(path string) *config.TypeDefinition {
	if c == nil {
		return nil
	}
	return c.files[path]
}

func (c *configuredFileSet) Slice() []string {
	if c == nil || len(c.files) == 0 {
		return nil
	}
	paths := make([]string, 0, len(c.files))
	for path := range c.files {
		paths = append(paths, path)
	}
	sort.Strings(paths)
	return paths
}

func collectConfiguredFiles(root string, cfg *config.Config) (*configuredFileSet, error) {
	set := &configuredFileSet{
		files: make(map[string]*config.TypeDefinition),
	}

	if cfg == nil {
		return set, nil
	}

	for _, typeDef := range cfg.Types {
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
					absMatch = filepath.Join(root, match)
				}
				absMatch = filepath.Clean(absMatch)
				if err := set.add(absMatch, typeDef); err != nil {
					return nil, err
				}
			}
		}
	}

	return set, nil
}

func buildFormatSchema(typeDef *config.TypeDefinition) *format.Schema {
	if typeDef == nil || len(typeDef.FieldOrder) == 0 {
		return nil
	}

	fields := make([]*format.SchemaField, 0, len(typeDef.FieldOrder))
	for _, name := range typeDef.FieldOrder {
		field := typeDef.Fields[name]
		if field == nil {
			continue
		}
		if schemaField := buildSchemaField(field); schemaField != nil {
			fields = append(fields, schemaField)
		}
	}

	return format.NewSchema(fields)
}

func buildSchemaField(field *config.FieldDefinition) *format.SchemaField {
	if field == nil || field.Name == "" {
		return nil
	}
	schemaField := &format.SchemaField{
		Name:     field.Name,
		Repeated: field.Repeated,
	}
	if field.Type == "object" && len(field.PropertyOrder) > 0 {
		children := make([]*format.SchemaField, 0, len(field.PropertyOrder))
		for _, propName := range field.PropertyOrder {
			child := field.Properties[propName]
			if child == nil {
				continue
			}
			if nested := buildSchemaField(child); nested != nil {
				children = append(children, nested)
			}
		}
		schemaField.Nested = format.NewSchema(children)
	}
	return schemaField
}

func expandFmtTargets(root string, inputs []string) ([]string, error) {
	var targets []string
	seen := make(map[string]struct{})

	for _, raw := range inputs {
		if raw == "" {
			continue
		}

		pattern := raw
		if !filepath.IsAbs(pattern) {
			pattern = filepath.Join(root, pattern)
		}
		matches, err := filepath.Glob(pattern)
		if err != nil {
			return nil, fmt.Errorf("expand %s: %w", raw, err)
		}
		if len(matches) == 0 {
			matches = []string{pattern}
		}
		sort.Strings(matches)
		for _, match := range matches {
			if _, exists := seen[match]; exists {
				continue
			}
			info, err := os.Stat(match)
			if err != nil {
				if errors.Is(err, fs.ErrNotExist) {
					return nil, fmt.Errorf("%s does not exist", displayPath(root, match))
				}
				return nil, fmt.Errorf("stat %s: %w", displayPath(root, match), err)
			}
			if info.IsDir() {
				return nil, fmt.Errorf("%s is a directory", displayPath(root, match))
			}
			seen[match] = struct{}{}
			targets = append(targets, match)
		}
	}

	sort.Strings(targets)
	return targets, nil
}

func fmtWriteStdout(ctx *Context, paths []string, schemaFor func(string) *format.Schema) int {
	for _, path := range paths {
		result, err := format.FormatFile(path, schemaFor(path))
		if err != nil {
			_, _ = fmt.Fprintf(ctx.Stderr, "fmt: %v\n", err)
			return 1
		}

		if _, err := ctx.Stdout.Write(result.Content); err != nil {
			_, _ = fmt.Fprintf(ctx.Stderr, "fmt: write stdout: %v\n", err)
			return 1
		}
	}
	return 0
}

func fmtWriteInPlace(ctx *Context, root string, paths []string, schemaFor func(string) *format.Schema) int {
	for _, path := range paths {
		result, err := format.FormatFile(path, schemaFor(path))
		if err != nil {
			_, _ = fmt.Fprintf(ctx.Stderr, "fmt: %v\n", err)
			return 1
		}
		if !result.Changed {
			continue
		}
		if err := writeAtomic(path, result.Content); err != nil {
			_, _ = fmt.Fprintf(ctx.Stderr, "fmt: rewrite %s: %v\n", displayPath(root, path), err)
			return 1
		}
	}
	return 0
}

func fmtLint(ctx *Context, root string, paths []string, schemaFor func(string) *format.Schema) int {
	var needsFormat bool
	for _, path := range paths {
		result, err := format.FormatFile(path, schemaFor(path))
		if err != nil {
			_, _ = fmt.Fprintf(ctx.Stderr, "fmt: %v\n", err)
			return 1
		}
		if !result.Changed {
			continue
		}
		needsFormat = true
		_, _ = fmt.Fprintln(ctx.Stdout, displayPath(root, path))
	}
	if needsFormat {
		return 1
	}
	return 0
}

func writeAtomic(path string, data []byte) error {
	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, ".mwfmt-*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	defer func() {
		_ = os.Remove(tmpName)
	}()

	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Sync(); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}

	info, err := os.Stat(path)
	mode := fs.FileMode(0o644)
	if err == nil {
		mode = info.Mode()
	}

	if err := os.Chmod(tmpName, mode); err != nil {
		return err
	}

	return os.Rename(tmpName, path)
}

func displayPath(root, path string) string {
	if rel, err := filepath.Rel(root, path); err == nil && !strings.HasPrefix(rel, "..") {
		return rel
	}
	return path
}
