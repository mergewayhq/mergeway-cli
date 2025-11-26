package config

import (
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

type rawConfigDocument struct {
	Mergeway *rawMergewaySection    `yaml:"mergeway"`
	Include  []string               `yaml:"include"`
	Entities map[string]rawTypeSpec `yaml:"entities"`
}

type rawMergewaySection struct {
	Version *int `yaml:"version"`
}

type rawTypeSpec struct {
	Identifier  rawIdentifierSpec     `yaml:"identifier"`
	Include     []rawIncludeDirective `yaml:"include"`
	Fields      rawFieldMap           `yaml:"fields"`
	JSONSchema  string                `yaml:"json_schema"`
	Data        []map[string]any      `yaml:"data"`
	Description string                `yaml:"description"`
}

type rawIncludeDirective struct {
	Path     string `yaml:"path"`
	Selector string `yaml:"selector"`
}

func (r *rawIncludeDirective) UnmarshalYAML(node *yaml.Node) error {
	if node == nil {
		return nil
	}

	switch node.Kind {
	case yaml.ScalarNode:
		var path string
		if err := node.Decode(&path); err != nil {
			return err
		}
		path = strings.TrimSpace(path)
		if path == "" {
			return fmt.Errorf("config: include path must be a non-empty string")
		}
		r.Path = path
		r.Selector = ""
		return nil
	case yaml.MappingNode:
		type alias rawIncludeDirective
		var tmp alias
		if err := node.Decode(&tmp); err != nil {
			return err
		}
		tmp.Path = strings.TrimSpace(tmp.Path)
		tmp.Selector = strings.TrimSpace(tmp.Selector)
		if tmp.Path == "" {
			return fmt.Errorf("config: include path must be a non-empty string")
		}
		*r = rawIncludeDirective(tmp)
		return nil
	case yaml.AliasNode:
		return node.Decode(r)
	default:
		return fmt.Errorf("config: include entry must be a string or mapping, got %s", node.ShortTag())
	}
}

type rawIdentifierSpec struct {
	Field     string `yaml:"field"`
	Generated bool   `yaml:"generated"`
	Pattern   string `yaml:"pattern"`
	set       bool
}

func (r *rawIdentifierSpec) UnmarshalYAML(node *yaml.Node) error {
	if node == nil {
		return nil
	}

	if node.Tag == "!!null" {
		return nil
	}

	switch node.Kind {
	case yaml.ScalarNode:
		var field string
		if err := node.Decode(&field); err != nil {
			return err
		}
		if field == "" {
			return fmt.Errorf("config: identifier must be a non-empty string")
		}
		r.Field = field
		r.Generated = false
		r.Pattern = ""
		r.set = true
		return nil
	case yaml.MappingNode:
		type alias rawIdentifierSpec
		var tmp alias
		if err := node.Decode(&tmp); err != nil {
			return err
		}
		*r = rawIdentifierSpec(tmp)
		r.set = true
		return nil
	default:
		return fmt.Errorf("config: identifier must be a string or mapping, got %s", node.ShortTag())
	}
}

type rawFieldDefinition struct {
	Type        string      `yaml:"type"`
	Required    bool        `yaml:"required"`
	Repeated    bool        `yaml:"repeated"`
	Format      string      `yaml:"format"`
	Enum        []string    `yaml:"enum"`
	Default     any         `yaml:"default"`
	Properties  rawFieldMap `yaml:"properties"`
	Unique      *bool       `yaml:"unique"`
	Pattern     string      `yaml:"pattern"`
	Description string      `yaml:"description"`
}

func (r *rawFieldDefinition) UnmarshalYAML(node *yaml.Node) error {
	if node == nil {
		return nil
	}

	switch node.Kind {
	case yaml.ScalarNode:
		var typeName string
		if err := node.Decode(&typeName); err != nil {
			return err
		}
		typeName = strings.TrimSpace(typeName)
		if typeName == "" {
			return fmt.Errorf("config: field type must be a non-empty string")
		}
		r.Type = typeName
		r.Required = false
		r.Repeated = false
		r.Format = ""
		r.Enum = nil
		r.Default = nil
		r.Properties = rawFieldMap{}
		r.Unique = nil
		r.Pattern = ""
		return nil
	case yaml.MappingNode:
		type alias rawFieldDefinition
		var tmp alias
		if err := node.Decode(&tmp); err != nil {
			return err
		}
		*r = rawFieldDefinition(tmp)
		return nil
	case yaml.AliasNode:
		return node.Decode(r)
	default:
		return fmt.Errorf("config: field definition must be a string or mapping, got %s", node.ShortTag())
	}
}

type rawFieldMap struct {
	Entries []rawFieldEntry
}

type rawFieldEntry struct {
	Name  string
	Value rawFieldDefinition
}

func (m *rawFieldMap) UnmarshalYAML(node *yaml.Node) error {
	if node == nil {
		return nil
	}

	switch node.Kind {
	case 0:
		return nil
	case yaml.MappingNode:
		if len(node.Content)%2 != 0 {
			return fmt.Errorf("config: expected even number of nodes in mapping, got %d", len(node.Content))
		}
		entries := make([]rawFieldEntry, 0, len(node.Content)/2)
		for i := 0; i < len(node.Content); i += 2 {
			keyNode := node.Content[i]
			valNode := node.Content[i+1]
			var name string
			if err := keyNode.Decode(&name); err != nil {
				return err
			}
			name = strings.TrimSpace(name)
			if name == "" {
				return fmt.Errorf("config: field name must be a non-empty string")
			}
			var value rawFieldDefinition
			if err := valNode.Decode(&value); err != nil {
				return err
			}
			entries = append(entries, rawFieldEntry{Name: name, Value: value})
		}
		m.Entries = entries
		return nil
	case yaml.AliasNode:
		if node.Alias == nil {
			return fmt.Errorf("config: alias node missing target")
		}
		return node.Alias.Decode(m)
	case yaml.ScalarNode:
		if node.ShortTag() == "!!null" {
			return nil
		}
		fallthrough
	default:
		return fmt.Errorf("config: fields/properties must be a mapping, got %s", node.ShortTag())
	}
}

type aggregateConfig struct {
	Version    int
	VersionSet bool
	Entities   map[string]rawTypeWithSource
}

type rawTypeWithSource struct {
	Name   string
	Spec   rawTypeSpec
	Source string
}

func newAggregateConfig() *aggregateConfig {
	return &aggregateConfig{
		Entities: make(map[string]rawTypeWithSource),
	}
}
