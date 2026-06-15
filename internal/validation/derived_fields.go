package validation

import (
	"fmt"

	"github.com/mergewayhq/mergeway-cli/internal/config"
)

func fieldsWithDerivedValues(typeDef *config.TypeDefinition, sourcePath string, fields map[string]any) (map[string]any, error) {
	enriched := cloneMap(fields)
	if typeDef == nil || sourcePath == "" {
		return enriched, nil
	}

	for name, field := range typeDef.Fields {
		if field == nil || !field.Source.IsPathDerived() {
			continue
		}
		value, err := config.DerivePathSourceValue(field.Source, sourcePath)
		if err != nil {
			return nil, fmt.Errorf("derive field %q: %w", name, err)
		}
		enriched[name] = value
	}

	return enriched, nil
}
