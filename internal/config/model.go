package config

import (
	"fmt"
	"sort"
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

// FieldDefinition holds schema information for a field.
type FieldDefinition struct {
	Name          string
	Type          string
	Required      bool
	Repeated      bool
	Format        string
	Enum          []string
	Default       any
	Properties    map[string]*FieldDefinition
	Unique        bool
	Pattern       string
	Description   string
	PropertyOrder []string
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
