// Command mcp-example-client captures deterministic documentation examples from
// mergeway-mcp. It is documentation test tooling, not a shipped binary.
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strings"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"
)

type example struct {
	Tool    string         `json:"tool"`
	Request map[string]any `json:"request"`
	Result  any            `json:"result"`
}

type capture struct {
	ProtocolVersion string    `json:"protocol_version"`
	Tools           []string  `json:"tools"`
	Examples        []example `json:"examples"`
}

func main() {
	server := flag.String("server", "", "path to mergeway-mcp")
	root := flag.String("root", "", "repository root")
	endpoint := flag.String("endpoint", "", "streamable HTTP endpoint instead of stdio")
	flag.Parse()

	if *root == "" || (*server == "" && *endpoint == "") || (*server != "" && *endpoint != "") {
		fmt.Fprintln(os.Stderr, "usage: mcp-example-client --root <path> (--server <path> | --endpoint <url>)")
		os.Exit(2)
	}

	ctx := context.Background()
	var transport sdkmcp.Transport
	if *endpoint != "" {
		transport = &sdkmcp.StreamableClientTransport{Endpoint: *endpoint}
	} else {
		command := exec.Command(*server, "--root", *root)
		command.Stderr = os.Stderr
		transport = &sdkmcp.CommandTransport{Command: command}
	}

	client := sdkmcp.NewClient(&sdkmcp.Implementation{Name: "mergeway-docs-check", Version: "1"}, nil)
	session, err := client.Connect(ctx, transport, nil)
	check(err)
	defer func() { check(session.Close()) }()

	var tools []string
	for tool, err := range session.Tools(ctx, nil) {
		check(err)
		tools = append(tools, tool.Name)
	}
	sort.Strings(tools)

	cases := []struct {
		name string
		args map[string]any
	}{
		{name: "entity_list", args: map[string]any{}},
		{name: "entity_show", args: map[string]any{"entity": "User"}},
		{name: "object_list", args: map[string]any{"entity": "User"}},
		{name: "object_get", args: map[string]any{"entity": "User", "id": "user-alice"}},
		{name: "repository_export", args: map[string]any{"entities": []string{"Post"}}},
		{name: "files_list", args: map[string]any{"entity": "User"}},
	}

	result := capture{
		ProtocolVersion: session.InitializeResult().ProtocolVersion,
		Tools:           tools,
		Examples:        make([]example, 0, len(cases)),
	}
	for _, tc := range cases {
		response, err := session.CallTool(ctx, &sdkmcp.CallToolParams{Name: tc.name, Arguments: tc.args})
		check(err)
		if response.IsError {
			fmt.Fprintf(os.Stderr, "%s returned a tool error\n", tc.name)
			os.Exit(1)
		}
		var structured any
		raw, err := json.Marshal(response.StructuredContent)
		check(err)
		check(json.Unmarshal(raw, &structured))
		result.Examples = append(result.Examples, example{
			Tool:    tc.name,
			Request: tc.args,
			Result:  normalize(structured, *root),
		})
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	encoder.SetEscapeHTML(false)
	check(encoder.Encode(result))
}

func normalize(value any, root string) any {
	switch typed := value.(type) {
	case string:
		return strings.ReplaceAll(typed, root, "<ROOT>")
	case []any:
		for index := range typed {
			typed[index] = normalize(typed[index], root)
		}
		return typed
	case map[string]any:
		for key, item := range typed {
			typed[key] = normalize(item, root)
		}
		return typed
	default:
		return value
	}
}

func check(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
