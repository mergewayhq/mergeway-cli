package diff

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	internalconfig "github.com/mergewayhq/mergeway-cli/internal/config"
	"gopkg.in/yaml.v3"
)

type diffSnapshotSchema struct {
	Snapshot SnapshotRef
	Types    map[string]*diffSnapshotType
}

type diffSnapshotType struct {
	Name                string
	IdentifierField     string
	IdentifierFieldType string
	Includes            []diffSnapshotInclude
}

type diffSnapshotInclude struct {
	Path     string
	Selector string
}

func (t *diffSnapshotType) identifierIsPath() bool {
	return t != nil && t.IdentifierField == internalconfig.PathIdentifierField
}

func (s *diffSnapshotSchema) includePatterns() []string {
	if s == nil {
		return nil
	}

	patterns := make(map[string]struct{})
	for _, typeDef := range s.Types {
		for _, include := range typeDef.Includes {
			if include.Path == "" {
				continue
			}
			patterns[include.Path] = struct{}{}
		}
	}

	return sortedKeys(patterns)
}

func loadSnapshotDiffSchema(root, configPath string, snapshot SnapshotRef) (*diffSnapshotSchema, error) {
	configRel, err := rootRelativePath(root, configPath)
	if err != nil {
		return nil, fmt.Errorf("diff: config path %s: %w", configPath, err)
	}

	collector, err := newDiffSnapshotSchemaCollector(root, snapshot)
	if err != nil {
		return nil, err
	}

	agg, err := collector.collect(filepath.Clean(configRel))
	if err != nil {
		return nil, err
	}

	return agg.normalize(snapshot)
}

type diffSnapshotSchemaCollector struct {
	root     string
	snapshot SnapshotRef
	reader   *snapshotReader
	cache    map[string]*diffSnapshotAggregate
	stack    map[string]bool
}

func newDiffSnapshotSchemaCollector(root string, snapshot SnapshotRef) (*diffSnapshotSchemaCollector, error) {
	reader, err := newSnapshotReader(root, snapshot)
	if err != nil {
		return nil, err
	}

	return &diffSnapshotSchemaCollector{
		root:     root,
		snapshot: snapshot,
		reader:   reader,
		cache:    make(map[string]*diffSnapshotAggregate),
		stack:    make(map[string]bool),
	}, nil
}

func (c *diffSnapshotSchemaCollector) collect(configRel string) (*diffSnapshotAggregate, error) {
	if cached, ok := c.cache[configRel]; ok {
		return cached.clone(), nil
	}
	if c.stack[configRel] {
		return nil, fmt.Errorf("diff: detected config include cycle at %s", configRel)
	}

	c.stack[configRel] = true
	defer delete(c.stack, configRel)

	content, exists, err := c.reader.Read(configRel)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, fmt.Errorf("diff: config file %s not found in snapshot %s", configRel, c.snapshot)
	}

	var doc diffSnapshotConfigDocument
	if err := yaml.Unmarshal(content, &doc); err != nil {
		return nil, fmt.Errorf("diff: parse config %s: %w", configRel, err)
	}

	agg := newDiffSnapshotAggregate()
	if err := agg.addDocument(configRel, &doc); err != nil {
		return nil, err
	}

	baseDir := filepath.Dir(configRel)
	for _, include := range doc.Include {
		includePattern, err := normalizeSnapshotPattern(baseDir, include)
		if err != nil {
			return nil, fmt.Errorf("diff: config include %q in %s: %w", include, configRel, err)
		}

		matches, err := matchSnapshotPattern(c.root, c.snapshot, includePattern)
		if err != nil {
			return nil, err
		}
		for _, match := range matches {
			childAgg, err := c.collect(match)
			if err != nil {
				return nil, err
			}
			if err := agg.merge(childAgg); err != nil {
				return nil, err
			}
		}
	}

	c.cache[configRel] = agg.clone()
	return agg, nil
}

type diffSnapshotAggregate struct {
	Version    int
	VersionSet bool
	Types      map[string]diffSnapshotRawTypeWithSource
}

func newDiffSnapshotAggregate() *diffSnapshotAggregate {
	return &diffSnapshotAggregate{
		Types: make(map[string]diffSnapshotRawTypeWithSource),
	}
}

func (a *diffSnapshotAggregate) clone() *diffSnapshotAggregate {
	if a == nil {
		return nil
	}

	out := &diffSnapshotAggregate{
		Version:    a.Version,
		VersionSet: a.VersionSet,
		Types:      make(map[string]diffSnapshotRawTypeWithSource, len(a.Types)),
	}
	for name, rawType := range a.Types {
		out.Types[name] = rawType
	}
	return out
}

func (a *diffSnapshotAggregate) merge(other *diffSnapshotAggregate) error {
	if other == nil {
		return nil
	}

	if other.VersionSet {
		if a.VersionSet {
			if a.Version != other.Version {
				return fmt.Errorf("diff: mergeway.version mismatch (got %d and %d)", a.Version, other.Version)
			}
		} else {
			a.Version = other.Version
			a.VersionSet = true
		}
	}

	for name, rawType := range other.Types {
		if existing, ok := a.Types[name]; ok {
			return fmt.Errorf("diff: entity %q defined in both %s and %s", name, existing.Source, rawType.Source)
		}
		a.Types[name] = rawType
	}

	return nil
}

