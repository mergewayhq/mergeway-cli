package mcp

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/mergewayhq/mergeway-cli/internal/config"
	"github.com/mergewayhq/mergeway-cli/internal/data"
	"github.com/mergewayhq/mergeway-cli/internal/workspace"
)

var (
	// ErrUnknownEntity reports that the requested entity does not exist in the repository config.
	ErrUnknownEntity = errors.New("mcp: unknown entity")
	// ErrEntityNotAllowed reports that the requested entity is blocked by the configured allow-list.
	ErrEntityNotAllowed = errors.New("mcp: entity not allowed")
)

// FileEntry describes one configured backing file for an entity.
type FileEntry struct {
	Type string `json:"type" yaml:"type"`
	File string `json:"file" yaml:"file"`
}

// Service exposes read-only Mergeway repository queries for MCP handlers.
type Service struct {
	root     string
	allowed  map[string]struct{}
	entities []string
}

// NewService constructs a read-only query service rooted at the given repository.
func NewService(root string, allowedEntities []string) (*Service, error) {
	resolvedRoot, err := filepath.Abs(root)
	if err != nil {
		return nil, fmt.Errorf("mcp: resolve root: %w", err)
	}

	info, err := os.Stat(resolvedRoot)
	if err != nil {
		return nil, fmt.Errorf("mcp: stat root %s: %w", resolvedRoot, err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("mcp: root %s is not a directory", resolvedRoot)
	}

	cfgPath, found, err := workspace.DetectConfigPath(resolvedRoot)
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, fmt.Errorf("mcp: no mergeway config found under root %s", resolvedRoot)
	}

	cfg, err := config.Load(cfgPath)
	if err != nil {
		return nil, err
	}

	normalized, allowed, err := normalizeAllowedEntities(cfg, allowedEntities)
	if err != nil {
		return nil, err
	}

	return &Service{
		root:     resolvedRoot,
		allowed:  allowed,
		entities: normalized,
	}, nil
}

// Root reports the configured repository root.
func (s *Service) Root() string {
	if s == nil {
		return ""
	}
	return s.root
}

// AllowedEntities reports the normalized exact-name entity allow-list.
func (s *Service) AllowedEntities() []string {
	if s == nil || len(s.entities) == 0 {
		return nil
	}
	return append([]string(nil), s.entities...)
}

// EntityList returns the visible entity names after allow-list filtering.
func (s *Service) EntityList() ([]string, error) {
	state, err := s.loadState()
	if err != nil {
		return nil, err
	}
	return s.visibleEntityNames(state.Config), nil
}

// EntityShow returns the schema/config details for one visible entity.
func (s *Service) EntityShow(typeName string) (*config.TypeDefinition, error) {
	state, err := s.loadState()
	if err != nil {
		return nil, err
	}
	typeDef, err := s.requireVisibleEntity(state.Config, typeName)
	if err != nil {
		return nil, err
	}
	return cloneTypeDefinition(typeDef), nil
}

// ObjectList returns objects declared exactly as typeName after allow-list enforcement.
func (s *Service) ObjectList(typeName string) ([]*data.Object, error) {
	state, err := s.loadState()
	if err != nil {
		return nil, err
	}
	if _, err := s.requireVisibleEntity(state.Config, typeName); err != nil {
		return nil, err
	}

	objects, err := state.Store.LoadExactAll(typeName)
	if err != nil {
		return nil, err
	}
	sort.Slice(objects, func(i, j int) bool {
		return objects[i].ID < objects[j].ID
	})

	result := make([]*data.Object, len(objects))
	for i, obj := range objects {
		result[i] = cloneObject(obj)
	}
	return result, nil
}

// ObjectGet returns one object declared exactly as typeName.
func (s *Service) ObjectGet(typeName, id string) (*data.Object, error) {
	state, err := s.loadState()
	if err != nil {
		return nil, err
	}
	if _, err := s.requireVisibleEntity(state.Config, typeName); err != nil {
		return nil, err
	}

	obj, err := state.Store.GetExact(typeName, id)
	if err != nil {
		return nil, err
	}
	return cloneObject(obj), nil
}

// RepositoryExport returns a structured snapshot of the requested visible entities.
// When include is empty, all visible entities are exported.
func (s *Service) RepositoryExport(include []string) (map[string][]map[string]any, error) {
	state, err := s.loadState()
	if err != nil {
		return nil, err
	}

	entities, err := s.resolveRequestedEntities(state.Config, include)
	if err != nil {
		return nil, err
	}

	result := make(map[string][]map[string]any, len(entities))
	for _, typeName := range entities {
		objects, err := state.Store.LoadExactAll(typeName)
		if err != nil {
			return nil, err
		}
		sort.Slice(objects, func(i, j int) bool {
			return objects[i].ID < objects[j].ID
		})

		records := make([]map[string]any, len(objects))
		for i, obj := range objects {
			records[i] = cloneMap(obj.Fields)
		}
		result[typeName] = records
	}

	return result, nil
}

