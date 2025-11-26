package cli

import (
	"bytes"
	"flag"
	"strings"
	"testing"
)

func TestRunPrintsUsageWithoutArgs(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	code := Run([]string{}, stdout, stderr)
	if code == 0 {
		t.Fatalf("expected non-zero exit without command")
	}
	if !strings.Contains(stdout.String(), "Usage: mw") {
		t.Fatalf("expected usage text in stdout, got %s", stdout.String())
	}
}

func TestRunHandlesUnknownCommand(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	code := Run([]string{"unknown"}, stdout, stderr)
	if code == 0 {
		t.Fatalf("expected unknown command to fail")
	}
	if !strings.Contains(stderr.String(), "unknown command") {
		t.Fatalf("expected error message, got %s", stderr.String())
	}
}

func TestRunHelpFlag(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	code := Run([]string{"--help"}, stdout, stderr)
	if code != 0 {
		t.Fatalf("expected --help to exit 0, got %d", code)
	}
	if !strings.Contains(stdout.String(), "Commands:") {
		t.Fatalf("expected usage text for help flag")
	}
}

func TestDefaultFormattingHelpers(t *testing.T) {
	fs := flag.NewFlagSet("mw", flag.ContinueOnError)
	boolFlag := fs.Bool("verbose", false, "bool flag")
	stringFlag := fs.String("root", ".", "root dir")
	_ = *boolFlag
	_ = *stringFlag

	if shouldShowDefault(fs.Lookup("verbose")) {
		t.Fatalf("expected bool false default to be hidden")
	}
	if !shouldShowDefault(fs.Lookup("root")) {
		t.Fatalf("expected string flag default to be shown")
	}

	// Force bool default true and re-check formatting.
	fs.Bool("force", true, "force flag")

	if formatDefault(fs.Lookup("force")) != "true" {
		t.Fatalf("expected bool default to show bare true")
	}
	if formatDefault(fs.Lookup("root")) != "\".\"" {
		t.Fatalf("expected quoted default for string flag")
	}
}
