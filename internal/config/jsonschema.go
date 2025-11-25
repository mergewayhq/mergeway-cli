package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
)

func loadJSONSchemaFields(typeName, schemaPath, displayPath string) (map[string]*FieldDefinition, []string, error) {
	data, err := os.ReadFile(schemaPath)
	if err != nil {
		return nil, nil, fmt.Errorf("config: type %q read json_schema %s: %w", typeName, displayPath, err)
	}

	dec := json.NewDecoder(bytes.NewReader(data))
	dec.UseNumber()

	var document any
	if err := dec.Decode(&document); err != nil {
		return nil, nil, fmt.Errorf("config: type %q parse json_schema %s: %w", typeName, displayPath, err)
	}

	rootMap, ok := document.(map[string]any)
	if !ok {
		return nil, nil, fmt.Errorf("config: type %q json_schema %s must define an object schema", typeName, displayPath)
	}

	resolvedRoot, err := resolveSchemaMap(rootMap, document, displayPath, typeName)
	if err != nil {
		return nil, nil, err
	}

	fields, order, err := convertObjectSchema(typeName, resolvedRoot, document, displayPath, typeName)
	if err != nil {
		return nil, nil, err
	}

	return fields, order, nil
}

func convertObjectSchema(typeName string, schema map[string]any, document any, displayPath, context string) (map[string]*FieldDefinition, []string, error) {
	schemaType, err := extractSchemaType(schema["type"], displayPath, context)
	if err != nil {
		return nil, nil, err
	}
	if schemaType != "" && schemaType != "object" {
		return nil, nil, fmt.Errorf("config: type %q json_schema %s must describe an object (type=%s)", typeName, displayPath, schemaType)
	}
	return convertObjectProperties(schema, document, displayPath, context)
}

func convertObjectProperties(schema map[string]any, document any, displayPath, context string) (map[string]*FieldDefinition, []string, error) {
	rawProps, ok := schema["properties"]
	if !ok {
		return nil, nil, fmt.Errorf("config: %s in json_schema %s missing properties", context, displayPath)
	}

	props, ok := rawProps.(map[string]any)
	if !ok {
		return nil, nil, fmt.Errorf("config: %s in json_schema %s must declare properties as an object", context, displayPath)
	}

	requiredSet, err := parseRequiredSet(schema["required"], displayPath, context)
	if err != nil {
		return nil, nil, err
	}

	if len(props) == 0 {
		return nil, nil, nil
	}

	names := make([]string, 0, len(props))
	for name := range props {
		names = append(names, name)
	}
	sort.Strings(names)

	fields := make(map[string]*FieldDefinition, len(props))
	order := make([]string, 0, len(names))

	for _, name := range names {
		if name == "" {
			return nil, nil, fmt.Errorf("config: %s in json_schema %s has unnamed property", context, displayPath)
		}
		if !isValidIdentifier(name) {
			return nil, nil, fmt.Errorf("config: %s in json_schema %s has invalid property identifier %q", context, displayPath, name)
		}
		fieldSchema, err := schemaMapFromValue(props[name], displayPath, fmt.Sprintf("%s.%s", context, name))
		if err != nil {
			return nil, nil, err
		}
		field, err := buildFieldDefinitionFromSchema(name, fieldSchema, document, displayPath, fmt.Sprintf("%s.%s", context, name))
		if err != nil {
			return nil, nil, err
		}
		if requiredSet[name] {
			field.Required = true
		}
		fields[name] = field
		order = append(order, name)
	}

	for name := range requiredSet {
		if _, ok := fields[name]; !ok {
			return nil, nil, fmt.Errorf("config: %s in json_schema %s marks unknown property %q as required", context, displayPath, name)
		}
	}

	return fields, order, nil
}