// FilesList returns configured backing files for visible entities.
// When typeName is empty, it includes all visible entities.
func (s *Service) FilesList(typeName string) ([]FileEntry, error) {
	state, err := s.loadState()
	if err != nil {
		return nil, err
	}

	var requested []string
	if strings.TrimSpace(typeName) != "" {
		requested = []string{typeName}
	}

	entities, err := s.resolveRequestedEntities(state.Config, requested)
	if err != nil {
		return nil, err
	}

	entries, err := collectFileEntries(state.Workspace.Root, state.Config, entities)
	if err != nil {
		return nil, err
	}

	sort.Slice(entries, func(i, j int) bool {
		if entries[i].Type != entries[j].Type {
			return entries[i].Type < entries[j].Type
		}
		return entries[i].File < entries[j].File
	})

	return entries, nil
}

type state struct {
	Workspace *workspace.Workspace
	Config    *config.Config
	Store     *data.Store
}

func (s *Service) loadState() (*state, error) {
	if s == nil {
		return nil, errors.New("mcp: service is required")
	}

	ws, err := workspace.Load(s.root, "")
	if err != nil {
		return nil, err
	}

	store, err := data.NewStore(ws.Root, ws.Config)
	if err != nil {
		return nil, err
	}

	return &state{
		Workspace: ws,
		Config:    ws.Config,
		Store:     store,
	}, nil
}

func (s *Service) resolveRequestedEntities(cfg *config.Config, include []string) ([]string, error) {
	if len(include) == 0 {
		return s.visibleEntityNames(cfg), nil
	}

	seen := make(map[string]struct{}, len(include))
	entities := make([]string, 0, len(include))
	for _, raw := range include {
		name := strings.TrimSpace(raw)
		if name == "" {
			return nil, fmt.Errorf("%w: empty entity name", ErrUnknownEntity)
		}
		if _, dup := seen[name]; dup {
			continue
		}
		if _, err := s.requireVisibleEntity(cfg, name); err != nil {
			return nil, err
		}
		seen[name] = struct{}{}
		entities = append(entities, name)
	}

	sort.Strings(entities)
	return entities, nil
}

func (s *Service) visibleEntityNames(cfg *config.Config) []string {
	if cfg == nil {
		return nil
	}

	names := make([]string, 0, len(cfg.Types))
	for name := range cfg.Types {
		if s.isEntityAllowed(name) {
			names = append(names, name)
		}
	}
	sort.Strings(names)
	return names
}

func (s *Service) requireVisibleEntity(cfg *config.Config, typeName string) (*config.TypeDefinition, error) {
	name := strings.TrimSpace(typeName)
	if cfg == nil || cfg.Types == nil || cfg.Types[name] == nil {
		return nil, fmt.Errorf("%w: %s", ErrUnknownEntity, name)
	}
	if !s.isEntityAllowed(name) {
		return nil, fmt.Errorf("%w: %s", ErrEntityNotAllowed, name)
	}
	return cfg.Types[name], nil
}

func (s *Service) isEntityAllowed(typeName string) bool {
	if s == nil || len(s.allowed) == 0 {
		return true
	}
	_, ok := s.allowed[typeName]
	return ok
}

func normalizeAllowedEntities(cfg *config.Config, allowed []string) ([]string, map[string]struct{}, error) {
	if len(allowed) == 0 {
		return nil, nil, nil
	}
	if cfg == nil || cfg.Types == nil {
		return nil, nil, errors.New("mcp: config is required")
	}

	seen := make(map[string]struct{}, len(allowed))
	normalized := make([]string, 0, len(allowed))
	for _, raw := range allowed {
		name := strings.TrimSpace(raw)
		if name == "" {
			return nil, nil, errors.New("mcp: allowed entity cannot be empty")
		}
		if _, ok := cfg.Types[name]; !ok {
			return nil, nil, fmt.Errorf("%w: %s", ErrUnknownEntity, name)
		}
		if _, dup := seen[name]; dup {
			continue
		}
		seen[name] = struct{}{}
		normalized = append(normalized, name)
	}
	sort.Strings(normalized)

	allowedSet := make(map[string]struct{}, len(normalized))
	for _, name := range normalized {
		allowedSet[name] = struct{}{}
	}

	return normalized, allowedSet, nil
}

