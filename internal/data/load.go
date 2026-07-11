package data

import (
	"errors"
	"fmt"
	"path/filepath"
	"sort"

	"github.com/mergewayhq/mergeway-cli/internal/config"
)

type includeMatch struct {
	include config.IncludeDefinition
	path    string
}

func (s *Store) loadAll(typeDef *config.TypeDefinition) ([]*Object, error) {
	typeDefs := s.assignableTypeDefs(typeDef)
	seenIDs := make(map[string]*Object)
	var objects []*Object

	for _, concreteType := range typeDefs {
		loaded, err := s.loadExactAll(concreteType)
		if err != nil {
			return nil, err
		}
		for _, obj := range loaded {
			if existing := seenIDs[obj.ID]; existing != nil {
				if existing.Type != obj.Type {
					return nil, duplicateHierarchyError(typeDef.Name, obj.ID, existing, obj)
				}
				objects = append(objects, obj)
				continue
			}
			seenIDs[obj.ID] = obj
			objects = append(objects, obj)
		}
	}

	return objects, nil
}

func (s *Store) loadExactAll(typeDef *config.TypeDefinition) ([]*Object, error) {
	matches, err := s.resolveIncludeMatches(typeDef)
	if err != nil {
		return nil, err
	}

	seenIDs := make(map[string]struct{})
	var objects []*Object
	for _, match := range matches {
		fc, err := s.loadFile(match.path, typeDef.Name, match.include.Selector)
		if err != nil {
			return nil, err
		}

		if fc.TypeName != "" && fc.TypeName != typeDef.Name {
			return nil, fmt.Errorf("data: file %s declares type %s; expected %s", fc.Path, fc.TypeName, typeDef.Name)
		}

		if fc.Multi && typeDef.Identifier.IsPath() {
			return nil, fmt.Errorf("data: %s uses identifier %q, but file %s contains multiple objects", typeDef.Name, config.PathIdentifierField, fc.Path)
		}

		if fc.Multi {
			for _, item := range fc.Items {
				idVal, _, err := deriveIdentifierValue(typeDef, item, fc.Path, s.root)
				if err != nil {
					return nil, fmt.Errorf("data: %s in %s: %w", typeDef.Name, fc.Path, err)
				}
				fields, err := s.fieldsWithDerivedValues(typeDef, fc.Path, item)
				if err != nil {
					return nil, fmt.Errorf("data: %s in %s: %w", typeDef.Name, fc.Path, err)
				}
				objects = append(objects, &Object{
					Type:     typeDef.Name,
					ID:       idVal,
					Fields:   fields,
					File:     fc.Path,
					ReadOnly: fc.ReadOnly,
				})
				seenIDs[idVal] = struct{}{}
			}
			continue
		}

		idVal, _, err := deriveIdentifierValue(typeDef, fc.Single, fc.Path, s.root)
		if err != nil {
			return nil, fmt.Errorf("data: %s in %s: %w", typeDef.Name, fc.Path, err)
		}
		fields, err := s.fieldsWithDerivedValues(typeDef, fc.Path, fc.Single)
		if err != nil {
			return nil, fmt.Errorf("data: %s in %s: %w", typeDef.Name, fc.Path, err)
		}

		objects = append(objects, &Object{
			Type:     typeDef.Name,
			ID:       idVal,
			Fields:   fields,
			File:     fc.Path,
			ReadOnly: fc.ReadOnly,
		})
		seenIDs[idVal] = struct{}{}
	}

	for idx, item := range typeDef.InlineData {
		idVal, _, err := deriveIdentifierValue(typeDef, item, "", s.root)
		if err != nil {
			return nil, fmt.Errorf("data: %s inline item %d: %w", typeDef.Name, idx+1, err)
		}
		if _, exists := seenIDs[idVal]; exists {
			continue
		}
		objects = append(objects, &Object{
			Type:     typeDef.Name,
			ID:       idVal,
			Fields:   cloneMap(item),
			Inline:   true,
			ReadOnly: true,
		})
		seenIDs[idVal] = struct{}{}
	}

	return objects, nil
}

func (s *Store) findObject(typeDef *config.TypeDefinition, id string) (*objectLocation, error) {
	return s.findObjectInTypes(s.assignableTypeDefs(typeDef), typeDef.Name, id)
}

func (s *Store) findExactObject(typeDef *config.TypeDefinition, id string) (*objectLocation, error) {
	return s.findObjectInTypes([]*config.TypeDefinition{typeDef}, typeDef.Name, id)
}

func (s *Store) findHierarchyObject(typeDef *config.TypeDefinition, id string) (*objectLocation, error) {
	return s.findObjectInTypes(s.hierarchyTypeDefs(typeDef), typeDef.Name, id)
}

