package config

import (
	"errors"
	"fmt"
	"path/filepath"
	"sort"
	"strings"
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
		Write: WriteDefaults{
			Template: DefaultWriteTemplate,
			Format:   WriteFormatYAML,
		},
	}

	graph, err := buildInheritanceGraph(agg.Entities)
	if err != nil {
		return nil, err
	}
	if err := validateInheritanceCycles(graph); err != nil {
		return nil, err
	}

	names := make([]string, 0, len(agg.Entities))
	for name, rawType := range agg.Entities {
		if !isValidTypeName(name) {
			return nil, fmt.Errorf("config: invalid type identifier %q in %s", name, rawType.Source)
		}
		names = append(names, name)
	}
	sort.Strings(names)

	resolved := make(map[string]*TypeDefinition, len(agg.Entities))
	var resolve func(string) (*TypeDefinition, error)
	resolve = func(name string) (*TypeDefinition, error) {
		if typeDef := resolved[name]; typeDef != nil {
			return typeDef, nil
		}

		rawType := agg.Entities[name]
		parentName := graph.parent[name]
		var parentDef *TypeDefinition
		if parentName != "" {
			var err error
			parentDef, err = resolve(parentName)
			if err != nil {
				return nil, err
			}
		}

		typeDef, err := normalizeTypeDefinition(rawType, parentDef, len(graph.children[name]) > 0)
		if err != nil {
			return nil, err
		}

		typeDef.Write = WriteDefinition{
			Template: DefaultWriteTemplate,
			Format:   WriteFormatYAML,
		}
		if parentDef != nil {
			typeDef.Ancestors = append(append([]string(nil), parentDef.Ancestors...), parentDef.Name)
		}

		result.Types[name] = typeDef
		resolved[name] = typeDef
		return typeDef, nil
	}

	for _, name := range names {
		if _, err := resolve(name); err != nil {
			return nil, err
		}
	}

	for _, name := range names {
		result.Types[name].Descendants = result.DescendantTypes(name)
	}

	if err := validateFieldReferences(result.Types); err != nil {
		return nil, err
	}

	return result, nil
}

type inheritanceGraph struct {
	parent   map[string]string
	children map[string][]string
}

func buildInheritanceGraph(types map[string]rawTypeWithSource) (*inheritanceGraph, error) {
	graph := &inheritanceGraph{
		parent:   make(map[string]string, len(types)),
		children: make(map[string][]string, len(types)),
	}

	for name := range types {
		graph.children[name] = nil
	}

	for name, rawType := range types {
		parent := strings.TrimSpace(rawType.Spec.Extends)
		if parent == "" {
			continue
		}
		if !isValidTypeName(parent) {
			return nil, fmt.Errorf("config: type %q has invalid parent type %q", name, parent)
		}
		if _, ok := types[parent]; !ok {
			return nil, fmt.Errorf("config: type %q extends unknown type %q", name, parent)
		}
		graph.parent[name] = parent
		graph.children[parent] = append(graph.children[parent], name)
	}

	for name := range graph.children {
		sort.Strings(graph.children[name])
	}

	return graph, nil
}

func validateInheritanceCycles(graph *inheritanceGraph) error {
	if graph == nil {
		return nil
	}

	const (
		statePending = iota
		stateVisiting
		stateDone
	)

	state := make(map[string]int, len(graph.children))
	stack := make([]string, 0, len(graph.children))

	var visit func(string) error
	visit = func(name string) error {
		switch state[name] {
		case stateDone:
			return nil
		case stateVisiting:
			cycleStart := 0
			for idx, item := range stack {
				if item == name {
					cycleStart = idx
					break
				}
			}
			cycle := append(append([]string(nil), stack[cycleStart:]...), name)
			return fmt.Errorf("config: cyclic inheritance detected: %s", strings.Join(cycle, " -> "))
		}

		state[name] = stateVisiting
		stack = append(stack, name)
		if parent := graph.parent[name]; parent != "" {
			if err := visit(parent); err != nil {
				return err
			}
		}
		stack = stack[:len(stack)-1]
		state[name] = stateDone
		return nil
	}

	names := make([]string, 0, len(graph.children))
	for name := range graph.children {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		if err := visit(name); err != nil {
			return err
		}
	}

	return nil
}