func collectFileEntries(root string, cfg *config.Config, entityNames []string) ([]FileEntry, error) {
	if cfg == nil {
		return nil, nil
	}

	entries := make(map[string]FileEntry)
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return nil, fmt.Errorf("mcp: resolve root %s: %w", root, err)
	}

	for _, typeName := range entityNames {
		typeDef := cfg.Types[typeName]
		if typeDef == nil {
			continue
		}

		for _, include := range typeDef.Include {
			if include.Path == "" {
				continue
			}

			pattern := include.Path
			if !filepath.IsAbs(pattern) {
				pattern = filepath.Join(absRoot, filepath.Clean(pattern))
			}

			matches, err := filepath.Glob(pattern)
			if err != nil {
				return nil, fmt.Errorf("mcp: glob %s: %w", include.Path, err)
			}
			if len(matches) == 0 {
				continue
			}

			for _, match := range matches {
				info, err := os.Stat(match)
				if err != nil {
					if errors.Is(err, fs.ErrNotExist) {
						continue
					}
					return nil, fmt.Errorf("mcp: stat %s: %w", match, err)
				}
				if info.IsDir() {
					continue
				}

				absMatch := match
				if !filepath.IsAbs(absMatch) {
					absMatch, err = filepath.Abs(absMatch)
					if err != nil {
						return nil, fmt.Errorf("mcp: resolve match %s: %w", match, err)
					}
				}
				absMatch = filepath.Clean(absMatch)
				if !isSupportedDataPath(absMatch) {
					continue
				}

				entryPath := displayWorkspacePath(absRoot, absMatch)

				key := typeDef.Name + "\x00" + entryPath
				entries[key] = FileEntry{Type: typeDef.Name, File: entryPath}
			}
		}
	}

	result := make([]FileEntry, 0, len(entries))
	for _, entry := range entries {
		result = append(result, entry)
	}
	return result, nil
}

func cloneTypeDefinition(def *config.TypeDefinition) *config.TypeDefinition {
	if def == nil {
		return nil
	}

	cloned := *def
	cloned.Ancestors = append([]string(nil), def.Ancestors...)
	cloned.Descendants = append([]string(nil), def.Descendants...)
	cloned.Include = append([]config.IncludeDefinition(nil), def.Include...)
	cloned.FieldOrder = append([]string(nil), def.FieldOrder...)
	cloned.InlineData = cloneMapSlice(def.InlineData)
	cloned.Fields = cloneFieldDefinitions(def.Fields)
	return &cloned
}

func cloneFieldDefinitions(fields map[string]*config.FieldDefinition) map[string]*config.FieldDefinition {
	if fields == nil {
		return nil
	}
	cloned := make(map[string]*config.FieldDefinition, len(fields))
	for name, field := range fields {
		cloned[name] = cloneFieldDefinition(field)
	}
	return cloned
}

func cloneFieldDefinition(field *config.FieldDefinition) *config.FieldDefinition {
	if field == nil {
		return nil
	}

	cloned := *field
	cloned.ReferenceTypes = append([]string(nil), field.ReferenceTypes...)
	cloned.Enum = append([]string(nil), field.Enum...)
	cloned.PropertyOrder = append([]string(nil), field.PropertyOrder...)
	cloned.Default = cloneValue(field.Default)
	cloned.Properties = cloneFieldDefinitions(field.Properties)

	if field.Source != nil {
		source := *field.Source
		if field.Source.PathSegment != nil {
			value := *field.Source.PathSegment
			source.PathSegment = &value
		}
		if field.Source.PathSegmentRev != nil {
			value := *field.Source.PathSegmentRev
			source.PathSegmentRev = &value
		}
		cloned.Source = &source
	}

	return &cloned
}

func cloneObject(obj *data.Object) *data.Object {
	if obj == nil {
		return nil
	}
	return &data.Object{
		Type:     obj.Type,
		ID:       obj.ID,
		Fields:   cloneMap(obj.Fields),
		File:     obj.File,
		Inline:   obj.Inline,
		ReadOnly: obj.ReadOnly,
	}
}

func cloneMapSlice(values []map[string]any) []map[string]any {
	if values == nil {
		return nil
	}
	cloned := make([]map[string]any, len(values))
	for i, value := range values {
		cloned[i] = cloneMap(value)
	}
	return cloned
}

func cloneMap(value map[string]any) map[string]any {
	if value == nil {
		return nil
	}

	cloned := make(map[string]any, len(value))
	for key, item := range value {
		cloned[key] = cloneValue(item)
	}
	return cloned
}

func cloneValue(value any) any {
	switch v := value.(type) {
	case map[string]any:
		return cloneMap(v)
	case []any:
		cloned := make([]any, len(v))
		for i, item := range v {
			cloned[i] = cloneValue(item)
		}
		return cloned
	case []string:
		return append([]string(nil), v...)
	default:
		return v
	}
}

func isSupportedDataPath(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	return ext == ".yaml" || ext == ".yml" || ext == ".json"
}

func displayWorkspacePath(root, path string) string {
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return path
	}
	absPath, err := filepath.Abs(path)
	if err != nil {
		return path
	}
	rel, err := filepath.Rel(absRoot, absPath)
	if err != nil {
		return path
	}
	return filepath.ToSlash(rel)
}
