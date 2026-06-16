package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"

	"github.com/mergewayhq/mergeway-cli/internal/lsp"
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr))
}

func run(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("mergeway-lsp", flag.ContinueOnError)
	fs.SetOutput(stderr)

	logFile := fs.String("log-file", "", "Write LSP logs to a file")
	logLevel := fs.String("log-level", "", "Log level (debug|info|warn|error)")
	logStderr := fs.Bool("log-stderr", false, "Write LSP logs to stderr")

	if err := fs.Parse(args); err != nil {
		return 2
	}

	logger, closer, err := newLogger(stderr, loggerOptions{
		LogFile:   *logFile,
		LogLevel:  *logLevel,
		LogStderr: *logStderr,
	})
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "mergeway-lsp: %v\n", err)
		return 1
	}
	defer func() {
		if closer != nil {
			_ = closer.Close()
		}
	}()

	code, err := lsp.Run(context.Background(), stdioReadWriteCloser{
		Reader: stdin,
		Writer: stdout,
	}, lsp.Options{Logger: logger})
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "mergeway-lsp: %v\n", err)
		return 1
	}
	return code
}

type loggerOptions struct {
	LogFile   string
	LogLevel  string
	LogStderr bool
}

func newLogger(stderr io.Writer, opts loggerOptions) (*slog.Logger, io.Closer, error) {
	logFile := firstNonEmpty(opts.LogFile, os.Getenv("MERGEWAY_LSP_LOG_FILE"))
	levelName := firstNonEmpty(opts.LogLevel, os.Getenv("MERGEWAY_LSP_LOG_LEVEL"))
	useStderr := opts.LogStderr || envTruthy(os.Getenv("MERGEWAY_LSP_LOG_STDERR"))

	level := new(slog.LevelVar)
	if err := level.UnmarshalText([]byte(defaultLogLevel(levelName))); err != nil {
		return nil, nil, fmt.Errorf("invalid log level %q", levelName)
	}

	var writer = io.Discard
	var closer io.Closer

	switch {
	case logFile != "":
		f, err := os.OpenFile(logFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
		if err != nil {
			return nil, nil, fmt.Errorf("open log file %s: %w", logFile, err)
		}
		writer = f
		closer = f
	case useStderr:
		writer = stderr
	}

	handler := slog.NewTextHandler(writer, &slog.HandlerOptions{Level: level})
	return slog.New(handler), closer, nil
}

func defaultLogLevel(name string) string {
	if strings.TrimSpace(name) == "" {
		return "info"
	}
	return name
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func envTruthy(value string) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "1", "true", "yes", "on":
		return true
	default:
		return false
	}
}

type stdioReadWriteCloser struct {
	io.Reader
	io.Writer
}

func (stdioReadWriteCloser) Close() error {
	return nil
}