func normalizeTypeDefinition(rawType rawTypeWithSource, parentDef *TypeDefinition, hasChildren bool) (*TypeDefinition, error) {
	spec := rawType.Spec
	extends := strings.TrimSpace(spec.Extends)

	jsonSchemaPath := strings.TrimSpace(spec.JSONSchema)
	fieldCount := len(spec.Fields.Entries)

	switch {
	case jsonSchemaPath != "" && fieldCount > 0:
		return nil, fmt.Errorf("config: type %q cannot define both fields and json_schema", rawType.Name)
	case parentDef == nil && jsonSchemaPath == "" && fieldCount == 0:
		return nil, fmt.Errorf("config: type %q must define fields or json_schema", rawType.Name)
	}

	if jsonSchemaPath != "" && (extends != "" || hasChildren) {
		return nil, fmt.Errorf("config: type %q cannot use json_schema with inheritance", rawType.Name)
	}

	identifier, err := normalizeIdentifierDefinition(rawType, parentDef)
	if err != nil {
		return nil, err
	}

	if identifier.IsPath() && len(spec.Data) > 0 {
		return nil, fmt.Errorf("config: type %q cannot use identifier %q with inline data", rawType.Name, PathIdentifierField)
	}

	if len(spec.Include) == 0 && len(spec.Data) == 0 && !hasChildren {
		return nil, fmt.Errorf("config: type %q must declare at least one include or provide inline data", rawType.Name)
	}

	var fields map[string]*FieldDefinition
	var fieldOrder []string

	if jsonSchemaPath != "" {
		schemaPath := jsonSchemaPath
		if !filepath.IsAbs(schemaPath) {
			baseDir := filepath.Dir(rawType.Source)
			schemaPath = filepath.Join(baseDir, schemaPath)
		}
		var err error
		fields, fieldOrder, err = loadJSONSchemaFields(rawType.Name, schemaPath, jsonSchemaPath)
		if err != nil {
			return nil, err
		}
	} else {
		fields = make(map[string]*FieldDefinition, len(spec.Fields.Entries))
		fieldOrder = make([]string, 0, len(spec.Fields.Entries))
		for _, entry := range spec.Fields.Entries {
			fieldName := entry.Name
			rawField := entry.Value
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
			fieldOrder = append(fieldOrder, fieldName)
		}
	}

	for name, field := range fields {
		if field == nil || !field.Source.IsPathDerived() {
			continue
		}
		if len(spec.Data) > 0 {
			return nil, fmt.Errorf("config: type %q cannot use field %q source with inline data", rawType.Name, name)
		}
		if identifier.Field == name {
			return nil, fmt.Errorf("config: type %q identifier field %q cannot derive its value from field source", rawType.Name, name)
		}
	}

	fields, fieldOrder, err = mergeInheritedFields(rawType.Name, parentDef, fields, fieldOrder)
	if err != nil {
		return nil, err
	}

	inlineData := cloneInlineData(spec.Data)

	return &TypeDefinition{
		Name:        rawType.Name,
		Source:      rawType.Source,
		Description: spec.Description,
		Extends:     extends,
		JSONSchema:  jsonSchemaPath,
		Identifier:  identifier,
		Include:     normalizeIncludeDirectives(spec.Include),
		Fields:      fields,
		FieldOrder:  fieldOrder,
		InlineData:  inlineData,
	}, nil
}

func normalizeIdentifierDefinition(rawType rawTypeWithSource, parentDef *TypeDefinition) (IdentifierDefinition, error) {
	spec := rawType.Spec
	if parentDef == nil {
		if !spec.Identifier.set || spec.Identifier.Field == "" {
			return IdentifierDefinition{}, fmt.Errorf("config: type %q missing identifier in %s", rawType.Name, rawType.Source)
		}
		if !isValidIdentifierField(spec.Identifier.Field) {
			return IdentifierDefinition{}, fmt.Errorf("config: type %q has invalid identifier field %q", rawType.Name, spec.Identifier.Field)
		}
		return IdentifierDefinition{
			Field:     spec.Identifier.Field,
			Generated: spec.Identifier.Generated,
			Pattern:   spec.Identifier.Pattern,
		}, nil
	}

	if !spec.Identifier.set {
		return parentDef.Identifier, nil
	}
	if !isValidIdentifierField(spec.Identifier.Field) {
		return IdentifierDefinition{}, fmt.Errorf("config: type %q has invalid identifier field %q", rawType.Name, spec.Identifier.Field)
	}

	childIdentifier := IdentifierDefinition{
		Field:     spec.Identifier.Field,
		Generated: spec.Identifier.Generated,
		Pattern:   spec.Identifier.Pattern,
	}
	if childIdentifier != parentDef.Identifier {
		return IdentifierDefinition{}, fmt.Errorf("config: type %q cannot override inherited identifier from %q", rawType.Name, parentDef.Name)
	}

	return parentDef.Identifier, nil
}

