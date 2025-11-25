package format

// Schema describes the canonical field ordering for an entity.
type Schema struct {
	fields []*SchemaField
	index  map[string]*SchemaField
}

// SchemaField represents a field entry and any nested schema for object properties.
type SchemaField struct {
	Name     string
	Repeated bool
	Nested   *Schema
}

// NewSchema constructs a schema with the provided ordered fields.
func NewSchema(fields []*SchemaField) *Schema {
	filtered := make([]*SchemaField, 0, len(fields))
	index := make(map[string]*SchemaField, len(fields))
	for _, field := range fields {
		if field == nil {
			continue
		}
		name := field.Name
		if name == "" {
			continue
		}
		filtered = append(filtered, field)
		index[name] = field
	}
	if len(filtered) == 0 {
		return nil
	}
	return &Schema{
		fields: filtered,
		index:  index,
	}
}

func (s *Schema) orderedFields() []*SchemaField {
	if s == nil {
		return nil
	}
	return s.fields
}
