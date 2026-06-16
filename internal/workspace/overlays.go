package workspace

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"

	"github.com/mergewayhq/mergeway-cli/internal/config"
	"github.com/mergewayhq/mergeway-cli/internal/fileutil"
	"github.com/mergewayhq/mergeway-cli/internal/validation"
)

// ReloadDelay is the debounce period used before recomputing overlay-backed roots.
const ReloadDelay = 75 * time.Millisecond

// OpenDocument is the in-memory representation of an open buffer.
type OpenDocument struct {
	URI        string
	Path       string
	LanguageID string
	Version    int32
	Text       string
}

// RootRuntime tracks the current overlay-backed view of one detected root.
type RootRuntime struct {
	Index      *RootIndex
	Workspace  *Workspace
	LoadErr    error
	Validation *ValidationReport
}

// Runtime manages open documents and overlay-backed root reloads.
type Runtime struct {
	mu        sync.Mutex
	base      *RootSet
	roots     map[string]*RootRuntime
	documents map[string]*OpenDocument
	timer     *time.Timer
	onReload  func()
}

// Snapshot captures a read-only copy of the runtime state used by callers that
// need a consistent view across roots and open documents.
type Snapshot struct {
	Roots     map[string]*RootRuntime
	Documents map[string]*OpenDocument
}

// NewRuntime constructs a runtime around a detected root set.
func NewRuntime(set *RootSet) *Runtime {
	rt := &Runtime{
		base:      set,
		roots:     make(map[string]*RootRuntime),
		documents: make(map[string]*OpenDocument),
	}
	if set != nil {
		for _, root := range set.Roots {
			rt.roots[root.Root] = &RootRuntime{
				Index:     root,
				Workspace: root.Workspace,
			}
		}
	}
	return rt
}

// DidOpen stores full-document content and schedules a debounced recompute.
func (r *Runtime) DidOpen(doc *OpenDocument) error {
	if doc == nil {
		return errors.New("workspace: document is required")
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.documents[doc.Path] = cloneDocument(doc)
	r.scheduleReloadLocked()
	return nil
}

// DidChange replaces the full document text and schedules a debounced recompute.
func (r *Runtime) DidChange(path string, version int32, text string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	doc, ok := r.documents[path]
	if !ok {
		return errors.New("workspace: document is not open")
	}
	doc.Version = version
	doc.Text = text
	r.scheduleReloadLocked()
	return nil
}

// DidClose removes the in-memory buffer and schedules a debounced recompute.
func (r *Runtime) DidClose(path string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.documents, path)
	r.scheduleReloadLocked()
}

// FlushReload forces any pending debounced recompute to run immediately.
func (r *Runtime) FlushReload() error {
	r.mu.Lock()
	if r.timer != nil {
		r.timer.Stop()
		r.timer = nil
	}
	r.mu.Unlock()
	return r.reload()
}

// Document returns the currently open in-memory buffer, if any.
func (r *Runtime) Document(path string) *OpenDocument {
	r.mu.Lock()
	defer r.mu.Unlock()
	return cloneDocument(r.documents[path])
}

// RootByPath returns the current root runtime for the given file.
func (r *Runtime) RootByPath(path string) *RootRuntime {
	resolved, ok := normalizeOwnedPath(path)
	if !ok {
		return nil
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	for _, root := range r.roots {
		if root.Index != nil && root.Index.OwnsPath(resolved) {
			return cloneRuntime(root)
		}
	}
	return nil
}

// Snapshot returns a stable copy of the current runtime state.
func (r *Runtime) Snapshot() *Snapshot {
	r.mu.Lock()
	defer r.mu.Unlock()
	return &Snapshot{
		Roots:     cloneRuntimes(r.roots),
		Documents: cloneDocuments(r.documents),
	}
}

// SetReloadHook registers a callback that runs after each successful reload.
func (r *Runtime) SetReloadHook(hook func()) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.onReload = hook
}

func (r *Runtime) scheduleReloadLocked() {
	if r.timer != nil {
		r.timer.Stop()
	}
	r.timer = time.AfterFunc(ReloadDelay, func() {
		_ = r.reload()
	})
}