func mergeInheritedFields(typeName string, parentDef *TypeDefinition, localFields map[string]*FieldDefinition, localOrder []string) (map[string]*FieldDefinition, []string, error) {
	if parentDef == nil {
		return localFields, localOrder, nil
	}

	mergedFields := cloneFieldMap(parentDef.Fields)
	if mergedFields == nil {
		mergedFields = make(map[string]*FieldDefinition)
	}
	mergedOrder := append([]string(nil), parentDef.FieldOrder...)

	for _, fieldName := range localOrder {
		if _, ok := mergedFields[fieldName]; ok {
			return nil, nil, fmt.Errorf("config: type %q cannot redefine inherited field %q", typeName, fieldName)
		}
		mergedFields[fieldName] = cloneFieldDefinition(localFields[fieldName])
		mergedOrder = append(mergedOrder, fieldName)
	}

	return mergedFields, mergedOrder, nil
}

func cloneFieldMap(fields map[string]*FieldDefinition) map[string]*FieldDefinition {
	if len(fields) == 0 {
		return nil
	}

	cloned := make(map[string]*FieldDefinition, len(fields))
	for name, field := range fields {
		cloned[name] = cloneFieldDefinition(field)
	}
	return cloned
}

func cloneFieldDefinition(field *FieldDefinition) *FieldDefinition {
	if field == nil {
		return nil
	}

	cloned := &FieldDefinition{
		Name:           field.Name,
		Type:           field.Type,
		ReferenceTypes: append([]string(nil), field.ReferenceTypes...),
		Required:       field.Required,
		Repeated:       field.Repeated,
		Format:         field.Format,
		Enum:           append([]string(nil), field.Enum...),
		Default:        cloneInlineValue(field.Default),
		Properties:     cloneFieldMap(field.Properties),
		Unique:         field.Unique,
		Pattern:        field.Pattern,
		Description:    field.Description,
		PropertyOrder:  append([]string(nil), field.PropertyOrder...),
	}
	if field.Source != nil {
		cloned.Source = &FieldSourceDefinition{
			Path: field.Source.Path,
		}
		if field.Source.PathSegment != nil {
			value := *field.Source.PathSegment
			cloned.Source.PathSegment = &value
		}
		if field.Source.PathSegmentRev != nil {
			value := *field.Source.PathSegmentRev
			cloned.Source.PathSegmentRev = &value
		}
	}

	return cloned
}

func normalizeFieldDefinition(name string, raw rawFieldDefinition, typeName string) (*FieldDefinition, error) {
	if raw.Type == "" {
		return nil, fmt.Errorf("config: field %s.%s missing type", typeName, name)
	}

	if raw.Repeated && raw.Unique != nil && *raw.Unique {
		return nil, fmt.Errorf("config: field %s.%s cannot declare unique=true when repeated", typeName, name)
	}

	if len(raw.Properties.Entries) > 0 && raw.Type != "object" {
		return nil, fmt.Errorf("config: field %s.%s defines properties but type is %q", typeName, name, raw.Type)
	}
	if len(raw.Fields.Entries) > 0 {
		return nil, fmt.Errorf("config: field %s.%s uses unsupported nested fields; use properties for object fields", typeName, name)
	}

	source, err := normalizeFieldSource(raw.Source, name, typeName)
	if err != nil {
		return nil, err
	}
	if source != nil {
		if raw.Repeated {
			return nil, fmt.Errorf("config: field %s.%s cannot be repeated when source is defined", typeName, name)
		}
		if raw.Type == "integer" || raw.Type == "number" || raw.Type == "boolean" || raw.Type == "object" {
			return nil, fmt.Errorf("config: field %s.%s source requires a string-like field type", typeName, name)
		}
		if raw.Default != nil {
			return nil, fmt.Errorf("config: field %s.%s cannot define both source and default", typeName, name)
		}
	}

	properties := make(map[string]*FieldDefinition)
	var propertyOrder []string
	if raw.Type == "object" {
		if len(raw.Properties.Entries) == 0 {
			return nil, fmt.Errorf("config: field %s.%s type object must define properties", typeName, name)
		}
		for _, entry := range raw.Properties.Entries {
			propName := entry.Name
			propField := entry.Value
			if !isValidIdentifier(propName) {
				return nil, fmt.Errorf("config: field %s.%s has invalid property identifier %q", typeName, name, propName)
			}

			child, err := normalizeFieldDefinition(propName, propField, fmt.Sprintf("%s.%s", typeName, name))
			if err != nil {
				return nil, err
			}

			properties[propName] = child
			propertyOrder = append(propertyOrder, propName)
		}
	}

	var unique bool
	if raw.Unique != nil {
		unique = *raw.Unique
	}

	return &FieldDefinition{
		Name:          name,
		Type:          raw.Type,
		Required:      raw.Required,
		Repeated:      raw.Repeated,
		Format:        raw.Format,
		Enum:          append([]string(nil), raw.Enum...),
		Default:       raw.Default,
		Properties:    properties,
		PropertyOrder: propertyOrder,
		Unique:        unique,
		Pattern:       raw.Pattern,
		Description:   raw.Description,
		Source:        source,
	}, nil
}

