package cli

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/mergewayhq/mergeway-cli/internal/config"
	"github.com/mergewayhq/mergeway-cli/internal/data"
	"github.com/mergewayhq/mergeway-cli/internal/validation"
)

func loadConfig(ctx *Context) (*config.Config, error) {
	return config.Load(ctx.Config)
}

func loadStore(ctx *Context, cfg *config.Config) (*data.Store, error) {
	return data.NewStore(ctx.Root, cfg)
}

func readPayload(path string) (map[string]any, error) {
	var reader io.Reader

	if path == "" {
		reader = os.Stdin
	} else {
		f, err := os.Open(path)
		if err != nil {
			return nil, err
		}
		defer func() {
			_ = f.Close()
		}()
		reader = f
	}

	dataBytes, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	if len(dataBytes) == 0 {
		return map[string]any{}, nil
	}

	ext := strings.ToLower(filepath.Ext(path))

	var payload map[string]any
	switch ext {
	case ".json":
		if err := json.Unmarshal(dataBytes, &payload); err != nil {
			return nil, err
		}
	default:
		if err := yaml.Unmarshal(dataBytes, &payload); err != nil {
			return nil, err
		}
	}

	return payload, nil
}

func writeFormatted(ctx *Context, value any) int {
	switch ctx.Format {
	case "json":
		enc := json.NewEncoder(ctx.Stdout)
		enc.SetIndent("", "  ")
		if err := enc.Encode(value); err != nil {
			_, _ = fmt.Fprintf(ctx.Stderr, "encode json: %v\n", err)
			return 1
		}
	case "yaml":
		data, err := yaml.Marshal(value)
		if err != nil {
			_, _ = fmt.Fprintf(ctx.Stderr, "encode yaml: %v\n", err)
			return 1
		}
		if _, err := ctx.Stdout.Write(data); err != nil {
			_, _ = fmt.Fprintf(ctx.Stderr, "write output: %v\n", err)
			return 1
		}
	default:
		_, _ = fmt.Fprintf(ctx.Stderr, "unknown format %s\n", ctx.Format)
		return 1
	}
	return 0
}

func parseFilter(expr string) (string, string) {
	if expr == "" {
		return "", ""
	}
	parts := strings.SplitN(expr, "=", 2)
	if len(parts) != 2 {
		return "", ""
	}
	return parts[0], parts[1]
}

func confirm(in io.Reader, out io.Writer, prompt string) (bool, error) {
	if _, err := fmt.Fprint(out, prompt); err != nil {
		return false, err
	}
	scanner := bufio.NewScanner(in)
	if !scanner.Scan() {
		return false, scanner.Err()
	}
	answer := strings.ToLower(strings.TrimSpace(scanner.Text()))
	return answer == "y" || answer == "yes", nil
}

func (ctx *Context) Stdin() io.Reader {
	return os.Stdin
}

// multiFlag captures repeated --phase flags for validation commands.
type multiFlag struct {
	Values []validation.Phase
}

func (m *multiFlag) String() string {
	return strings.Join(phasesToStrings(m.Values), ",")
}

func (m *multiFlag) Set(value string) error {
	if value == "" {
		return nil
	}
	switch validation.Phase(value) {
	case validation.PhaseFormat, validation.PhaseSchema, validation.PhaseReferences:
		m.Values = append(m.Values, validation.Phase(value))
		return nil
	default:
		return fmt.Errorf("invalid phase %s", value)
	}
}

func phasesToStrings(phases []validation.Phase) []string {
	res := make([]string, len(phases))
	for i, p := range phases {
		res[i] = string(p)
	}
	return res
}
