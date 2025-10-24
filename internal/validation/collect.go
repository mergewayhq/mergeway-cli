package validation

import (
	"fmt"
	"path/filepath"
	"sort"

	"github.com/mergewayhq/mergeway-cli/internal/config"
)

type includeMatch struct {
	include config.IncludeDefinition
	path    string
}

func collectObjects(root string, cfg *config.Config) (map[string]*typeObjects, []Error) {
	result := make(map[string]*typeObjects)
	var errs []Error

	for _, typeDef := range cfg.Types {
		records, recordErrs := loadTypeObjects(root, typeDef)
		if len(recordErrs) > 0 {
			errs = append(errs, recordErrs...)
		}

		if len(records) > 0 {
			result[typeDef.Name] = &typeObjects{objects: records}
		} else if result[typeDef.Name] == nil {
			result[typeDef.Name] = &typeObjects{objects: []*rawObject{}}
		}
	}

	return result, errs
}

func loadTypeObjects(root string, typeDef *config.TypeDefinition) ([]*rawObject, []Error) {
	matches, collectErrs := resolveIncludeMatches(root, typeDef)
	if len(collectErrs) > 0 {
		return nil, collectErrs
	}

	var records []*rawObject
	var errs []Error

	for _, match := range matches {
		parsed, err := parseDataFile(match.path, typeDef.Name, match.include.Selector)
		if err != nil {
			errs = append(errs, Error{
				Phase:   PhaseFormat,
				Type:    typeDef.Name,
				File:    relPath(root, match.path),
				Message: err.Error(),
			})
			continue
		}

		if parsed.TypeName != typeDef.Name {
			errs = append(errs, Error{
				Phase:   PhaseFormat,
				Type:    typeDef.Name,
				File:    relPath(root, match.path),
				Message: fmt.Sprintf("file declares type %q", parsed.TypeName),
			})
			continue
		}

		if parsed.Multi {
			for idx, item := range parsed.Items {
				records = append(records, &rawObject{
					typeDef: typeDef,
					file:    relPath(root, match.path),
					index:   idx,
					data:    item,
				})
			}
			continue
		}

		records = append(records, &rawObject{
			typeDef: typeDef,
			file:    relPath(root, match.path),
			index:   -1,
			data:    parsed.Single,
		})
	}

	source := relPath(root, typeDef.Source)
	for idx, item := range typeDef.InlineData {
		label := source
		if label == "" {
			label = typeDef.Source
		}
		label = fmt.Sprintf("%s (inline %d)", label, idx+1)
		records = append(records, &rawObject{
			typeDef: typeDef,
			file:    label,
			index:   -1,
			data:    cloneMap(item),
		})
	}

	return records, errs
}

func resolveIncludeMatches(root string, typeDef *config.TypeDefinition) ([]includeMatch, []Error) {
	seen := make(map[string]struct{})
	var matches []includeMatch

	for _, include := range typeDef.Include {
		pattern := include.Path
		if pattern == "" {
			continue
		}

		absPattern := filepath.Join(root, filepath.Clean(pattern))
		globbed, err := filepath.Glob(absPattern)
		if err != nil {
			return nil, []Error{Error{
				Phase:   PhaseFormat,
				Type:    typeDef.Name,
				File:    relPath(root, absPattern),
				Message: fmt.Sprintf("invalid glob pattern: %v", err),
			}}
		}

		sort.Strings(globbed)

		for _, path := range globbed {
			key := include.Path + "\x00" + include.Selector + "\x00" + path
			if _, ok := seen[key]; ok {
				continue
			}
			seen[key] = struct{}{}
			matches = append(matches, includeMatch{include: include, path: path})
		}
	}

	return matches, nil
}
