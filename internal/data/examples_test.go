package data

import (
	"os"
	"path/filepath"
	"sort"
	"testing"

	"github.com/mergewayhq/mergeway-cli/internal/config"
)

func TestStoreLoadExamples(t *testing.T) {
	for _, root := range exampleRoots(t) {
		root := root
		t.Run(filepath.Base(root), func(t *testing.T) {
			t.Helper()
			cfgPath := filepath.Join(root, "mergeway.yaml")
			cfg, err := config.Load(cfgPath)
			if err != nil {
				t.Fatalf("load config: %v", err)
			}

			store, err := NewStore(root, cfg)
			if err != nil {
				t.Fatalf("new store: %v", err)
			}

			for typeName := range cfg.Types {
				objects, err := store.LoadAll(typeName)
				if err != nil {
					t.Fatalf("load %s: %v", typeName, err)
				}
				if len(objects) == 0 {
					t.Fatalf("expected at least one %s object in %s", typeName, root)
				}
			}
		})
	}
}

func TestStoreLoadsOKFExampleBodyAndPathDerivedID(t *testing.T) {
	root, err := filepath.Abs(filepath.Join("..", "..", "examples", "okf"))
	if err != nil {
		t.Fatalf("abs example root: %v", err)
	}
	cfg, err := config.Load(filepath.Join(root, "mergeway.yaml"))
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	store, err := NewStore(root, cfg)
	if err != nil {
		t.Fatalf("new store: %v", err)
	}

	obj, err := store.Get("Knowledge", "customer-orders.md")
	if err != nil {
		t.Fatalf("get OKF concept: %v", err)
	}
	if got, want := obj.Fields["id"], "customer-orders.md"; got != want {
		t.Fatalf("expected path-derived id %q, got %#v", want, got)
	}
	if got, want := obj.Fields["body"], "# Schema\n\n| Column | Type | Description |\n| --- | --- | --- |\n| `order_id` | STRING | Globally unique order identifier. |\n\n# Joins\n\nJoined with the customer concept on `customer_id`.\n"; got != want {
		t.Fatalf("expected OKF body %q, got %#v", want, got)
	}

	okfObject, err := store.Get("Knowledge", "order-returns.okf")
	if err != nil {
		t.Fatalf("get .okf concept: %v", err)
	}
	if got, want := okfObject.Fields["body"], "# Returns\n\nOrders may be returned within 30 days of delivery.\n"; got != want {
		t.Fatalf("expected .okf body %q, got %#v", want, got)
	}
}

func exampleRoots(t *testing.T) []string {
	t.Helper()
	base := filepath.Join("..", "..", "examples")
	abs, err := filepath.Abs(base)
	if err != nil {
		t.Fatalf("abs path: %v", err)
	}
	entries, err := os.ReadDir(abs)
	if err != nil {
		t.Fatalf("readdir: %v", err)
	}
	var roots []string
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		dir := filepath.Join(abs, entry.Name())
		cfgPath := filepath.Join(dir, "mergeway.yaml")
		if _, err := os.Stat(cfgPath); err != nil {
			continue
		}
		roots = append(roots, dir)
	}
	if len(roots) == 0 {
		t.Fatalf("no example directories with mergeway.yaml under %s", abs)
	}
	sort.Strings(roots)
	return roots
}