func buildFieldDefinitionFromSchema(name string, schema map[string]any, document any, displayPath, context string) (*FieldDefinition, error) {
	resolved, err := resolveSchemaMap(schema, document, displayPath, context)
	if err != nil {
		return nil, err
	}

	field := &FieldDefinition{
		Name:        name,
		Description: stringValue(resolved["description"]),
		Format:      stringValue(resolved["format"]),
		Pattern:     stringValue(resolved["pattern"]),
	}
	if defaultValue, ok := resolved["default"]; ok {
		field.Default = defaultValue
	}

	if enums, err := extractEnumFromOneOf(resolved, document, displayPath, context); err != nil {
		return nil, err
	} else if len(enums) > 0 {
		field.Type = "enum"
		field.Enum = enums
		return field, nil
	}

	if constValue, ok := resolved["const"]; ok {
		value, ok := constValue.(string)
		if !ok {
			return nil, fmt.Errorf("config: field %s in json_schema %s has non-string const value", context, displayPath)
		}
		field.Type = "enum"
		field.Enum = []string{value}
		return field, nil
	}

	if enums, err := extractEnum(resolved["enum"], displayPath, context); err != nil {
		return nil, err
	} else if len(enums) > 0 {
		field.Type = "enum"
		field.Enum = enums
		return field, nil
	}

	refType := stringValue(resolved["x-reference-type"])

	schemaType, err := extractSchemaType(resolved["type"], displayPath, context)
	if err != nil {
		return nil, err
	}
	if schemaType == "" {
		if _, ok := resolved["properties"]; ok {
			schemaType = "object"
		}
	}

	if refType != "" {
		field.Type = refType
		return field, nil
	}

	switch schemaType {
	case "string", "integer", "number", "boolean":
		field.Type = schemaType
		return field, nil
	case "object":
		props, order, err := convertObjectProperties(resolved, document, displayPath, context)
		if err != nil {
			return nil, err
		}
		field.Type = "object"
		field.Properties = props
		field.PropertyOrder = order
		return field, nil
	case "array":
		rawItems, ok := resolved["items"]
		if !ok || rawItems == nil {
			return nil, fmt.Errorf("config: field %s in json_schema %s defines an array without 'items'", context, displayPath)
		}
		itemSchema, err := schemaMapFromValue(rawItems, displayPath, context+"[]")
		if err != nil {
			return nil, err
		}
		itemField, err := buildFieldDefinitionFromSchema(name, itemSchema, document, displayPath, context+"[]")
		if err != nil {
			return nil, err
		}
		field.Type = itemField.Type
		field.Format = itemField.Format
		field.Enum = append([]string(nil), itemField.Enum...)
		field.Properties = itemField.Properties
		field.PropertyOrder = itemField.PropertyOrder
		field.Pattern = itemField.Pattern
		field.Repeated = true
		return field, nil
	case "":
		return nil, fmt.Errorf("config: field %s in json_schema %s missing type", context, displayPath)
	default:
		return nil, fmt.Errorf("config: field %s in json_schema %s has unsupported type %q", context, displayPath, schemaType)
	}
}

func resolveSchemaMap(schema map[string]any, document any, displayPath, context string) (map[string]any, error) {
	if schema == nil {
		return nil, fmt.Errorf("config: %s in json_schema %s is null", context, displayPath)
	}
	ref, _ := schema["$ref"].(string)
	if ref == "" {
		return schema, nil
	}
	target, err := resolveJSONPointer(document, ref, displayPath)
	if err != nil {
		return nil, fmt.Errorf("config: %s in json_schema %s: %v", context, displayPath, err)
	}
	targetMap, ok := target.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("config: %s in json_schema %s resolved $ref %q to non-object", context, displayPath, ref)
	}
	merged := cloneInlineMap(targetMap)
	for key, value := range schema {
		if key == "$ref" {
			continue
		}
		merged[key] = value
	}
	return merged, nil
}

func resolveJSONPointer(document any, ref string, displayPath string) (any, error) {
	if ref == "" {
		return nil, fmt.Errorf("json_schema %s contains empty $ref", displayPath)
	}
	if !strings.HasPrefix(ref, "#") {
		return nil, fmt.Errorf("json_schema %s references external $ref %q (only in-document refs are supported)", displayPath, ref)
	}
	if ref == "#" {
		return document, nil
	}

	path := strings.TrimPrefix(ref, "#")
	if path == "" {
		return document, nil
	}

	segments := strings.Split(path, "/")
	current := document
	for _, rawSegment := range segments {
		if rawSegment == "" {
			continue
		}
		segment := strings.ReplaceAll(strings.ReplaceAll(rawSegment, "~1", "/"), "~0", "~")
		switch node := current.(type) {
		case map[string]any:
			val, ok := node[segment]
			if !ok {
				return nil, fmt.Errorf("json_schema %s could not resolve pointer %q", displayPath, ref)
			}
			current = val
		case []any:
			index, err := strconv.Atoi(segment)
			if err != nil || index < 0 || index >= len(node) {
				return nil, fmt.Errorf("json_schema %s has invalid array index in pointer %q", displayPath, ref)
			}
			current = node[index]
		default:
			return nil, fmt.Errorf("json_schema %s resolved pointer %q into non-container value", displayPath, ref)
		}
	}

	return current, nil
}

