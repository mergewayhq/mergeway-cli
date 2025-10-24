package data

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/mergewayhq/mergeway-cli/internal/config"
)

func (s *Store) loadAll(typeDef *config.TypeDefinition) ([]*Object, error) {
	files, err := s.matchTypeFiles(typeDef)
	if err != nil {
		return nil, err
	}

	var objects []*Object
	for _, file := range files {
		fc, err := s.loadFile(file, typeDef.Name)
		if err != nil {
			return nil, err
		}

		if fc.TypeName != typeDef.Name {
			return nil, fmt.Errorf("data: file %s declares type %s; expected %s", fc.Path, fc.TypeName, typeDef.Name)
		}

		if fc.Multi {
			for _, item := range fc.Items {
				idField := typeDef.Identifier.Field
				idVal, err := requiredString(item, idField)
				if err != nil {
					return nil, fmt.Errorf("data: %s in %s: %w", typeDef.Name, fc.Path, err)
				}
				objects = append(objects, &Object{Type: typeDef.Name, ID: idVal, Fields: cloneMap(item), File: fc.Path})
			}
			continue
		}

		idField := typeDef.Identifier.Field
		idVal, err := requiredString(fc.Single, idField)
		if err != nil {
			return nil, fmt.Errorf("data: %s in %s: %w", typeDef.Name, fc.Path, err)
		}

		objects = append(objects, &Object{Type: typeDef.Name, ID: idVal, Fields: cloneMap(fc.Single), File: fc.Path})
	}

	return objects, nil
}

func (s *Store) findObject(typeDef *config.TypeDefinition, id string) (*objectLocation, error) {
	files, err := s.matchTypeFiles(typeDef)
	if err != nil {
		return nil, err
	}

	idField := typeDef.Identifier.Field

	for _, file := range files {
		fc, err := s.loadFile(file, typeDef.Name)
		if err != nil {
			if errors.Is(err, errFileNotFound) {
				continue
			}
			return nil, err
		}

		if fc.TypeName != typeDef.Name {
			return nil, fmt.Errorf("data: file %s declares type %s; expected %s", fc.Path, fc.TypeName, typeDef.Name)
		}

		if fc.Multi {
			for idx, item := range fc.Items {
				val, _ := getString(item, idField)
				if val == id {
					return &objectLocation{
						FilePath: file,
						Format:   fc.Format,
						Multi:    true,
						Index:    idx,
						Object:   cloneMap(item),
						File:     fc,
						TypeName: typeDef.Name,
						IDField:  idField,
					}, nil
				}
			}
			continue
		}

		val, _ := getString(fc.Single, idField)
		if val == id {
			return &objectLocation{
				FilePath: file,
				Format:   fc.Format,
				Multi:    false,
				Object:   cloneMap(fc.Single),
				File:     fc,
				TypeName: typeDef.Name,
				IDField:  idField,
			}, nil
		}
	}

	return nil, nil
}

func (s *Store) matchTypeFiles(typeDef *config.TypeDefinition) ([]string, error) {
	seen := make(map[string]struct{})
	var files []string

	for _, pattern := range typeDef.Include {
		absPattern := pattern
		if !filepath.IsAbs(absPattern) {
			absPattern = filepath.Join(s.root, filepath.Clean(pattern))
		}

		matches, err := filepath.Glob(absPattern)
		if err != nil {
			return nil, fmt.Errorf("data: glob %s: %w", pattern, err)
		}

		for _, match := range matches {
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
			if _, ok := seen[match]; ok {
				continue
			}
			seen[match] = struct{}{}
			files = append(files, match)
		}
	}

	sort.Strings(files)
	return files, nil
}
