package validation

import (
	"fmt"

	"github.com/mergewayhq/mergeway-cli/internal/config"
)

func validateFieldValue(field *config.FieldDefinition, value any, obj *rawObject, fieldName string) *Error {
	switch field.Type {
	case "string":
		if _, ok := value.(string); !ok {
			return typeError(obj, fieldName, "string")
		}
	case "integer":
		switch value.(type) {
		case int, int64, uint64:
		default:
			return typeError(obj, fieldName, "integer")
		}
	case "number":
		switch value.(type) {
		case int, int64, uint64, float64:
		default:
			return typeError(obj, fieldName, "number")
		}
	case "boolean":
		if _, ok := value.(bool); !ok {
			return typeError(obj, fieldName, "boolean")
		}
	case "enum":
		str, ok := value.(string)
		if !ok {
			return typeError(obj, fieldName, "enum value (string)")
		}
		if len(field.Enum) > 0 {
			valid := false
			for _, candidate := range field.Enum {
				if candidate == str {
					valid = true
					break
				}
			}
			if !valid {
				return &Error{
					Phase:   PhaseSchema,
					Type:    obj.typeDef.Name,
					ID:      obj.id,
					File:    objectLocation(obj),
					Message: fmt.Sprintf("field %q must be one of %v", fieldName, field.Enum),
				}
			}
		}
	case "object":
		child, ok := value.(map[string]any)
		if !ok {
			return typeError(obj, fieldName, "object")
		}
		for name, def := range field.Properties {
			childErrs := validateField(def, child[name], obj, fmt.Sprintf("%s.%s", fieldName, name))
			if len(childErrs) > 0 {
				return &childErrs[0]
			}
		}
	default:
		if _, ok := value.(string); !ok {
			return typeError(obj, fieldName, "string")
		}
	}

	return nil
}

func typeError(obj *rawObject, field, expected string) *Error {
	return &Error{
		Phase:   PhaseSchema,
		Type:    obj.typeDef.Name,
		ID:      obj.id,
		File:    objectLocation(obj),
		Message: fmt.Sprintf("field %q must be %s", field, expected),
	}
}

func collectReferenceValues(value any, repeated bool) []string {
	if value == nil {
		return nil
	}

	if repeated {
		slice, ok := value.([]any)
		if !ok {
			return nil
		}

		var values []string
		for _, item := range slice {
			if str, ok := item.(string); ok && str != "" {
				values = append(values, str)
			}
		}
		return values
	}

	if str, ok := value.(string); ok && str != "" {
		return []string{str}
	}

	return nil
}

func objectLocation(obj *rawObject) string {
	if obj == nil {
		return ""
	}
	if obj.index >= 0 {
		return fmt.Sprintf("%s (item %d)", obj.file, obj.index+1)
	}
	return obj.file
}
