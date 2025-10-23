package config

type rawConfigDocument struct {
	Version  *int                      `yaml:"version"`
	Includes []string                  `yaml:"includes"`
	Entities map[string]rawTypeWrapper `yaml:"entities"`
}

type rawTypeWrapper struct {
	Definition rawTypeDefinition `yaml:"definition"`
}

type rawTypeDefinition struct {
	ID           rawIdentifierDefinition       `yaml:"id"`
	FilePatterns []string                      `yaml:"file_patterns"`
	Fields       map[string]rawFieldDefinition `yaml:"fields"`
}

type rawIdentifierDefinition struct {
	Field     string `yaml:"field"`
	Generated bool   `yaml:"generated"`
	Pattern   string `yaml:"pattern"`
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
	Name       string
	Definition rawTypeDefinition
	Source     string
}

func newAggregateConfig() *aggregateConfig {
	return &aggregateConfig{
		Entities: make(map[string]rawTypeWithSource),
	}
}