func normalizeFieldSource(raw rawFieldSourceDefinition, name, typeName string) (*FieldSourceDefinition, error) {
	selected := 0
	if raw.Path {
		selected++
	}
	if raw.PathSegment != nil {
		selected++
	}
	if raw.PathSegmentRev != nil {
		selected++
	}
	if selected == 0 {
		return nil, nil
	}
	if selected > 1 {
		return nil, fmt.Errorf("config: field %s.%s source must declare exactly one path selector", typeName, name)
	}
	if raw.PathSegment != nil && *raw.PathSegment < 0 {
		return nil, fmt.Errorf("config: field %s.%s source.path_segment must be >= 0", typeName, name)
	}
	if raw.PathSegmentRev != nil && *raw.PathSegmentRev < 0 {
		return nil, fmt.Errorf("config: field %s.%s source.path_segment_rev must be >= 0", typeName, name)
	}

	source := &FieldSourceDefinition{
		Path: raw.Path,
	}
	if raw.PathSegment != nil {
		value := *raw.PathSegment
		source.PathSegment = &value
	}
	if raw.PathSegmentRev != nil {
		value := *raw.PathSegmentRev
		source.PathSegmentRev = &value
	}
	return source, nil
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

	refTypes, err := parseReferenceTypes(field.Type)
	if err != nil {
		return fmt.Errorf("config: field %s.%s %w", typeName, field.Name, err)
	}
	field.ReferenceTypes = nil
	if len(refTypes) > 0 {
		for _, refType := range refTypes {
			if _, ok := types[refType]; !ok {
				return fmt.Errorf("config: field %s.%s references unknown type %q", typeName, field.Name, refType)
			}
		}
		field.ReferenceTypes = append(field.ReferenceTypes[:0], refTypes...)
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

func parseReferenceTypes(typeName string) ([]string, error) {
	typeName = strings.TrimSpace(typeName)
	if typeName == "" {
		return nil, fmt.Errorf("references invalid type name %q", typeName)
	}

	if _, ok := primitiveTypes[typeName]; ok {
		return nil, nil
	}

	if strings.Contains(typeName, "|") {
		parts := strings.Split(typeName, "|")
		if len(parts) < 2 {
			return nil, fmt.Errorf("uses invalid reference union %q", typeName)
		}

		refTypes := make([]string, 0, len(parts))
		seen := make(map[string]struct{}, len(parts))
		for _, rawPart := range parts {
			part := strings.TrimSpace(rawPart)
			if part == "" {
				return nil, fmt.Errorf("uses invalid reference union %q", typeName)
			}
			if _, ok := primitiveTypes[part]; ok {
				return nil, fmt.Errorf("uses invalid reference union %q: unions are only supported for references", typeName)
			}
			if !isValidTypeName(part) {
				return nil, fmt.Errorf("references invalid type name %q", part)
			}
			if _, ok := seen[part]; ok {
				return nil, fmt.Errorf("uses invalid reference union %q: duplicate member %q", typeName, part)
			}
			seen[part] = struct{}{}
			refTypes = append(refTypes, part)
		}
		return refTypes, nil
	}

	if !isValidTypeName(typeName) {
		return nil, fmt.Errorf("references invalid type name %q", typeName)
	}

	return []string{typeName}, nil
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

func isValidIdentifierField(value string) bool {
	return value == PathIdentifierField || isValidIdentifier(value)
}

func isValidTypeName(value string) bool {
	if !isValidIdentifier(value) {
		return false
	}
	first := rune(value[0])
	return first >= 'A' && first <= 'Z'
}
