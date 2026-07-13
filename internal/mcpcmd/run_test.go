package mcpcmd

import (
	"bytes"
	"context"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunHelp(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := Run(context.Background(), []string{"--help"}, strings.NewReader(""), &stdout, &stderr, Options{})
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d", code)
	}
	if stdout.Len() != 0 {
		t.Fatalf("expected no stdout, got %q", stdout.String())
	}
	if !strings.Contains(stderr.String(), "Usage: mergeway-mcp [flags]") {
		t.Fatalf("expected help output, got %q", stderr.String())
	}
}

func TestRunPassesDefaultTransportAndResolvedRoot(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	var got Invocation
	started := false

	code := Run(context.Background(), []string{"--root", filepath.Join("..", "..", "examples", "full")}, strings.NewReader(""), &stdout, &stderr, Options{
		Start: func(_ context.Context, invocation Invocation) error {
			started = true
			got = invocation
			return nil
		},
	})
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d; stderr=%q", code, stderr.String())
	}
	if !started {
		t.Fatal("expected start function to be called")
	}

	wantRoot, err := filepath.Abs(filepath.Join("..", "..", "examples", "full"))
	if err != nil {
		t.Fatalf("resolve root: %v", err)
	}
	wantConfig, err := filepath.Abs(filepath.Join("..", "..", "examples", "full", "mergeway.yaml"))
	if err != nil {
		t.Fatalf("resolve config: %v", err)
	}

	if got.Root != wantRoot {
		t.Fatalf("expected root %q, got %q", wantRoot, got.Root)
	}
	if got.ConfigPath != wantConfig {
		t.Fatalf("expected config path %q, got %q", wantConfig, got.ConfigPath)
	}
	if got.Transport != "stdio" {
		t.Fatalf("expected default transport stdio, got %q", got.Transport)
	}
	if got.HTTPListen != defaultHTTPListen {
		t.Fatalf("expected default http listen %q, got %q", defaultHTTPListen, got.HTTPListen)
	}
	if got.HTTPBasePath != defaultHTTPBasePath {
		t.Fatalf("expected default http base path %q, got %q", defaultHTTPBasePath, got.HTTPBasePath)
	}
	if got.Stdin == nil || got.Stdout == nil || got.Stderr == nil {
		t.Fatal("expected invocation io to be populated")
	}
	if stderr.Len() != 0 {
		t.Fatalf("expected empty stderr, got %q", stderr.String())
	}
}

func TestRunRejectsInvalidTransport(t *testing.T) {
	code, _, stderr := runForTest(t, []string{"--root", filepath.Join("..", "..", "examples", "full"), "--transport", "socket"}, nil)
	if code != 2 {
		t.Fatalf("expected exit code 2, got %d", code)
	}
	if !strings.Contains(stderr, "invalid --transport") {
		t.Fatalf("expected invalid transport error, got %q", stderr)
	}
}

func TestRunRejectsHTTPOnlyFlagForStdio(t *testing.T) {
	code, _, stderr := runForTest(t, []string{"--root", filepath.Join("..", "..", "examples", "full"), "--http-listen", "127.0.0.1:9000"}, nil)
	if code != 2 {
		t.Fatalf("expected exit code 2, got %d", code)
	}
	if !strings.Contains(stderr, "--http-listen requires --transport=http") {
		t.Fatalf("expected http flag validation error, got %q", stderr)
	}
}

func TestRunRejectsUnknownEntity(t *testing.T) {
	code, _, stderr := runForTest(t, []string{"--root", filepath.Join("..", "..", "examples", "full"), "--entity", "Missing"}, nil)
	if code != 2 {
		t.Fatalf("expected exit code 2, got %d", code)
	}
	if !strings.Contains(stderr, `unknown --entity "Missing"`) {
		t.Fatalf("expected unknown entity error, got %q", stderr)
	}
}

func TestRunRejectsInvalidHTTPListen(t *testing.T) {
	code, _, stderr := runForTest(t, []string{"--root", filepath.Join("..", "..", "examples", "full"), "--transport", "http", "--http-listen", "bad listen"}, nil)
	if code != 2 {
		t.Fatalf("expected exit code 2, got %d", code)
	}
	if !strings.Contains(stderr, `invalid --http-listen "bad listen"`) {
		t.Fatalf("expected invalid http listen error, got %q", stderr)
	}
}

func TestRunRejectsInvalidHTTPBasePath(t *testing.T) {
	code, _, stderr := runForTest(t, []string{"--root", filepath.Join("..", "..", "examples", "full"), "--transport", "http", "--http-base-path", "mcp"}, nil)
	if code != 2 {
		t.Fatalf("expected exit code 2, got %d", code)
	}
	if !strings.Contains(stderr, `invalid --http-base-path "mcp"`) {
		t.Fatalf("expected invalid http base path error, got %q", stderr)
	}
}

func TestRunNormalizesHTTPBasePath(t *testing.T) {
	var got Invocation
	code, _, stderr := runForTest(t, []string{"--root", filepath.Join("..", "..", "examples", "full"), "--transport", "http", "--http-base-path", "/mcp/"}, func(_ context.Context, invocation Invocation) error {
		got = invocation
		return nil
	})
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d; stderr=%q", code, stderr)
	}
	if got.HTTPBasePath != "/mcp" {
		t.Fatalf("expected normalized base path /mcp, got %q", got.HTTPBasePath)
	}
}

func runForTest(t *testing.T, args []string, start StartFunc) (int, string, string) {
	t.Helper()

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run(context.Background(), args, strings.NewReader(""), &stdout, &stderr, Options{Start: start})
	return code, stdout.String(), stderr.String()
}