func (a *diffSnapshotAggregate) addDocument(source string, doc *diffSnapshotConfigDocument) error {
	if doc == nil {
		return nil
	}
	if doc.Mergeway == nil {
		return fmt.Errorf("diff: mergeway block is required in %s", source)
	}
	if doc.Mergeway.Version == nil {
		return fmt.Errorf("diff: mergeway.version is required in %s", source)
	}

	version := *doc.Mergeway.Version
	if version != internalconfig.CurrentVersion {
		return fmt.Errorf("diff: mergeway.version must be %d in %s (got %d)", internalconfig.CurrentVersion, source, version)
	}

	if a.VersionSet {
		if a.Version != version {
			return fmt.Errorf("diff: mergeway.version mismatch (got %d and %d)", a.Version, version)
		}
	} else {
		a.Version = version
		a.VersionSet = true
	}

	for name, spec := range doc.Entities {
		if _, exists := a.Types[name]; exists {
			return fmt.Errorf("diff: entity %q defined in both %s and %s", name, a.Types[name].Source, source)
		}
		a.Types[name] = diffSnapshotRawTypeWithSource{
			Name:   name,
			Source: source,
			Spec:   spec,
		}
	}

	return nil
}

func (a *diffSnapshotAggregate) normalize(snapshot SnapshotRef) (*diffSnapshotSchema, error) {
	if a == nil {
		return nil, fmt.Errorf("diff: missing snapshot schema")
	}
	if !a.VersionSet {
		return nil, fmt.Errorf("diff: mergeway.version is required")
	}

	schema := &diffSnapshotSchema{
		Snapshot: snapshot,
		Types:    make(map[string]*diffSnapshotType, len(a.Types)),
	}

	for name, rawType := range a.Types {
		typeDef, err := normalizeDiffSnapshotType(rawType)
		if err != nil {
			return nil, err
		}
		schema.Types[name] = typeDef
	}

	return schema, nil
}

func normalizeDiffSnapshotType(rawType diffSnapshotRawTypeWithSource) (*diffSnapshotType, error) {
	if !rawType.Spec.Identifier.set || strings.TrimSpace(rawType.Spec.Identifier.Field) == "" {
		return nil, fmt.Errorf("diff: type %q missing identifier in %s", rawType.Name, rawType.Source)
	}

	includes := make([]diffSnapshotInclude, 0, len(rawType.Spec.Include))
	seen := make(map[string]struct{}, len(rawType.Spec.Include))
	for _, entry := range rawType.Spec.Include {
		path, err := normalizeSnapshotPattern(".", entry.Path)
		if err != nil {
			return nil, fmt.Errorf("diff: data include %q in %s: %w", entry.Path, rawType.Source, err)
		}
		key := path + "\x00" + entry.Selector
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		includes = append(includes, diffSnapshotInclude{
			Path:     path,
			Selector: entry.Selector,
		})
	}

	sort.Slice(includes, func(i, j int) bool {
		if includes[i].Path != includes[j].Path {
			return includes[i].Path < includes[j].Path
		}
		return includes[i].Selector < includes[j].Selector
	})

	identifierType := ""
	if field, ok := rawType.Spec.Fields[rawType.Spec.Identifier.Field]; ok {
		identifierType = field.Type
	}

	return &diffSnapshotType{
		Name:                rawType.Name,
		IdentifierField:     rawType.Spec.Identifier.Field,
		IdentifierFieldType: identifierType,
		Includes:            includes,
	}, nil
}

type diffSnapshotConfigDocument struct {
	Mergeway *diffSnapshotMergewaySection `yaml:"mergeway"`
	Include  []string                     `yaml:"include"`
	Entities map[string]diffSnapshotTypeSpec
}

type diffSnapshotMergewaySection struct {
	Version *int `yaml:"version"`
}

type diffSnapshotTypeSpec struct {
	Identifier diffSnapshotIdentifierSpec       `yaml:"identifier"`
	Include    []diffConfigInclude              `yaml:"include"`
	Fields     map[string]diffSnapshotFieldSpec `yaml:"fields"`
}

type diffSnapshotRawTypeWithSource struct {
	Name   string
	Source string
	Spec   diffSnapshotTypeSpec
}

type diffSnapshotIdentifierSpec struct {
	Field string `yaml:"field"`
	set   bool
}

func (r *diffSnapshotIdentifierSpec) UnmarshalYAML(node *yaml.Node) error {
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
		field = strings.TrimSpace(field)
		if field == "" {
			return fmt.Errorf("identifier must be a non-empty string")
		}
		r.Field = field
		r.set = true
		return nil
	case yaml.MappingNode, yaml.AliasNode:
		var tmp struct {
			Field string `yaml:"field"`
		}
		if err := node.Decode(&tmp); err != nil {
			return err
		}
		tmp.Field = strings.TrimSpace(tmp.Field)
		if tmp.Field == "" {
			return fmt.Errorf("identifier field must be a non-empty string")
		}
		r.Field = tmp.Field
		r.set = true
		return nil
	default:
		return fmt.Errorf("identifier must be a string or mapping, got %s", node.ShortTag())
	}
}

type diffSnapshotFieldSpec struct {
	Type string `yaml:"type"`
}

func (r *diffSnapshotFieldSpec) UnmarshalYAML(node *yaml.Node) error {
	if node == nil {
		return nil
	}

	switch node.Kind {
	case yaml.ScalarNode:
		var typeName string
		if err := node.Decode(&typeName); err != nil {
			return err
		}
		r.Type = strings.TrimSpace(typeName)
		if r.Type == "" {
			return fmt.Errorf("field type must be a non-empty string")
		}
		return nil
	case yaml.MappingNode, yaml.AliasNode:
		type alias diffSnapshotFieldSpec
		var tmp alias
		if err := node.Decode(&tmp); err != nil {
			return err
		}
		tmp.Type = strings.TrimSpace(tmp.Type)
		r.Type = tmp.Type
		return nil
	default:
		return fmt.Errorf("field definition must be a string or mapping, got %s", node.ShortTag())
	}
}
