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
	Extends     string
	Ancestors   []string
	Descendants []string
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
	Source         *FieldSourceDefinition `yaml:"source,omitempty" json:"source,omitempty"`
}

// FieldSourceDefinition describes a synthetic field value derived at read time.
type FieldSourceDefinition struct {
	Path           bool `yaml:"path,omitempty" json:"path,omitempty"`
	PathSegment    *int `yaml:"path_segment,omitempty" json:"path_segment,omitempty"`
	PathSegmentRev *int `yaml:"path_segment_rev,omitempty" json:"path_segment_rev,omitempty"`
}

// IsPathDerived returns true when the field value is derived from the backing file path.
func (d *FieldSourceDefinition) IsPathDerived() bool {
	return d != nil && (d.Path || d.PathSegment != nil || d.PathSegmentRev != nil)
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

// IsA reports whether child is the same as or inherits from parent.
func (c *Config) IsA(child, parent string) bool {
	if c == nil || child == "" || parent == "" {
		return false
	}
	if c.Types == nil {
		return false
	}
	if _, ok := c.Types[child]; !ok {
		return false
	}
	if _, ok := c.Types[parent]; !ok {
		return false
	}
	if child == parent {
		return true
	}

	seen := map[string]struct{}{child: {}}
	current := c.Types[child]
	for current != nil && current.Extends != "" {
		next := current.Extends
		if next == parent {
			return true
		}
		if _, ok := seen[next]; ok {
			return false
		}
		seen[next] = struct{}{}
		current = c.Types[next]
	}

	return false
}

// AssignableTypes returns the named type followed by any descendants.
func (c *Config) AssignableTypes(typeName string) []string {
	if c == nil || typeName == "" || c.Types[typeName] == nil {
		return nil
	}

	assignable := []string{typeName}
	assignable = append(assignable, c.DescendantTypes(typeName)...)
	return assignable
}

// DescendantTypes returns the names of all descendants for the given type.
func (c *Config) DescendantTypes(typeName string) []string {
	if c == nil || typeName == "" || c.Types[typeName] == nil {
		return nil
	}

	descendants := make([]string, 0)
	for candidate := range c.Types {
		if candidate == typeName {
			continue
		}
		if c.IsA(candidate, typeName) {
			descendants = append(descendants, candidate)
		}
	}
	sort.Strings(descendants)
	return descendants
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
