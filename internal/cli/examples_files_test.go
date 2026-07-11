package cli

import (
	"bytes"
	"encoding/json"
	"path/filepath"
	"reflect"
	"strings"
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

func TestInheritanceExampleCommands(t *testing.T) {
	root, err := filepath.Abs(filepath.Join("..", "..", "examples", "inheritance"))
	if err != nil {
		t.Fatalf("abs path: %v", err)
	}

	validateOut := &bytes.Buffer{}
	validateErr := &bytes.Buffer{}
	code := Run([]string{"--root", root, "validate"}, validateOut, validateErr)
	if code != 0 {
		t.Fatalf("validate inheritance example exit %d stdout %s stderr %s", code, validateOut.String(), validateErr.String())
	}

	listOut := &bytes.Buffer{}
	listErr := &bytes.Buffer{}
	code = Run([]string{"--root", root, "list", "--type", "Animal"}, listOut, listErr)
	if code != 0 {
		t.Fatalf("list inheritance example exit %d stdout %s stderr %s", code, listOut.String(), listErr.String())
	}
	if got := strings.Fields(listOut.String()); !reflect.DeepEqual(got, []string{"dog-1"}) {
		t.Fatalf("expected inherited parent list [dog-1], got %v", got)
	}

	getOut := &bytes.Buffer{}
	getErr := &bytes.Buffer{}
	code = Run([]string{"--root", root, "--format", "json", "get", "--type", "Animal", "dog-1"}, getOut, getErr)
	if code != 0 {
		t.Fatalf("get inheritance example exit %d stdout %s stderr %s", code, getOut.String(), getErr.String())
	}

	var payload map[string]any
	if err := json.Unmarshal(getOut.Bytes(), &payload); err != nil {
		t.Fatalf("expected json output, got parse error: %v\nbody:\n%s", err, getOut.String())
	}
	if payload["breed"] != "collie" || payload["name"] != "Fido" {
		t.Fatalf("expected inherited get payload, got %v", payload)
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

func TestFilesFullExampleGroupsContainers(t *testing.T) {
	root, err := filepath.Abs(filepath.Join("..", "..", "examples", "full"))
	if err != nil {
		t.Fatalf("abs path: %v", err)
	}
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	code := Run([]string{"--root", root, "--format", "json", "files", "--group"}, stdout, stderr)
	if code != 0 {
		t.Fatalf("files --group full example exit %d stdout %s stderr %s", code, stdout.String(), stderr.String())
	}

	var entries []map[string]string
	if err := json.Unmarshal(stdout.Bytes(), &entries); err != nil {
		t.Fatalf("expected json output, got parse error: %v\nbody:\n%s", err, stdout.String())
	}

	expected := []map[string]string{
		{"type": "Post", "file": "data/posts/*.yaml"},
		{"type": "Tag", "file": "data/tags/product.yaml"},
		{"type": "User", "file": "data/users/*.yaml"},
	}
	if !reflect.DeepEqual(entries, expected) {
		t.Fatalf("expected %v, got %v", expected, entries)
	}
}

func TestFilesExternalRootPathExampleGroupsContainers(t *testing.T) {
	root, err := filepath.Abs(filepath.Join("..", "..", "examples", "external-root-path", "primary"))
	if err != nil {
		t.Fatalf("abs path: %v", err)
	}
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	code := Run([]string{"--root", root, "--format", "json", "files", "--group"}, stdout, stderr)
	if code != 0 {
		t.Fatalf("files --group external-root example exit %d stdout %s stderr %s", code, stdout.String(), stderr.String())
	}

	var entries []map[string]string
	if err := json.Unmarshal(stdout.Bytes(), &entries); err != nil {
		t.Fatalf("expected json output, got parse error: %v\nbody:\n%s", err, stdout.String())
	}

	expected := []map[string]string{
		{"type": "OrderLine", "file": "data/order-lines/*.yaml"},
		{"type": "Product", "file": "../secondary/products/*.yaml"},
	}
	if !reflect.DeepEqual(entries, expected) {
		t.Fatalf("expected %v, got %v", expected, entries)
	}
}
