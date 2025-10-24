package data

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/mergewayhq/mergeway-cli/internal/config"
)

type includeMatch struct {
	include config.IncludeDefinition
	path    string
}

func (s *Store) loadAll(typeDef *config.TypeDefinition) ([]*Object, error) {
	matches, err := s.resolveIncludeMatches(typeDef)
	if err != nil {
		return nil, err
	}

	seenIDs := make(map[string]struct{})
	idField := typeDef.Identifier.Field
	var objects []*Object
	for _, match := range matches {
		fc, err := s.loadFile(match.path, typeDef.Name, match.include.Selector)
		if err != nil {
			return nil, err
		}

		if fc.TypeName != "" && fc.TypeName != typeDef.Name {
			return nil, fmt.Errorf("data: file %s declares type %s; expected %s", fc.Path, fc.TypeName, typeDef.Name)
		}

		if fc.Multi {
			for _, item := range fc.Items {
				idVal, err := requiredString(item, idField)
				if err != nil {
					return nil, fmt.Errorf("data: %s in %s: %w", typeDef.Name, fc.Path, err)
				}
				objects = append(objects, &Object{Type: typeDef.Name, ID: idVal, Fields: cloneMap(item), File: fc.Path})
				seenIDs[idVal] = struct{}{}
			}
			continue
		}

		idVal, err := requiredString(fc.Single, idField)
		if err != nil {
			return nil, fmt.Errorf("data: %s in %s: %w", typeDef.Name, fc.Path, err)
		}

		objects = append(objects, &Object{Type: typeDef.Name, ID: idVal, Fields: cloneMap(fc.Single), File: fc.Path})
		seenIDs[idVal] = struct{}{}
	}

	for idx, item := range typeDef.InlineData {
		idVal, err := requiredString(item, idField)
		if err != nil {
			return nil, fmt.Errorf("data: %s inline item %d: %w", typeDef.Name, idx+1, err)
		}
		if _, exists := seenIDs[idVal]; exists {
			continue
		}
		objects = append(objects, &Object{Type: typeDef.Name, ID: idVal, Fields: cloneMap(item)})
		seenIDs[idVal] = struct{}{}
	}

	return objects, nil
}

func (s *Store) findObject(typeDef *config.TypeDefinition, id string) (*objectLocation, error) {
	matches, err := s.resolveIncludeMatches(typeDef)
	if err != nil {
		return nil, err
	}

	idField := typeDef.Identifier.Field

	for _, match := range matches {
		fc, err := s.loadFile(match.path, typeDef.Name, match.include.Selector)
		if err != nil {
			if errors.Is(err, errFileNotFound) {
				continue
			}
			return nil, err
		}

		if fc.TypeName != "" && fc.TypeName != typeDef.Name {
			return nil, fmt.Errorf("data: file %s declares type %s; expected %s", fc.Path, fc.TypeName, typeDef.Name)
		}

		if fc.Multi {
			for idx, item := range fc.Items {
				val, _ := getString(item, idField)
				if val == id {
					return &objectLocation{
						FilePath: match.path,
						Format:   fc.Format,
						Multi:    true,
						Index:    idx,
						Object:   cloneMap(item),
						File:     fc,
						TypeName: typeDef.Name,
						IDField:  idField,
						ReadOnly: fc.ReadOnly,
					}, nil
				}
			}
			continue
		}

		val, _ := getString(fc.Single, idField)
		if val == id {
			return &objectLocation{
				FilePath: match.path,
				Format:   fc.Format,
				Multi:    false,
				Object:   cloneMap(fc.Single),
				File:     fc,
				TypeName: typeDef.Name,
				IDField:  idField,
				ReadOnly: fc.ReadOnly,
			}, nil
		}
	}

	for idx, item := range typeDef.InlineData {
		val, err := requiredString(item, idField)
		if err != nil {
			return nil, fmt.Errorf("data: %s inline item %d: %w", typeDef.Name, idx+1, err)
		}
		if val == id {
			return &objectLocation{
				FilePath: "",
				Format:   formatYAML,
				Multi:    false,
				Index:    -1,
				Object:   cloneMap(item),
				File:     nil,
				TypeName: typeDef.Name,
				IDField:  idField,
				Inline:   true,
			}, nil
		}
	}

	return nil, nil
}

func (s *Store) resolveIncludeMatches(typeDef *config.TypeDefinition) ([]includeMatch, error) {
	seen := make(map[string]struct{})
	var matches []includeMatch

	for _, include := range typeDef.Include {
		pattern := include.Path
		if pattern == "" {
			continue
		}

		absPattern := pattern
		if !filepath.IsAbs(absPattern) {
			absPattern = filepath.Join(s.root, filepath.Clean(pattern))
		}

		globbed, err := filepath.Glob(absPattern)
		if err != nil {
			return nil, fmt.Errorf("data: glob %s: %w", pattern, err)
		}

		sort.Strings(globbed)

		for _, match := range globbed {
			info, err := os.Stat(match)
			if err != nil {
				if errors.Is(err, errFileNotFound) {
					continue
				}
				return nil, fmt.Errorf("data: stat %s: %w", match, err)
			}
			if info.IsDir() {
				continue
			}

			key := include.Path + "\x00" + include.Selector + "\x00" + match
			if _, ok := seen[key]; ok {
				continue
			}
			seen[key] = struct{}{}

			matches = append(matches, includeMatch{include: include, path: match})
		}
	}

	return matches, nil
}
