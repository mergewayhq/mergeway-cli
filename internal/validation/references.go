package validation

import (
	"fmt"
	"strings"

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
				if field == nil || !field.IsReference() {
					continue
				}

				targets := collectReferenceValues(obj.data[fieldName], field.Repeated)
				if len(targets) == 0 {
					continue
				}

				for _, refID := range targets {
					matches := resolveReferenceTypes(index, field.ReferenceTypes, refID)
					if len(matches) == 0 {
						errs = append(errs, Error{
							Phase:   PhaseReferences,
							Type:    obj.typeDef.Name,
							ID:      obj.id,
							File:    objectLocation(obj),
							Message: fmt.Sprintf("field %q references missing %s %q", fieldName, field.ReferenceLabel(), refID),
						})
						continue
					}
					if len(matches) > 1 {
						errs = append(errs, Error{
							Phase:   PhaseReferences,
							Type:    obj.typeDef.Name,
							ID:      obj.id,
							File:    objectLocation(obj),
							Message: fmt.Sprintf("field %q reference %q is ambiguous across %s", fieldName, refID, joinReferenceTypes(matches)),
						})
					}
				}
			}
		}
	}

	return errs
}

func resolveReferenceTypes(index *schemaIndex, refTypes []string, refID string) []string {
	var matches []string
	for _, refType := range refTypes {
		targetTypeIndex := index.byType[refType]
		if targetTypeIndex == nil || targetTypeIndex[refID] == nil {
			continue
		}
		matches = append(matches, refType)
	}
	return matches
}

func joinReferenceTypes(refTypes []string) string {
	return strings.Join(refTypes, " | ")
}