func (s *Store) findObjectInTypes(typeDefs []*config.TypeDefinition, queryTypeName, id string) (*objectLocation, error) {
	var found *objectLocation
	for _, typeDef := range typeDefs {
		loc, err := s.findObjectInType(typeDef, id)
		if err != nil {
			return nil, err
		}
		if loc == nil {
			continue
		}
		if found != nil {
			return nil, duplicateHierarchyLocationError(queryTypeName, id, found, loc)
		}
		found = loc
	}
	return found, nil
}

func (s *Store) findObjectInType(typeDef *config.TypeDefinition, id string) (*objectLocation, error) {
	matches, err := s.resolveIncludeMatches(typeDef)
	if err != nil {
		return nil, err
	}

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

		if fc.Multi && typeDef.Identifier.IsPath() {
			return nil, fmt.Errorf("data: %s uses identifier %q, but file %s contains multiple objects", typeDef.Name, config.PathIdentifierField, fc.Path)
		}

		if fc.Multi {
			for idx, item := range fc.Items {
				val, _, err := deriveIdentifierValue(typeDef, item, fc.Path, s.root)
				if err != nil {
					return nil, fmt.Errorf("data: %s in %s: %w", typeDef.Name, fc.Path, err)
				}
				if val == id {
					return &objectLocation{
						FilePath: match.path,
						Format:   fc.Format,
						Multi:    true,
						Index:    idx,
						ID:       val,
						Object:   cloneMap(item),
						File:     fc,
						TypeName: typeDef.Name,
						ReadOnly: fc.ReadOnly,
					}, nil
				}
			}
			continue
		}

		val, _, err := deriveIdentifierValue(typeDef, fc.Single, fc.Path, s.root)
		if err != nil {
			return nil, fmt.Errorf("data: %s in %s: %w", typeDef.Name, fc.Path, err)
		}
		if val == id {
			return &objectLocation{
				FilePath: match.path,
				Format:   fc.Format,
				Multi:    false,
				ID:       val,
				Object:   cloneMap(fc.Single),
				File:     fc,
				TypeName: typeDef.Name,
				ReadOnly: fc.ReadOnly,
			}, nil
		}
	}

	for idx, item := range typeDef.InlineData {
		val, _, err := deriveIdentifierValue(typeDef, item, "", s.root)
		if err != nil {
			return nil, fmt.Errorf("data: %s inline item %d: %w", typeDef.Name, idx+1, err)
		}
		if val == id {
			return &objectLocation{
				FilePath: "",
				Format:   formatYAML,
				Multi:    false,
				Index:    -1,
				ID:       val,
				Object:   cloneMap(item),
				File:     nil,
				TypeName: typeDef.Name,
				Inline:   true,
			}, nil
		}
	}

	return nil, nil
}

func (s *Store) assignableTypeDefs(typeDef *config.TypeDefinition) []*config.TypeDefinition {
	if s == nil || s.config == nil || typeDef == nil {
		return nil
	}
	names := s.config.AssignableTypes(typeDef.Name)
	typeDefs := make([]*config.TypeDefinition, 0, len(names))
	for _, name := range names {
		if concreteType := s.config.Types[name]; concreteType != nil {
			typeDefs = append(typeDefs, concreteType)
		}
	}
	return typeDefs
}

func (s *Store) hierarchyTypeDefs(typeDef *config.TypeDefinition) []*config.TypeDefinition {
	if s == nil || s.config == nil || typeDef == nil {
		return nil
	}

	rootName := typeDef.Name
	if len(typeDef.Ancestors) > 0 {
		rootName = typeDef.Ancestors[0]
	}
	rootType := s.config.Types[rootName]
	if rootType == nil {
		return []*config.TypeDefinition{typeDef}
	}
	return s.assignableTypeDefs(rootType)
}

func duplicateHierarchyError(queryTypeName, id string, first, second *Object) error {
	return fmt.Errorf(
		"data: duplicate identifier across assignable hierarchy for %s %q; defined by %s in %s and %s in %s",
		queryTypeName,
		id,
		first.Type,
		first.File,
		second.Type,
		second.File,
	)
}

func duplicateHierarchyLocationError(queryTypeName, id string, first, second *objectLocation) error {
	return fmt.Errorf(
		"data: duplicate identifier across assignable hierarchy for %s %q; defined by %s in %s and %s in %s",
		queryTypeName,
		id,
		first.TypeName,
		first.FilePath,
		second.TypeName,
		second.FilePath,
	)
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

		globbed, err := s.ops.Glob(absPattern)
		if err != nil {
			return nil, fmt.Errorf("data: glob %s: %w", pattern, err)
		}

		sort.Strings(globbed)

		for _, match := range globbed {
			info, err := s.ops.Stat(match)
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
