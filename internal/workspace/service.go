package workspace

import (
	"errors"
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/mergewayhq/mergeway-cli/internal/config"
	"github.com/mergewayhq/mergeway-cli/internal/data"
	"github.com/mergewayhq/mergeway-cli/internal/fileutil"
	"github.com/mergewayhq/mergeway-cli/internal/validation"
)

// Workspace captures a loaded mergeway workspace and a reusable entity index.
type Workspace struct {
	Root          string
	ConfigPath    string
	Config        *config.Config
	ObjectsByType map[string][]*data.Object
	Index         *Index
}

// Index maps entity types and identifiers to the loaded objects that match them.
type Index struct {
	ByType map[string]map[string][]*data.Object
}

// ValidationReport combines semantic validation output with the best-effort
// loaded workspace/index for callers that need both in one step.
type ValidationReport struct {
	Root               string
	ConfigPath         string
	Config             *config.Config
	Result             *validation.Result
	Workspace          *Workspace
	WorkspaceLoadError error
}

// Load reads the workspace config, loads all configured entities, and builds
// a per-type/per-identifier index without invoking CLI code.
func Load(root, configPath string) (*Workspace, error) {
	return LoadWithOps(root, configPath, fileutil.OS)
}

// LoadWithOps reads the workspace config and builds an index using the
// provided file operations.
func LoadWithOps(root, configPath string, ops fileutil.Ops) (*Workspace, error) {
	resolvedRoot, resolvedConfig, err := resolvePaths(root, configPath)
	if err != nil {
		return nil, err
	}

	cfg, err := config.LoadWithOps(resolvedConfig, ops)
	if err != nil {
		return nil, err
	}

	return LoadWithConfigAndOps(resolvedRoot, resolvedConfig, cfg, ops)
}

// LoadWithConfig loads all configured entities and builds an entity index using
// a caller-provided config.
func LoadWithConfig(root, configPath string, cfg *config.Config) (*Workspace, error) {
	return LoadWithConfigAndOps(root, configPath, cfg, fileutil.OS)
}

// LoadWithConfigAndOps loads all configured entities and builds an entity index using
// a caller-provided config and file operations.
func LoadWithConfigAndOps(root, configPath string, cfg *config.Config, ops fileutil.Ops) (*Workspace, error) {
	if cfg == nil {
		return nil, errors.New("workspace: config is required")
	}

	resolvedRoot, resolvedConfig, err := resolvePaths(root, configPath)
	if err != nil {
		return nil, err
	}

	store, err := data.NewStoreWithOps(resolvedRoot, cfg, ops)
	if err != nil {
		return nil, err
	}

	objectsByType := make(map[string][]*data.Object, len(cfg.Types))
	index := &Index{ByType: make(map[string]map[string][]*data.Object, len(cfg.Types))}

	for _, typeName := range sortedTypeNames(cfg.Types) {
		objects, err := store.LoadAll(typeName)
		if err != nil {
			return nil, fmt.Errorf("workspace: load %s: %w", typeName, err)
		}

		objectsByType[typeName] = objects
		if index.ByType[typeName] == nil {
			index.ByType[typeName] = make(map[string][]*data.Object)
		}
		for _, obj := range objects {
			index.ByType[typeName][obj.ID] = append(index.ByType[typeName][obj.ID], obj)
		}
	}

	return &Workspace{
		Root:          resolvedRoot,
		ConfigPath:    resolvedConfig,
		Config:        cfg,
		ObjectsByType: objectsByType,
		Index:         index,
	}, nil
}

// Objects returns the loaded objects for a type.
func (w *Workspace) Objects(typeName string) []*data.Object {
	if w == nil {
		return nil
	}
	return w.ObjectsByType[typeName]
}

// Find returns all indexed objects that match the given type and identifier.
func (w *Workspace) Find(typeName, id string) []*data.Object {
	if w == nil || w.Index == nil || w.Index.ByType[typeName] == nil {
		return nil
	}
	return w.Index.ByType[typeName][id]
}

