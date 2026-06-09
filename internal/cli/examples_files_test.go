package cli

import (
	"bytes"
	"encoding/json"
	"path/filepath"
	"reflect"
	"testing"
)

func TestFilesExamplesJSON(t *testing.T) {
	for _, root := range exampleRoots(t) {
		root := root
		t.Run(filepath.Base(root), func(t *testing.T) {
			stdout := &bytes.Buffer{}
			stderr := &bytes.Buffer{}

			code := Run([]string{"--root", root, "--format", "json", "files"}, stdout, stderr)
			if code != 0 {
				t.Fatalf("files --format json exit %d stdout %s stderr %s", code, stdout.String(), stderr.String())
			}

			var entries []map[string]string
			if err := json.Unmarshal(stdout.Bytes(), &entries); err != nil {
				t.Fatalf("expected json output, got parse error: %v\nbody:\n%s", err, stdout.String())
			}
		})
	}
}

func TestFilesFullExample(t *testing.T) {
	root, err := filepath.Abs(filepath.Join("..", "..", "examples", "full"))
	if err != nil {
		t.Fatalf("abs path: %v", err)
	}
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	code := Run([]string{"--root", root, "--format", "json", "files"}, stdout, stderr)
	if code != 0 {
		t.Fatalf("files full example exit %d stdout %s stderr %s", code, stdout.String(), stderr.String())
	}

	var entries []map[string]string
	if err := json.Unmarshal(stdout.Bytes(), &entries); err != nil {
		t.Fatalf("expected json output, got parse error: %v\nbody:\n%s", err, stdout.String())
	}

	expected := []map[string]string{
		{"type": "Post", "file": "data/posts/launch.yaml"},
		{"type": "Tag", "file": "data/tags/product.yaml"},
		{"type": "User", "file": "data/users/alice.yaml"},
	}
	if !reflect.DeepEqual(entries, expected) {
		t.Fatalf("expected %v, got %v", expected, entries)
	}
}

func TestFilesExternalRootPathExample(t *testing.T) {
	root, err := filepath.Abs(filepath.Join("..", "..", "examples", "external-root-path", "primary"))
	if err != nil {
		t.Fatalf("abs path: %v", err)
	}
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	code := Run([]string{"--root", root, "--format", "json", "files"}, stdout, stderr)
	if code != 0 {
		t.Fatalf("files external-root example exit %d stdout %s stderr %s", code, stdout.String(), stderr.String())
	}

	var entries []map[string]string
	if err := json.Unmarshal(stdout.Bytes(), &entries); err != nil {
		t.Fatalf("expected json output, got parse error: %v\nbody:\n%s", err, stdout.String())
	}

	expected := []map[string]string{
		{"type": "OrderLine", "file": "data/order-lines/line-1001.yaml"},
		{"type": "OrderLine", "file": "data/order-lines/line-1002.yaml"},
		{"type": "Product", "file": "../secondary/products/gadget.yaml"},
		{"type": "Product", "file": "../secondary/products/widget.yaml"},
	}
	if !reflect.DeepEqual(entries, expected) {
		t.Fatalf("expected %v, got %v", expected, entries)
	}
}
