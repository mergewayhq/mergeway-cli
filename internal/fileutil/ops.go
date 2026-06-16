package fileutil

import (
	"os"
	"path/filepath"
)

// Ops describes the file operations needed by config/data/validation loaders.
type Ops struct {
	ReadFile func(path string) ([]byte, error)
	Stat     func(path string) (os.FileInfo, error)
	Glob     func(pattern string) ([]string, error)
}

// OS provides file operations backed by the local filesystem.
var OS = Ops{
	ReadFile: os.ReadFile,
	Stat:     os.Stat,
	Glob:     filepath.Glob,
}

// WithDefaults fills any nil operations with OS-backed defaults.
func (o Ops) WithDefaults() Ops {
	if o.ReadFile == nil {
		o.ReadFile = OS.ReadFile
	}
	if o.Stat == nil {
		o.Stat = OS.Stat
	}
	if o.Glob == nil {
		o.Glob = OS.Glob
	}
	return o
}
