package cli

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestReadPayloadParsesYAML(t *testing.T) {
	path := filepath.Join(t.TempDir(), "payload.yaml")
	if err := os.WriteFile(path, []byte("name: Example\n"), 0o644); err != nil {
		t.Fatalf("write payload: %v", err)
	}

	payload, err := readPayload(path)
	if err != nil {
		t.Fatalf("readPayload: %v", err)
	}
	if payload["name"] != "Example" {
		t.Fatalf("expected field, got %v", payload["name"])
	}
}

func TestReadPayloadParsesJSON(t *testing.T) {
	path := filepath.Join(t.TempDir(), "payload.json")
	if err := os.WriteFile(path, []byte(`{"name":"Json"}`), 0o644); err != nil {
		t.Fatalf("write payload: %v", err)
	}

	payload, err := readPayload(path)
	if err != nil {
		t.Fatalf("readPayload json: %v", err)
	}
	if payload["name"] != "Json" {
		t.Fatalf("expected json field, got %v", payload["name"])
	}
}

func TestReadPayloadFromStdin(t *testing.T) {
	withStdin(t, "name: Inline\n", func() {
		payload, err := readPayload("")
		if err != nil {
			t.Fatalf("readPayload stdin: %v", err)
		}
		if payload["name"] != "Inline" {
			t.Fatalf("expected stdin payload, got %v", payload["name"])
		}
	})
}

func TestWriteFormattedJSON(t *testing.T) {
	ctx := &Context{
		Format: "json",
		Stdout: &bytes.Buffer{},
		Stderr: &bytes.Buffer{},
	}
	code := writeFormatted(ctx, map[string]any{"foo": "bar"})
	if code != 0 {
		t.Fatalf("writeFormatted json exit %d stderr %s", code, ctx.Stderr.(*bytes.Buffer).String())
	}
	if body := ctx.Stdout.(*bytes.Buffer).String(); !strings.Contains(body, "\"foo\": \"bar\"") {
		t.Fatalf("expected json output, got %s", body)
	}
}

func TestWriteFormattedYAML(t *testing.T) {
	ctx := &Context{
		Format: "yaml",
		Stdout: &bytes.Buffer{},
		Stderr: &bytes.Buffer{},
	}
	code := writeFormatted(ctx, map[string]any{"foo": "bar"})
	if code != 0 {
		t.Fatalf("writeFormatted yaml exit %d", code)
	}
	if body := ctx.Stdout.(*bytes.Buffer).String(); !strings.Contains(body, "foo: bar") {
		t.Fatalf("expected yaml output, got %s", body)
	}
}

func TestWriteFormattedUnknownFormat(t *testing.T) {
	ctx := &Context{
		Format: "xml",
		Stdout: &bytes.Buffer{},
		Stderr: &bytes.Buffer{},
	}
	code := writeFormatted(ctx, map[string]any{"foo": "bar"})
	if code == 0 {
		t.Fatalf("expected unknown format to fail")
	}
	if !strings.Contains(ctx.Stderr.(*bytes.Buffer).String(), "unknown format") {
		t.Fatalf("expected error message, got %s", ctx.Stderr.(*bytes.Buffer).String())
	}
}

func TestConfirmResponses(t *testing.T) {
	cases := []struct {
		name     string
		input    string
		expected bool
	}{
		{name: "yes", input: "y\n", expected: true},
		{name: "word", input: "YES\n", expected: true},
		{name: "no", input: "n\n", expected: false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			ok, err := confirm(strings.NewReader(tc.input), buf, "prompt")
			if err != nil {
				t.Fatalf("confirm error: %v", err)
			}
			if ok != tc.expected {
				t.Fatalf("expected %v, got %v", tc.expected, ok)
			}
			if buf.String() != "prompt" {
				t.Fatalf("expected prompt write, got %s", buf.String())
			}
		})
	}
}

func TestConfirmWriterError(t *testing.T) {
	_, err := confirm(strings.NewReader("y\n"), failingWriter{}, "prompt")
	if err == nil {
		t.Fatalf("expected error when prompt write fails")
	}
}

type failingWriter struct{}

func (failingWriter) Write([]byte) (int, error) {
	return 0, io.ErrClosedPipe
}

func TestMultiFlagSetAndString(t *testing.T) {
	var m multiFlag
	if err := m.Set("format"); err != nil {
		t.Fatalf("set format: %v", err)
	}
	if err := m.Set("schema"); err != nil {
		t.Fatalf("set schema: %v", err)
	}
	if err := m.Set("references"); err != nil {
		t.Fatalf("set references: %v", err)
	}
	if got := m.String(); got != "format,schema,references" {
		t.Fatalf("expected all phases, got %s", got)
	}
	if err := m.Set("unknown"); err == nil {
		t.Fatalf("expected invalid phase to error")
	}
}
