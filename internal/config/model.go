package config

import (
	"fmt"
	"sort"
	"strings"
)

// Config captures the normalized database configuration.
type Config struct {
	Version int
	Types   map[string]*TypeDefinition
	Write   WriteDefaults
}

// TypeDefinition describes a single object type.
type TypeDefinition struct {
	Name        string
	Source      string
	Description string
	JSONSchema  string
	Identifier  IdentifierDefinition
	Include     []IncludeDefinition
	Fields      map[string]*FieldDefinition
	FieldOrder  []string
	InlineData  []map[string]any
	Write       WriteDefinition
}

// WriteDefaults captures global defaults for write behaviour.
type WriteDefaults struct {
	Template string
	Format   WriteFormat
}

// WriteDefinition describes how to emit a single object to disk.
type WriteDefinition struct {
	Template string
	Format   WriteFormat
}

// WriteFormat enumerates supported serialization formats for writes.
type WriteFormat string

const (
	WriteFormatYAML WriteFormat = "yaml"
	WriteFormatJSON WriteFormat = "json"
)

// PathIdentifierField is the reserved identifier value that derives IDs from file paths.
const PathIdentifierField = "$path"

// DefaultWriteTemplate is the canonical template used for new object files.
const DefaultWriteTemplate = "{id}.yaml"

// IncludeDefinition links a type to data files and selectors.
type IncludeDefinition struct {
	Path     string
	Selector string
}

// IdentifierDefinition specifies the identifier field metadata.
type IdentifierDefinition struct {
	Field     string
	Generated bool
	Pattern   string
}

// IsPath returns true when identifiers are derived from workspace-relative file paths.
func (d IdentifierDefinition) IsPath() bool {
	return d.Field == PathIdentifierField
}

// FieldDefinition holds schema information for a field.
type FieldDefinition struct {
	Name           string
	Type           string
	ReferenceTypes []string
	Required       bool
	Repeated       bool
	Format         string
	Enum           []string
	Default        any
	Properties     map[string]*FieldDefinition
	Unique         bool
	Pattern        string
	Description    string
	PropertyOrder  []string
}

// IsReference returns true when the field stores an identifier for one or more entity types.
func (d *FieldDefinition) IsReference() bool {
	return d != nil && len(d.ReferenceTypes) > 0
}

// HasReferenceUnion returns true when the field can reference more than one entity type.
func (d *FieldDefinition) HasReferenceUnion() bool {
	return d != nil && len(d.ReferenceTypes) > 1
}

// ReferenceLabel returns the canonical textual representation of the field's reference targets.
func (d *FieldDefinition) ReferenceLabel() string {
	if d == nil {
		return ""
	}
	if len(d.ReferenceTypes) == 0 {
		return d.Type
	}
	return strings.Join(d.ReferenceTypes, " | ")
}

// String returns a string representation for debugging purposes.
func (c *Config) String() string {
	if c == nil {
		return "<nil>"
	}
	return fmt.Sprintf("Config{version:%d, types:%d}", c.Version, len(c.Types))
}

// Sources returns a sorted list of files that contributed entity definitions.
func (c *Config) Sources() []string {
	if c == nil {
		return nil
	}
	sources := make([]string, 0, len(c.Types))
	for _, t := range c.Types {
		sources = append(sources, t.Source)
	}
	sort.Strings(sources)
	return sources
}
