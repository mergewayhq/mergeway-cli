package data

import (
	internaldata "github.com/mergewayhq/mergeway-cli/internal/data"
	pkgconfig "github.com/mergewayhq/mergeway-cli/pkg/config"
)

// Store exposes the reusable data store implementation.
type Store = internaldata.Store

// Object re-exports the object type managed by the store.
type Object = internaldata.Object

// NewStore constructs a data store rooted at the given directory.
func NewStore(root string, cfg *pkgconfig.Config) (*Store, error) {
	return internaldata.NewStore(root, cfg)
}
