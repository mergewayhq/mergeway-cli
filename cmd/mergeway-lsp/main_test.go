package main

import (
	"bytes"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunHelpUsesStderrAndLeavesStdoutProtocolOnly(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	if code := run([]string{"--help"}, strings.NewReader(""), &stdout, &stderr); code != 2 {
		t.Fatalf("run(--help) exit code = %d, want 2", code)
	}
	if stdout.Len() != 0 {
		t.Fatalf("help wrote to stdout: %q", stdout.String())
	}
	for _, flag := range []string{"-log-file", "-log-level", "-log-stderr"} {
		if !strings.Contains(stderr.String(), flag) {
			t.Fatalf("help does not contain %s:\n%s", flag, stderr.String())
		}
	}
}

func TestNewLoggerReadsEnvironment(t *testing.T) {
	logPath := filepath.Join(t.TempDir(), "environment.log")
	t.Setenv("MERGEWAY_LSP_LOG_FILE", logPath)
	t.Setenv("MERGEWAY_LSP_LOG_LEVEL", "debug")
	t.Setenv("MERGEWAY_LSP_LOG_STDERR", "true")
	var stderr bytes.Buffer

	logger, closer, err := newLogger(&stderr, loggerOptions{})
	if err != nil {
		t.Fatalf("newLogger: %v", err)
	}
	logger.Debug("environment-selected")
	if err := closer.Close(); err != nil {
		t.Fatalf("close logger: %v", err)
	}

	body, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("read environment log: %v", err)
	}
	if !bytes.Contains(body, []byte("environment-selected")) {
		t.Fatalf("environment log does not contain debug message: %q", body)
	}
	if stderr.Len() != 0 {
		t.Fatalf("log file should take priority over environment stderr: %q", stderr.String())
	}
}

func TestNewLoggerFlagsOverrideFileAndLevelEnvironment(t *testing.T) {
	tmp := t.TempDir()
	environmentPath := filepath.Join(tmp, "environment.log")
	flagPath := filepath.Join(tmp, "flag.log")
	t.Setenv("MERGEWAY_LSP_LOG_FILE", environmentPath)
	t.Setenv("MERGEWAY_LSP_LOG_LEVEL", "error")
	var stderr bytes.Buffer

	logger, closer, err := newLogger(&stderr, loggerOptions{
		LogFile:  flagPath,
		LogLevel: "debug",
	})
	if err != nil {
		t.Fatalf("newLogger: %v", err)
	}
	logger.Debug("flag-selected")
	if err := closer.Close(); err != nil {
		t.Fatalf("close logger: %v", err)
	}

	body, err := os.ReadFile(flagPath)
	if err != nil {
		t.Fatalf("read flag log: %v", err)
	}
	if !bytes.Contains(body, []byte("flag-selected")) {
		t.Fatalf("flag log does not contain debug message: %q", body)
	}
	if _, err := os.Stat(environmentPath); !os.IsNotExist(err) {
		t.Fatalf("environment log should not be selected, stat error = %v", err)
	}
}

func TestNewLoggerStderrSelectionAndDefaultLevel(t *testing.T) {
	t.Setenv("MERGEWAY_LSP_LOG_FILE", "")
	t.Setenv("MERGEWAY_LSP_LOG_LEVEL", "")
	t.Setenv("MERGEWAY_LSP_LOG_STDERR", "yes")
	var stderr bytes.Buffer

	logger, closer, err := newLogger(&stderr, loggerOptions{})
	if err != nil {
		t.Fatalf("newLogger: %v", err)
	}
	if closer != nil {
		t.Fatal("stderr logger returned an unexpected closer")
	}
	logger.Debug("hidden-debug")
	logger.LogAttrs(nil, slog.LevelInfo, "visible-info")

	if strings.Contains(stderr.String(), "hidden-debug") || !strings.Contains(stderr.String(), "visible-info") {
		t.Fatalf("unexpected default-level stderr log: %q", stderr.String())
	}
}

func TestNewLoggerRejectsInvalidEnvironmentLevel(t *testing.T) {
	t.Setenv("MERGEWAY_LSP_LOG_LEVEL", "verbose")
	if _, _, err := newLogger(&bytes.Buffer{}, loggerOptions{}); err == nil || !strings.Contains(err.Error(), `invalid log level "verbose"`) {
		t.Fatalf("newLogger error = %v, want invalid environment level", err)
	}
}
