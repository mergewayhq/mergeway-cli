package validation

import (
	"fmt"
	"net/mail"
	"net/url"
	"regexp"
	"strings"
	"sync"

	"github.com/mergewayhq/mergeway-cli/internal/config"
	"github.com/mergewayhq/mergeway-cli/internal/scalar"
)

var (
	patternCache     sync.Map
	formatValidators = map[string]func(string) bool{
		"email": isValidEmail,
		"uri":   isValidURI,
		"url":   isValidURI,
	}
)

func validateFieldValue(field *config.FieldDefinition, value any, obj *rawObject, fieldName string) *Error {
	switch field.Type {
	case "string":
		str, ok := value.(string)
		if !ok {
			return typeError(obj, fieldName, "string")
		}
		if err := enforceStringConstraints(field, str, obj, fieldName); err != nil {
			return err
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
		if err := enforceStringConstraints(field, str, obj, fieldName); err != nil {
			return err
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
		if _, ok := scalar.AsString(value); !ok {
			return typeError(obj, fieldName, "string or number")
		}
	}

	return nil
}

func enforceStringConstraints(field *config.FieldDefinition, value string, obj *rawObject, fieldName string) *Error {
	if field.Pattern != "" {
		ok, err := matchPattern(field.Pattern, value)
		if err != nil {
			return &Error{
				Phase:   PhaseSchema,
				Type:    obj.typeDef.Name,
				ID:      obj.id,
				File:    objectLocation(obj),
				Message: fmt.Sprintf("field %q has invalid pattern %q: %v", fieldName, field.Pattern, err),
			}
		}
		if !ok {
			return &Error{
				Phase:   PhaseSchema,
				Type:    obj.typeDef.Name,
				ID:      obj.id,
				File:    objectLocation(obj),
				Message: fmt.Sprintf("field %q must match pattern %q", fieldName, field.Pattern),
			}
		}
	}

	if field.Format != "" {
		if validator := formatValidators[strings.ToLower(field.Format)]; validator != nil {
			if !validator(value) {
				return &Error{
					Phase:   PhaseSchema,
					Type:    obj.typeDef.Name,
					ID:      obj.id,
					File:    objectLocation(obj),
					Message: fmt.Sprintf("field %q must satisfy format %q", fieldName, field.Format),
				}
			}
		}
	}

	return nil
}

func matchPattern(pattern, value string) (bool, error) {
	if pattern == "" {
		return true, nil
	}
	if cached, ok := patternCache.Load(pattern); ok {
		return cached.(*regexp.Regexp).MatchString(value), nil
	}
	re, err := regexp.Compile(pattern)
	if err != nil {
		return false, err
	}
	patternCache.Store(pattern, re)
	return re.MatchString(value), nil
}

func isValidEmail(value string) bool {
	addr, err := mail.ParseAddress(value)
	if err != nil {
		return false
	}
	return addr.Address == value
}

func isValidURI(value string) bool {
	u, err := url.ParseRequestURI(value)
	if err != nil {
		return false
	}
	return u.Scheme != "" && u.Host != ""
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
			if str, ok := scalar.AsString(item); ok {
				values = append(values, str)
			}
		}
		return values
	}

	if str, ok := scalar.AsString(value); ok {
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