// Validate loads config, runs the existing semantic validator, and attaches a
// best-effort loaded workspace/index for valid callers that need both outputs.
func Validate(root, configPath string, opts validation.Options) (*ValidationReport, error) {
	return ValidateWithOps(root, configPath, opts, fileutil.OS)
}

// ValidateWithOps loads config and validates using the provided file operations.
func ValidateWithOps(root, configPath string, opts validation.Options, ops fileutil.Ops) (*ValidationReport, error) {
	resolvedRoot, resolvedConfig, err := resolvePaths(root, configPath)
	if err != nil {
		return nil, err
	}

	cfg, err := config.LoadWithOps(resolvedConfig, ops)
	if err != nil {
		return nil, err
	}

	return ValidateWithConfigAndOps(resolvedRoot, resolvedConfig, cfg, opts, ops)
}

// ValidateWithConfig runs the existing semantic validator with a caller-provided
// config and attaches a best-effort loaded workspace/index.
func ValidateWithConfig(root, configPath string, cfg *config.Config, opts validation.Options) (*ValidationReport, error) {
	return ValidateWithConfigAndOps(root, configPath, cfg, opts, fileutil.OS)
}

// ValidateWithConfigAndOps runs semantic validation with caller-provided config
// and file operations.
func ValidateWithConfigAndOps(root, configPath string, cfg *config.Config, opts validation.Options, ops fileutil.Ops) (*ValidationReport, error) {
	if cfg == nil {
		return nil, errors.New("workspace: config is required")
	}

	resolvedRoot, resolvedConfig, err := resolvePaths(root, configPath)
	if err != nil {
		return nil, err
	}

	result, err := validation.ValidateWithOps(resolvedRoot, cfg, opts, ops)
	if err != nil {
		return nil, err
	}

	report := &ValidationReport{
		Root:       resolvedRoot,
		ConfigPath: resolvedConfig,
		Config:     cfg,
		Result:     result,
	}

	ws, loadErr := LoadWithConfigAndOps(resolvedRoot, resolvedConfig, cfg, ops)
	if loadErr != nil {
		report.WorkspaceLoadError = loadErr
		return report, nil
	}

	report.Workspace = ws
	return report, nil
}

func resolvePaths(root, configPath string) (string, string, error) {
	if root == "" {
		return "", "", errors.New("workspace: root is required")
	}

	resolvedRoot, err := filepath.Abs(root)
	if err != nil {
		return "", "", fmt.Errorf("workspace: resolve root: %w", err)
	}

	resolvedConfig := configPath
	if resolvedConfig == "" {
		detected, ok, err := DetectConfigPath(resolvedRoot)
		if err != nil {
			return "", "", err
		}
		if ok {
			resolvedConfig = detected
		} else {
			resolvedConfig = filepath.Join(resolvedRoot, "mergeway.yaml")
		}
	}
	if !filepath.IsAbs(resolvedConfig) {
		absConfig, absErr := filepath.Abs(resolvedConfig)
		if absErr == nil && pathWithinRoot(resolvedRoot, absConfig) {
			resolvedConfig = absConfig
		} else {
			resolvedConfig = filepath.Join(resolvedRoot, resolvedConfig)
		}
	}
	resolvedConfig, err = filepath.Abs(resolvedConfig)
	if err != nil {
		return "", "", fmt.Errorf("workspace: resolve config: %w", err)
	}

	return resolvedRoot, resolvedConfig, nil
}

func pathWithinRoot(root, path string) bool {
	rel, err := filepath.Rel(root, path)
	if err != nil {
		return false
	}
	return rel == "." || (rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)))
}

func sortedTypeNames(types map[string]*config.TypeDefinition) []string {
	names := make([]string, 0, len(types))
	for name := range types {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}
