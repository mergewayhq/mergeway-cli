package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"gopkg.in/yaml.v3"
)

// Load reads the configuration entry file and resolves includes.
func Load(path string) (*Config, error) {
	if path == "" {
		return nil, errors.New("config: path is required")
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("config: resolve path: %w", err)
	}

	cache := make(map[string]*aggregateConfig)
	stack := make(map[string]bool)

	agg, err := loadRecursive(absPath, cache, stack)
	if err != nil {
		return nil, err
	}

	return normalizeAggregate(agg)
}

func loadRecursive(path string, cache map[string]*aggregateConfig, stack map[string]bool) (*aggregateConfig, error) {
	if cached, ok := cache[path]; ok {
		return cached, nil
	}

	if stack[path] {
		return nil, fmt.Errorf("config: detected include cycle at %s", path)
	}

	stack[path] = true
	defer delete(stack, path)

	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("config: read %s: %w", path, err)
	}

	var doc rawConfigDocument
	if err := yaml.Unmarshal(content, &doc); err != nil {
		return nil, fmt.Errorf("config: parse %s: %w", path, err)
	}

	agg := newAggregateConfig()
	baseDir := filepath.Dir(path)

	for _, include := range doc.Include {
		includePath := include
		if !filepath.IsAbs(includePath) {
			includePath = filepath.Join(baseDir, includePath)
		}

		matches, err := filepath.Glob(includePath)
		if err != nil {
			return nil, fmt.Errorf("config: glob %s from %s: %w", include, path, err)
		}

		if len(matches) == 0 {
			return nil, fmt.Errorf("config: include pattern %q in %s matched no files", include, path)
		}

		sort.Strings(matches)

		for _, match := range matches {
			absMatch, err := filepath.Abs(match)
			if err != nil {
				return nil, fmt.Errorf("config: resolve include %s: %w", match, err)
			}

			childAgg, err := loadRecursive(absMatch, cache, stack)
			if err != nil {
				return nil, err
			}

			if err := agg.merge(childAgg); err != nil {
				return nil, err
			}
		}
	}

	if err := agg.addDocument(&doc, path); err != nil {
		return nil, err
	}

	cache[path] = agg
	return agg, nil
}

func (a *aggregateConfig) merge(other *aggregateConfig) error {
	if other == nil {
		return nil
	}

	if other.VersionSet {
		if a.VersionSet {
			if a.Version != other.Version {
				return fmt.Errorf("config: version mismatch (got %d and %d)", a.Version, other.Version)
			}
		} else {
			a.Version = other.Version
			a.VersionSet = true
		}
	}

	for name, t := range other.Entities {
		if existing, ok := a.Entities[name]; ok {
			return fmt.Errorf("config: entity %q defined in both %s and %s", name, existing.Source, t.Source)
		}
		a.Entities[name] = t
	}

	return nil
}

func (a *aggregateConfig) addDocument(doc *rawConfigDocument, source string) error {
	if doc == nil {
		return nil
	}

	if doc.Version != nil {
		if a.VersionSet {
			if a.Version != *doc.Version {
				return fmt.Errorf("config: version mismatch (got %d and %d)", a.Version, *doc.Version)
			}
		} else {
			a.Version = *doc.Version
			a.VersionSet = true
		}
	}

	for name, spec := range doc.Entities {
		if name == "" {
			return fmt.Errorf("config: unnamed entity in %s", source)
		}

		if existing, ok := a.Entities[name]; ok {
			return fmt.Errorf("config: entity %q defined in both %s and %s", name, existing.Source, source)
		}

		a.Entities[name] = rawTypeWithSource{
			Name:   name,
			Spec:   spec,
			Source: source,
		}
	}

	return nil
}
