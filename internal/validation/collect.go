package validation

import (
	"fmt"
	"path/filepath"
	"sort"

	"github.com/mergewayhq/mergeway-cli/internal/config"
)

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
	seenFiles := make(map[string]struct{})
	var files []string

	for _, pattern := range typeDef.Include {
		absPattern := filepath.Join(root, filepath.Clean(pattern))
		matches, err := filepath.Glob(absPattern)
		if err != nil {
			err := Error{
				Phase:   PhaseFormat,
				Type:    typeDef.Name,
				File:    relPath(root, absPattern),
				Message: fmt.Sprintf("invalid glob pattern: %v", err),
			}
			return nil, []Error{err}
		}

		for _, match := range matches {
			if _, err := filepath.Glob(match); err != nil {
				continue
			}
			if _, ok := seenFiles[match]; ok {
				continue
			}
			seenFiles[match] = struct{}{}
			files = append(files, match)
		}
	}

	sort.Strings(files)

	var records []*rawObject
	var errs []Error

	for _, file := range files {
		parsed, err := parseDataFile(file, typeDef.Name)
		if err != nil {
			errs = append(errs, Error{
				Phase:   PhaseFormat,
				Type:    typeDef.Name,
				File:    relPath(root, file),
				Message: err.Error(),
			})
			continue
		}

		if parsed.TypeName != typeDef.Name {
			errs = append(errs, Error{
				Phase:   PhaseFormat,
				Type:    typeDef.Name,
				File:    relPath(root, file),
				Message: fmt.Sprintf("file declares type %q", parsed.TypeName),
			})
			continue
		}

		if parsed.Multi {
			for idx, item := range parsed.Items {
				records = append(records, &rawObject{
					typeDef: typeDef,
					file:    relPath(root, file),
					index:   idx,
					data:    item,
				})
			}
			continue
		}

		records = append(records, &rawObject{
			typeDef: typeDef,
			file:    relPath(root, file),
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
