package config

import (
	"fmt"
	"sort"
)

// Config captures the normalized database configuration.
type Config struct {
	Version int
	Types   map[string]*TypeDefinition
}

// TypeDefinition describes a single object type.
type TypeDefinition struct {
	Name       string
	Source     string
	Identifier IdentifierDefinition
	Include    []string
	Fields     map[string]*FieldDefinition
}

// IdentifierDefinition specifies the identifier field metadata.
type IdentifierDefinition struct {
	Field     string
	Generated bool
	Pattern   string
}

// FieldDefinition holds schema information for a field.
type FieldDefinition struct {
	Name       string
	Type       string
	Required   bool
	Repeated   bool
	Format     string
	Enum       []string
	Default    any
	Properties map[string]*FieldDefinition
	Unique     bool
	Computed   bool
	Pattern    string
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
