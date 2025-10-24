package config

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

type rawConfigDocument struct {
	Version  *int                   `yaml:"version"`
	Include  []string               `yaml:"include"`
	Entities map[string]rawTypeSpec `yaml:"entities"`
}

type rawTypeSpec struct {
	Identifier   rawIdentifierSpec             `yaml:"identifier"`
	FilePatterns []string                      `yaml:"file_patterns"`
	Fields       map[string]rawFieldDefinition `yaml:"fields"`
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
	Type       string                        `yaml:"type"`
	Required   bool                          `yaml:"required"`
	Repeated   bool                          `yaml:"repeated"`
	Format     string                        `yaml:"format"`
	Enum       []string                      `yaml:"enum"`
	Default    any                           `yaml:"default"`
	Properties map[string]rawFieldDefinition `yaml:"properties"`
	Unique     *bool                         `yaml:"unique"`
	Computed   bool                          `yaml:"computed"`
	Pattern    string                        `yaml:"pattern"`
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
