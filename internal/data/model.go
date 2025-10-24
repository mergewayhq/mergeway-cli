package data

import (
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"

	"github.com/mergewayhq/mergeway-cli/internal/config"
)

// Object represents a single database object loaded from disk.
type Object struct {
	Type   string
	ID     string
	Fields map[string]any
	File   string
}

// Store coordinates reading and writing objects on disk.
type Store struct {
	root   string
	config *config.Config
}

// NewStore constructs a data store rooted at the given directory.
func NewStore(root string, cfg *config.Config) (*Store, error) {
	if root == "" {
		return nil, errors.New("data: root is required")
	}
	if cfg == nil {
		return nil, errors.New("data: config is required")
	}

	absRoot, err := filepath.Abs(root)
	if err != nil {
		return nil, fmt.Errorf("data: resolve root: %w", err)
	}

	return &Store{root: absRoot, config: cfg}, nil
}

// objectLocation captures where an object lives.
type objectLocation struct {
	FilePath string
	Format   fileFormat
	Multi    bool
	Index    int
	Object   map[string]any
	File     *fileContent
	TypeName string
	IDField  string
	Inline   bool
	ReadOnly bool
}

func (loc *objectLocation) cloneObject() *Object {
	if loc == nil {
		return nil
	}
	id, _ := getString(loc.Object, loc.IDField)
	return &Object{Type: loc.TypeName, ID: id, Fields: cloneMap(loc.Object), File: loc.FilePath}
}

func (obj *Object) clone() *Object {
	if obj == nil {
		return nil
	}
	return &Object{Type: obj.Type, ID: obj.ID, Fields: cloneMap(obj.Fields), File: obj.File}
}

// fileFormat represents the serialization format of a file.
type fileFormat int

const (
	formatYAML fileFormat = iota
	formatJSON
)

// fileContent captures parsed file state.
type fileContent struct {
	Path     string
	TypeName string
	Format   fileFormat
	Multi    bool
	Single   map[string]any
	Items    []map[string]any
	Selector string
	ReadOnly bool
}

type createTarget struct {
	Path   string
	Format fileFormat
	Multi  bool
}

func (s *Store) requireType(typeName string) (*config.TypeDefinition, error) {
	typeDef, ok := s.config.Types[typeName]
	if !ok {
		return nil, fmt.Errorf("data: unknown type %q", typeName)
	}
	return typeDef, nil
}

// sentinel used when files are missing but optional.
var errFileNotFound = fs.ErrNotExist
