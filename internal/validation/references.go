package validation

import (
	"fmt"

	"github.com/mergewayhq/mergeway-cli/internal/config"
)

func validateReferences(all map[string]*typeObjects, index *schemaIndex, cfg *config.Config) []Error {
	var errs []Error

	for typeName, typeDef := range cfg.Types {
		objects := all[typeName]
		if objects == nil {
			continue
		}

		for _, obj := range objects.objects {
			if obj.id == "" {
				continue
			}

			for fieldName, field := range typeDef.Fields {
				if isPrimitive(field.Type) || field.Type == "object" || field.Type == "enum" {
					continue
				}

				targets := collectReferenceValues(obj.data[fieldName], field.Repeated)
				if len(targets) == 0 {
					continue
				}

				targetTypeIndex := index.byType[field.Type]
				for _, refID := range targets {
					if targetTypeIndex == nil || targetTypeIndex[refID] == nil {
						errs = append(errs, Error{
							Phase:   PhaseReferences,
							Type:    obj.typeDef.Name,
							ID:      obj.id,
							File:    objectLocation(obj),
							Message: fmt.Sprintf("field %q references missing %s %q", fieldName, field.Type, refID),
						})
					}
				}
			}
		}
	}

	return errs
}

func isPrimitive(t string) bool {
	switch t {
	case "string", "integer", "number", "boolean":
		return true
	}
	return false
}
