package validation

import (
	"encoding/json"
	"fmt"
	"sort"

	"github.com/mergewayhq/mergeway-cli/internal/config"
)

func validateSchema(all map[string]*typeObjects, cfg *config.Config) (*schemaIndex, []Error) {
	index := &schemaIndex{
		byType:       make(map[string]map[string]*rawObject),
		byAssignable: make(map[string]map[string][]*rawObject),
	}
	var errs []Error

	typeNames := sortedTypeNames(cfg)
	for _, typeName := range typeNames {
		typeDef := cfg.Types[typeName]
		objects := all[typeName]
		if objects == nil {
			continue
		}

		typeErrs := validateTypeSchema(objects.objects, typeDef, index)
		errs = append(errs, typeErrs...)
	}

	errs = append(errs, buildAssignableIndex(index, cfg)...)

	return index, errs
}

func validateTypeSchema(objects []*rawObject, typeDef *config.TypeDefinition, index *schemaIndex) []Error {
	var errs []Error

	if index.byType[typeDef.Name] == nil {
		index.byType[typeDef.Name] = make(map[string]*rawObject)
	}

	// Track seen values per field to surface uniqueness errors, normalizing values
	// via fmt.Sprintf since fields may be typed differently (string vs integer).
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

		idValue, err := identifierForObject(obj)
		if err != nil {
			errs = append(errs, Error{
				Phase:   PhaseSchema,
				Type:    typeDef.Name,
				File:    objectLocation(obj),
				Message: err.Error(),
			})
			hadError = true
			continue
		}

		objID := idValue
		if pat := typeDef.Identifier.Pattern; pat != "" {
			if ok, err := matchPattern(pat, objID); err != nil {
				errs = append(errs, Error{
					Phase:   PhaseSchema,
					Type:    typeDef.Name,
					ID:      objID,
					File:    objectLocation(obj),
					Message: fmt.Sprintf("identifier pattern %q is invalid: %v", pat, err),
				})
				hadError = true
				continue
			} else if !ok {
				errs = append(errs, Error{
					Phase:   PhaseSchema,
					Type:    typeDef.Name,
					ID:      objID,
					File:    objectLocation(obj),
					Message: fmt.Sprintf("identifier must match pattern %q", pat),
				})
				hadError = true
				continue
			}
		}

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
					key := normalizedUniqueKey(value)
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

func buildAssignableIndex(index *schemaIndex, cfg *config.Config) []Error {
	if index == nil || cfg == nil {
		return nil
	}

	var errs []Error
	typeNames := sortedTypeNames(cfg)
	for _, typeName := range typeNames {
		typeIndex := index.byType[typeName]
		if len(typeIndex) == 0 {
			continue
		}

		ids := make([]string, 0, len(typeIndex))
		for id := range typeIndex {
			ids = append(ids, id)
		}
		sort.Strings(ids)

		for _, id := range ids {
			obj := typeIndex[id]
			if obj == nil || obj.typeDef == nil {
				continue
			}

			assignableTo := append([]string{obj.typeDef.Name}, obj.typeDef.Ancestors...)
			var conflict *rawObject
			for _, queryType := range assignableTo {
				if matches := index.byAssignable[queryType][id]; len(matches) > 0 {
					conflict = matches[0]
					break
				}
			}
			if conflict != nil {
				errs = append(errs, Error{
					Phase:   PhaseSchema,
					Type:    obj.typeDef.Name,
					ID:      id,
					File:    objectLocation(obj),
					Message: fmt.Sprintf("duplicate identifier across assignable hierarchy; already defined as %s in %s", conflict.typeDef.Name, objectLocation(conflict)),
				})
			}

			for _, queryType := range assignableTo {
				if index.byAssignable[queryType] == nil {
					index.byAssignable[queryType] = make(map[string][]*rawObject)
				}
				index.byAssignable[queryType][id] = append(index.byAssignable[queryType][id], obj)
			}
		}
	}

	return errs
}

func sortedTypeNames(cfg *config.Config) []string {
	if cfg == nil {
		return nil
	}
	names := make([]string, 0, len(cfg.Types))
	for name := range cfg.Types {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func normalizedUniqueKey(value any) string {
	switch v := value.(type) {
	case string:
		return v
	case fmt.Stringer:
		return v.String()
	case bool:
		if v {
			return "true"
		}
		return "false"
	case int, int8, int16, int32, int64:
		return fmt.Sprintf("%d", v)
	case uint, uint8, uint16, uint32, uint64:
		return fmt.Sprintf("%d", v)
	case float32, float64:
		return fmt.Sprintf("%g", v)
	case []any, map[string]any:
		if data, err := json.Marshal(v); err == nil {
			return string(data)
		}
	}
	return fmt.Sprintf("%v", value)
}

func validateField(field *config.FieldDefinition, value any, obj *rawObject, fieldName string) []Error {
	var errs []Error

	if value == nil {
		if field.Default != nil {
			value = field.Default
		} else {
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
