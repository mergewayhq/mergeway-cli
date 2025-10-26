package config

import (
	"errors"
	"fmt"
)

func normalizeAggregate(agg *aggregateConfig) (*Config, error) {
	if agg == nil {
		return nil, errors.New("config: missing aggregate configuration")
	}

	if !agg.VersionSet {
		return nil, errors.New("config: mergeway.version is required")
	}

	if agg.Version != CurrentVersion {
		return nil, fmt.Errorf("config: unsupported mergeway.version %d", agg.Version)
	}

	result := &Config{
		Version: agg.Version,
		Types:   make(map[string]*TypeDefinition),
	}

	for name, rawType := range agg.Entities {
		if !isValidTypeName(name) {
			return nil, fmt.Errorf("config: invalid type identifier %q in %s", name, rawType.Source)
		}

		typeDef, err := normalizeTypeDefinition(rawType)
		if err != nil {
			return nil, err
		}

		result.Types[name] = typeDef
	}

	if err := validateFieldReferences(result.Types); err != nil {
		return nil, err
	}

	return result, nil
}

func normalizeTypeDefinition(rawType rawTypeWithSource) (*TypeDefinition, error) {
	spec := rawType.Spec

	if !spec.Identifier.set || spec.Identifier.Field == "" {
		return nil, fmt.Errorf("config: type %q missing identifier in %s", rawType.Name, rawType.Source)
	}

	if !isValidIdentifier(spec.Identifier.Field) {
		return nil, fmt.Errorf("config: type %q has invalid identifier field %q", rawType.Name, spec.Identifier.Field)
	}

	if len(spec.Include) == 0 && len(spec.Data) == 0 {
		return nil, fmt.Errorf("config: type %q must declare at least one include or provide inline data", rawType.Name)
	}

	fields := make(map[string]*FieldDefinition, len(spec.Fields))

	for fieldName, rawField := range spec.Fields {
		if fieldName == "" {
			return nil, fmt.Errorf("config: type %q has unnamed field", rawType.Name)
		}

		if !isValidIdentifier(fieldName) {
			return nil, fmt.Errorf("config: type %q has invalid field identifier %q", rawType.Name, fieldName)
		}

		fieldDef, err := normalizeFieldDefinition(fieldName, rawField, rawType.Name)
		if err != nil {
			return nil, err
		}

		fields[fieldName] = fieldDef
	}

	inlineData := cloneInlineData(spec.Data)

	return &TypeDefinition{
		Name:   rawType.Name,
		Source: rawType.Source,
		Identifier: IdentifierDefinition{
			Field:     spec.Identifier.Field,
			Generated: spec.Identifier.Generated,
			Pattern:   spec.Identifier.Pattern,
		},
		Include:    normalizeIncludeDirectives(spec.Include),
		Fields:     fields,
		InlineData: inlineData,
	}, nil
}

func normalizeFieldDefinition(name string, raw rawFieldDefinition, typeName string) (*FieldDefinition, error) {
	if raw.Type == "" {
		return nil, fmt.Errorf("config: field %s.%s missing type", typeName, name)
	}

	if raw.Repeated && raw.Unique != nil && *raw.Unique {
		return nil, fmt.Errorf("config: field %s.%s cannot declare unique=true when repeated", typeName, name)
	}

	if raw.Properties != nil && raw.Type != "object" {
		return nil, fmt.Errorf("config: field %s.%s defines properties but type is %q", typeName, name, raw.Type)
	}

	properties := make(map[string]*FieldDefinition)
	if raw.Type == "object" {
		for propName, propField := range raw.Properties {
			if !isValidIdentifier(propName) {
				return nil, fmt.Errorf("config: field %s.%s has invalid property identifier %q", typeName, name, propName)
			}

			child, err := normalizeFieldDefinition(propName, propField, fmt.Sprintf("%s.%s", typeName, name))
			if err != nil {
				return nil, err
			}

			properties[propName] = child
		}
	}

	var unique bool
	if raw.Unique != nil {
		unique = *raw.Unique
	}

	return &FieldDefinition{
		Name:       name,
		Type:       raw.Type,
		Required:   raw.Required,
		Repeated:   raw.Repeated,
		Format:     raw.Format,
		Enum:       append([]string(nil), raw.Enum...),
		Default:    raw.Default,
		Properties: properties,
		Unique:     unique,
		Computed:   raw.Computed,
		Pattern:    raw.Pattern,
	}, nil
}

var primitiveTypes = map[string]struct{}{
	"string":  {},
	"integer": {},
	"number":  {},
	"boolean": {},
	"enum":    {},
	"object":  {},
}

func validateFieldReferences(types map[string]*TypeDefinition) error {
	for _, typeDef := range types {
		for _, field := range typeDef.Fields {
			if err := validateFieldReference(typeDef.Name, field, types); err != nil {
				return err
			}
		}
	}
	return nil
}

func validateFieldReference(typeName string, field *FieldDefinition, types map[string]*TypeDefinition) error {
	if field == nil {
		return nil
	}

	if _, ok := primitiveTypes[field.Type]; !ok {
		if !isValidTypeName(field.Type) {
			return fmt.Errorf("config: field %s.%s references invalid type name %q", typeName, field.Name, field.Type)
		}

		if _, ok := types[field.Type]; !ok {
			return fmt.Errorf("config: field %s.%s references unknown type %q", typeName, field.Name, field.Type)
		}
	}

	for _, child := range field.Properties {
		if err := validateFieldReference(fmt.Sprintf("%s.%s", typeName, field.Name), child, types); err != nil {
			return err
		}
	}

	return nil
}

func cloneInlineData(items []map[string]any) []map[string]any {
	if len(items) == 0 {
		return nil
	}
	result := make([]map[string]any, 0, len(items))
	for _, item := range items {
		result = append(result, cloneInlineMap(item))
	}
	return result
}

func cloneInlineMap(src map[string]any) map[string]any {
	if src == nil {
		return nil
	}
	dst := make(map[string]any, len(src))
	for k, v := range src {
		dst[k] = cloneInlineValue(v)
	}
	return dst
}

func cloneInlineValue(value any) any {
	switch v := value.(type) {
	case map[string]any:
		return cloneInlineMap(v)
	case []any:
		res := make([]any, len(v))
		for i, item := range v {
			res[i] = cloneInlineValue(item)
		}
		return res
	default:
		return v
	}
}

func normalizeIncludeDirectives(entries []rawIncludeDirective) []IncludeDefinition {
	if len(entries) == 0 {
		return nil
	}

	seen := make(map[string]struct{}, len(entries))
	result := make([]IncludeDefinition, 0, len(entries))
	for _, entry := range entries {
		if entry.Path == "" {
			continue
		}
		key := entry.Path + "\x00" + entry.Selector
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		result = append(result, IncludeDefinition(entry))
	}

	return result
}

func isValidIdentifier(value string) bool {
	if value == "" {
		return false
	}
	for _, r := range value {
		if r >= 'a' && r <= 'z' {
			continue
		}
		if r >= 'A' && r <= 'Z' {
			continue
		}
		if r >= '0' && r <= '9' {
			continue
		}
		if r == '-' || r == '_' {
			continue
		}
		return false
	}
	return true
}

func isValidTypeName(value string) bool {
	if !isValidIdentifier(value) {
		return false
	}
	first := rune(value[0])
	return first >= 'A' && first <= 'Z'
}
