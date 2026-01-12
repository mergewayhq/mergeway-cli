package cli

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
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

func emitList(ctx *Context, store *data.Store, typeName, filterExpr string) error {
	key, value := parseFilter(filterExpr)

	if key == "" {
		// Fast path: identifier-only listing can use the store summary without
		// decoding every record.
		ids, err := store.List(typeName)
		if err != nil {
			return err
		}
		for _, id := range ids {
			_, _ = fmt.Fprintln(ctx.Stdout, id)
		}
		return nil
	}

	// Fallback to loading full objects only when filtering is needed so we avoid
	// decoding every record when the caller just needs identifiers.
	// Filter needs object fields, so load the dataset despite the heavier cost.
	objects, err := store.LoadAll(typeName)
	if err != nil {
		return err
	}

	var filtered []string
	for _, obj := range objects {
		if val, ok := obj.Fields[key]; ok && fmt.Sprint(val) == value {
			filtered = append(filtered, obj.ID)
		}
	}
	// Emit identifiers deterministically so list output is stable even when
	// filtering narrows down the set.
	sort.Strings(filtered)
	for _, id := range filtered {
		_, _ = fmt.Fprintln(ctx.Stdout, id)
	}
	return nil
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

func (m *multiFlag) Type() string {
	return "phase"
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
