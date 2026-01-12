package cli

import (
	"bytes"
	"fmt"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"text/template"

	"github.com/mergewayhq/mergeway-cli/internal/config"
	"github.com/spf13/cobra"
)

const erdTemplate = `
digraph ERD {
    graph [pad="0.5", nodesep="1", ranksep="1" fontsize="10"];
    node [shape=plain fontsize="10" fontname="Arial"];
    edge [fontsize="10"];
    rankdir=LR;

    {{range .Types}}
    "{{.Name}}" [label=<
        <table border="0" cellborder="1" cellspacing="0" color="#666666">
            <tr><td bgcolor="#eeeeee" colspan="2"><b>{{.Name}}</b></td></tr>
            {{range .Fields}}
            <tr>
                <td align="left" valign="middle" sides="BL"><font color="#666666">{{.Name}}</font></td>
                <td align="left" valign="middle" sides="BR"><font color="#666666">{{.Type}}</font></td>
            </tr>
            {{end}}
            {{if .Paths}}
            <tr>
                <td colspan="2" align="left" sides="T"><font color="#666666">{{range .Paths}}{{.}}<br/>{{end}}</font></td>
            </tr>
            {{end}}
        </table>
    >];
    {{end}}

    {{range .Edges}}
    "{{.Source}}" -> "{{.Target}}" [label="{{.Label}}"];
    {{end}}
}
`

type erdType struct {
	Name   string
	Fields []erdField
	Paths  []string
}

type erdField struct {
	Name string
	Type string
}

type erdEdge struct {
	Source string
	Target string
	Label  string
}

type erdData struct {
	Types []erdType
	Edges []erdEdge
}

func newGenERDCommand() *cobra.Command {
	var path string

	cmd := &cobra.Command{
		Use:   "gen-erd",
		Short: "Generate an entity-relationship diagram",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, err := contextFromCommand(cmd)
			if err != nil {
				return err
			}

			if path == "" {
				_, _ = fmt.Fprintln(ctx.Stderr, "gen-erd: --path is required")
				return newExitError(1)
			}

			cfg, err := loadConfig(ctx)
			if err != nil {
				_, _ = fmt.Fprintf(ctx.Stderr, "gen-erd: %v\n", err)
				return newExitError(1)
			}

			data := prepareERDData(cfg)

			tmpl, err := template.New("erd").Parse(erdTemplate)
			if err != nil {
				_, _ = fmt.Fprintf(ctx.Stderr, "gen-erd: parse template: %v\n", err)
				return newExitError(1)
			}

			var buf bytes.Buffer
			if err := tmpl.Execute(&buf, data); err != nil {
				_, _ = fmt.Fprintf(ctx.Stderr, "gen-erd: execute template: %v\n", err)
				return newExitError(1)
			}

			ext := strings.TrimPrefix(filepath.Ext(path), ".")
			if ext == "" {
				ext = "png"
			}

			runCmd := exec.Command("dot", "-T"+ext, "-o", path)
			runCmd.Stdin = &buf
			runCmd.Stdout = ctx.Stdout
			runCmd.Stderr = ctx.Stderr

			if err := runCmd.Run(); err != nil {
				_, _ = fmt.Fprintf(ctx.Stderr, "gen-erd: graphviz execution failed: %v\n", err)
				return newExitError(1)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&path, "path", "", "Output path for the generated image")

	return cmd
}

func prepareERDData(cfg *config.Config) erdData {
	var types []erdType
	var edges []erdEdge
	typeNames := make(map[string]struct{})

	// First pass: collect type names
	for name := range cfg.Types {
		typeNames[name] = struct{}{}
	}

	// Collect types and edges
	for name, def := range cfg.Types {
		t := erdType{
			Name: name,
		}

		// Collect fields
		var fields []erdField
		for fName, fDef := range def.Fields {
			fields = append(fields, erdField{Name: fName, Type: fDef.Type})

			// Infer edges
			if _, ok := typeNames[fDef.Type]; ok {
				edges = append(edges, erdEdge{
					Source: name,
					Target: fDef.Type,
					Label:  fName,
				})
			}
		}
		// Sort fields for consistent output
		sort.Slice(fields, func(i, j int) bool {
			return fields[i].Name < fields[j].Name
		})
		t.Fields = fields

		// Collect paths from Include
		for _, inc := range def.Include {
			if inc.Path != "" {
				t.Paths = append(t.Paths, inc.Path)
			}
		}
		sort.Strings(t.Paths)

		types = append(types, t)
	}

	sort.Slice(types, func(i, j int) bool {
		return types[i].Name < types[j].Name
	})

	sort.Slice(edges, func(i, j int) bool {
		if edges[i].Source != edges[j].Source {
			return edges[i].Source < edges[j].Source
		}
		return edges[i].Target < edges[j].Target
	})

	return erdData{Types: types, Edges: edges}
}