func (r *Runtime) reload() error {
	r.mu.Lock()
	docs := cloneDocuments(r.documents)
	base := r.base
	r.mu.Unlock()

	ops := overlayOps{base: fileutil.OS, overlays: docs}.fileOps()

	next := make(map[string]*RootRuntime)
	if base != nil {
		for _, root := range base.Roots {
			cfg, err := loadRootConfig(root.Root, root.ConfigPath, ops)
			if err != nil {
				next[root.Root] = &RootRuntime{Index: root, LoadErr: err}
				continue
			}

			ws, loadErr := LoadWithConfigAndOps(root.Root, root.ConfigPath, cfg, ops)
			report, validateErr := ValidateWithConfigAndOps(root.Root, root.ConfigPath, cfg, validationOptions(), ops)
			if validateErr != nil && loadErr == nil {
				loadErr = validateErr
			}

			next[root.Root] = &RootRuntime{
				Index:      root,
				Workspace:  ws,
				LoadErr:    loadErr,
				Validation: report,
			}
		}
	}

	r.mu.Lock()
	hook := r.onReload
	for root, state := range next {
		r.roots[root] = state
	}
	r.mu.Unlock()
	if hook != nil {
		hook()
	}
	return nil
}

func loadRootConfig(root, configPath string, ops fileutil.Ops) (*config.Config, error) {
	return config.LoadWithOps(configPath, ops)
}

func validationOptions() validation.Options {
	return validation.Options{}
}

type overlayOps struct {
	base     fileutil.Ops
	overlays map[string]*OpenDocument
}

func (o overlayOps) fileOps() fileutil.Ops {
	base := o.base.WithDefaults()
	return fileutil.Ops{
		ReadFile: func(path string) ([]byte, error) {
			resolved, ok := normalizeOwnedPath(path)
			if ok {
				if doc := o.overlays[resolved]; doc != nil {
					return []byte(doc.Text), nil
				}
			}
			return base.ReadFile(path)
		},
		Stat: func(path string) (os.FileInfo, error) {
			resolved, ok := normalizeOwnedPath(path)
			if ok {
				if doc := o.overlays[resolved]; doc != nil {
					return overlayFileInfo{name: filepath.Base(doc.Path), size: int64(len(doc.Text))}, nil
				}
			}
			return base.Stat(path)
		},
		Glob: func(pattern string) ([]string, error) {
			matches, err := base.Glob(pattern)
			if err != nil {
				return nil, err
			}
			seen := make(map[string]struct{}, len(matches))
			for _, match := range matches {
				seen[filepath.Clean(match)] = struct{}{}
			}
			for path := range o.overlays {
				ok, matchErr := filepath.Match(pattern, path)
				if matchErr != nil {
					return nil, matchErr
				}
				if ok {
					seen[path] = struct{}{}
				}
			}
			result := make([]string, 0, len(seen))
			for path := range seen {
				result = append(result, path)
			}
			sort.Strings(result)
			return result, nil
		},
	}
}

type overlayFileInfo struct {
	name string
	size int64
}

func (o overlayFileInfo) Name() string       { return o.name }
func (o overlayFileInfo) Size() int64        { return o.size }
func (o overlayFileInfo) Mode() fs.FileMode  { return 0o644 }
func (o overlayFileInfo) ModTime() time.Time { return time.Unix(0, 0) }
func (o overlayFileInfo) IsDir() bool        { return false }
func (o overlayFileInfo) Sys() interface{}   { return nil }

func cloneDocument(doc *OpenDocument) *OpenDocument {
	if doc == nil {
		return nil
	}
	cloned := *doc
	return &cloned
}

func cloneDocuments(docs map[string]*OpenDocument) map[string]*OpenDocument {
	cloned := make(map[string]*OpenDocument, len(docs))
	for path, doc := range docs {
		cloned[path] = cloneDocument(doc)
	}
	return cloned
}

func cloneRuntime(runtime *RootRuntime) *RootRuntime {
	if runtime == nil {
		return nil
	}
	cloned := *runtime
	return &cloned
}

func cloneRuntimes(roots map[string]*RootRuntime) map[string]*RootRuntime {
	cloned := make(map[string]*RootRuntime, len(roots))
	for root, runtime := range roots {
		cloned[root] = cloneRuntime(runtime)
	}
	return cloned
}
