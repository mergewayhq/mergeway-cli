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
		return nil, errors.New("config: version is required")
	}

	result := &Config{
		Version: agg.Version,
		Types:   make(map[string]*TypeDefinition),
	}

	for name, rawType := range agg.Types {
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
	def := rawType.Definition

	if def.ID.Field == "" {
		return nil, fmt.Errorf("config: type %q missing id.field in %s", rawType.Name, rawType.Source)
	}

	if !isValidIdentifier(def.ID.Field) {
		return nil, fmt.Errorf("config: type %q has invalid identifier field %q", rawType.Name, def.ID.Field)
	}

	if len(def.FilePatterns) == 0 {
		return nil, fmt.Errorf("config: type %q must declare at least one file_patterns entry", rawType.Name)
	}

	fields := make(map[string]*FieldDefinition, len(def.Fields))

	for fieldName, rawField := range def.Fields {
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

	return &TypeDefinition{
		Name:   rawType.Name,
		Source: rawType.Source,
		Identifier: IdentifierDefinition{
			Field:     def.ID.Field,
			Generated: def.ID.Generated,
			Pattern:   def.ID.Pattern,
		},
		FilePatterns: deduplicateStrings(def.FilePatterns),
		Fields:       fields,
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

func deduplicateStrings(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	result := make([]string, 0, len(values))
	for _, v := range values {
		if v == "" {
			continue
		}
		if _, ok := seen[v]; ok {
			continue
		}
		seen[v] = struct{}{}
		result = append(result, v)
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