func schemaMapFromValue(value any, displayPath, context string) (map[string]any, error) {
	if value == nil {
		return nil, fmt.Errorf("config: field %s in json_schema %s is null", context, displayPath)
	}
	switch node := value.(type) {
	case map[string]any:
		return node, nil
	default:
		return nil, fmt.Errorf("config: field %s in json_schema %s must be an object", context, displayPath)
	}
}

func parseRequiredSet(value any, displayPath, context string) (map[string]bool, error) {
	if value == nil {
		return map[string]bool{}, nil
	}
	rawList, ok := value.([]any)
	if !ok {
		return nil, fmt.Errorf("config: %s in json_schema %s has invalid required list", context, displayPath)
	}
	result := make(map[string]bool, len(rawList))
	for i, item := range rawList {
		name, ok := item.(string)
		if !ok || name == "" {
			return nil, fmt.Errorf("config: %s in json_schema %s has non-string required entry at index %d", context, displayPath, i)
		}
		result[name] = true
	}
	return result, nil
}

func extractSchemaType(value any, displayPath, context string) (string, error) {
	if value == nil {
		return "", nil
	}
	switch typed := value.(type) {
	case string:
		return strings.TrimSpace(typed), nil
	case []any:
		types := make([]string, 0, len(typed))
		for _, item := range typed {
			str, ok := item.(string)
			if !ok {
				return "", fmt.Errorf("config: field %s in json_schema %s has non-string type entry", context, displayPath)
			}
			if str == "null" {
				continue
			}
			types = append(types, strings.TrimSpace(str))
		}
		if len(types) == 0 {
			return "", nil
		}
		if len(types) == 1 {
			return types[0], nil
		}
		return "", fmt.Errorf("config: field %s in json_schema %s combines multiple types (%s)", context, displayPath, strings.Join(types, ", "))
	default:
		return "", fmt.Errorf("config: field %s in json_schema %s uses unsupported type declaration", context, displayPath)
	}
}

func extractEnum(value any, displayPath, context string) ([]string, error) {
	if value == nil {
		return nil, nil
	}
	rawList, ok := value.([]any)
	if !ok {
		return nil, fmt.Errorf("config: field %s in json_schema %s has non-array enum", context, displayPath)
	}
	result := make([]string, 0, len(rawList))
	for i, item := range rawList {
		str, ok := item.(string)
		if !ok {
			return nil, fmt.Errorf("config: field %s in json_schema %s has non-string enum entry at index %d", context, displayPath, i)
		}
		result = append(result, str)
	}
	return result, nil
}

func extractEnumFromOneOf(schema map[string]any, document any, displayPath, context string) ([]string, error) {
	rawOneOf, ok := schema["oneOf"]
	if !ok || rawOneOf == nil {
		return nil, nil
	}

	list, ok := rawOneOf.([]any)
	if !ok {
		return nil, fmt.Errorf("config: field %s in json_schema %s has non-array oneOf", context, displayPath)
	}

	result := make([]string, 0, len(list))
	for idx, option := range list {
		optionMap, err := schemaMapFromValue(option, displayPath, fmt.Sprintf("%s.oneOf[%d]", context, idx))
		if err != nil {
			return nil, err
		}
		resolved, err := resolveSchemaMap(optionMap, document, displayPath, fmt.Sprintf("%s.oneOf[%d]", context, idx))
		if err != nil {
			return nil, err
		}
		if constValue, ok := resolved["const"]; ok {
			value, ok := constValue.(string)
			if !ok {
				return nil, fmt.Errorf("config: %s.oneOf[%d] in json_schema %s must use string const values", context, idx, displayPath)
			}
			result = append(result, value)
			continue
		}
		if enums, err := extractEnum(resolved["enum"], displayPath, fmt.Sprintf("%s.oneOf[%d]", context, idx)); err != nil {
			return nil, err
		} else if len(enums) == 1 {
			result = append(result, enums[0])
			continue
		} else if len(enums) > 1 {
			return nil, fmt.Errorf("config: %s.oneOf[%d] in json_schema %s defines multiple enum entries", context, idx, displayPath)
		}
		return nil, fmt.Errorf("config: %s.oneOf[%d] in json_schema %s must declare a const or single-entry enum", context, idx, displayPath)
	}

	return result, nil
}

func stringValue(value any) string {
	str, _ := value.(string)
	return strings.TrimSpace(str)
}
