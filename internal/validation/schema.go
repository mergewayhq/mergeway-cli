package validation

import (
	"fmt"

	"github.com/mergewayhq/mergeway-cli/internal/config"
)

func validateSchema(all map[string]*typeObjects, cfg *config.Config) (*schemaIndex, []Error) {
	index := &schemaIndex{byType: make(map[string]map[string]*rawObject)}
	var errs []Error

	for typeName, typeDef := range cfg.Types {
		objects := all[typeName]
		if objects == nil {
			continue
		}

		typeErrs := validateTypeSchema(objects.objects, typeDef, index)
		errs = append(errs, typeErrs...)
	}

	return index, errs
}

func validateTypeSchema(objects []*rawObject, typeDef *config.TypeDefinition, index *schemaIndex) []Error {
	var errs []Error

	idField := typeDef.Identifier.Field
	if index.byType[typeDef.Name] == nil {
		index.byType[typeDef.Name] = make(map[string]*rawObject)
	}

	uniqueTrack := make(map[string]map[string]string)

	for _, obj := range objects {
		hadError := false
		if obj.data == nil {
			errs = append(errs, Error{
				Phase:   PhaseSchema,
				Type:    typeDef.Name,
				File:    objectLocation(obj),
				Message: "object is empty",
			})
			hadError = true
			continue
		}

		idValue, ok := getString(obj.data, idField)
		if !ok {
			errs = append(errs, Error{
				Phase:   PhaseSchema,
				Type:    typeDef.Name,
				File:    objectLocation(obj),
				Message: fmt.Sprintf("identifier field %q must be a non-empty string", idField),
			})
			hadError = true
			continue
		}

		objID := idValue
		objIDExisting := index.byType[typeDef.Name][objID]
		if objIDExisting != nil {
			errs = append(errs, Error{
				Phase:   PhaseSchema,
				Type:    typeDef.Name,
				ID:      objID,
				File:    objectLocation(obj),
				Message: fmt.Sprintf("duplicate identifier; already defined in %s", objectLocation(objIDExisting)),
			})
			hadError = true
			continue
		}

		obj.id = objID

		for fieldName, fieldDef := range typeDef.Fields {
			fieldErrs := validateField(fieldDef, obj.data[fieldName], obj, fieldName)
			if len(fieldErrs) > 0 {
				errs = append(errs, fieldErrs...)
				hadError = true
			}

			if fieldDef.Unique {
				if uniqueTrack[fieldName] == nil {
					uniqueTrack[fieldName] = make(map[string]string)
				}
				if value, exists := obj.data[fieldName]; exists {
					key := fmt.Sprintf("%v", value)
					if key != "" {
						if firstID, seen := uniqueTrack[fieldName][key]; seen {
							errs = append(errs, Error{
								Phase:   PhaseSchema,
								Type:    typeDef.Name,
								ID:      objID,
								File:    objectLocation(obj),
								Message: fmt.Sprintf("field %q must be unique; conflict with %s", fieldName, firstID),
							})
							hadError = true
							continue
						}
						uniqueTrack[fieldName][key] = objID
					}
				}
			}
		}

		if !hadError {
			index.byType[typeDef.Name][objID] = obj
		}
	}

	return errs
}

func validateField(field *config.FieldDefinition, value any, obj *rawObject, fieldName string) []Error {
	var errs []Error

	if value == nil {
		if field.Required {
			errs = append(errs, Error{
				Phase:   PhaseSchema,
				Type:    obj.typeDef.Name,
				ID:      obj.id,
				File:    objectLocation(obj),
				Message: fmt.Sprintf("missing required field %q", fieldName),
			})
		}
		return errs
	}

	if field.Repeated {
		slice, ok := value.([]any)
		if !ok {
			errs = append(errs, Error{
				Phase:   PhaseSchema,
				Type:    obj.typeDef.Name,
				ID:      obj.id,
				File:    objectLocation(obj),
				Message: fmt.Sprintf("field %q must be an array", fieldName),
			})
			return errs
		}

		for idx, item := range slice {
			err := validateFieldValue(field, item, obj, fmt.Sprintf("%s[%d]", fieldName, idx))
			if err != nil {
				errs = append(errs, *err)
			}
		}
		return errs
	}

	err := validateFieldValue(field, value, obj, fieldName)
	if err != nil {
		errs = append(errs, *err)
	}

	return errs
}
