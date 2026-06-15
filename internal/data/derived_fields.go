package data

import (
	"fmt"

	"github.com/mergewayhq/mergeway-cli/internal/config"
)

func (s *Store) fieldsWithDerivedValues(typeDef *config.TypeDefinition, sourcePath string, fields map[string]any) (map[string]any, error) {
	enriched := cloneMap(fields)
	if typeDef == nil || sourcePath == "" {
		return enriched, nil
	}

	pathID, err := relativePathIdentifier(s.root, sourcePath, true)
	if err != nil {
		return nil, err
	}

	for name, field := range typeDef.Fields {
		if field == nil || !field.Source.IsPathDerived() {
			continue
		}
		value, err := config.DerivePathSourceValue(field.Source, pathID)
		if err != nil {
			return nil, fmt.Errorf("derive field %q: %w", name, err)
		}
		enriched[name] = value
	}

	return enriched, nil
}

func removeDerivedFieldKeys(typeDef *config.TypeDefinition, fields map[string]any) {
	if typeDef == nil || fields == nil {
		return
	}
	for name, field := range typeDef.Fields {
		if field == nil || !field.Source.IsPathDerived() {
			continue
		}
		delete(fields, name)
	}
}

func cleanFieldsForType(typeDef *config.TypeDefinition, fields map[string]any) map[string]any {
	data := cloneMap(fields)
	removeSystemKeys(data)
	removeDerivedFieldKeys(typeDef, data)
	return data
}
