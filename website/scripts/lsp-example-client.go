// Command lsp-example-client verifies the documented stdio initialization and
// diagnostics contract against a mergeway-lsp executable.
package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

type message struct {
	ID     int             `json:"id"`
	Method string          `json:"method"`
	Params json.RawMessage `json:"params"`
	Result json.RawMessage `json:"result"`
}

type initializeResult struct {
	Capabilities struct {
		TextDocumentSync struct {
			OpenClose bool `json:"openClose"`
			Change    int  `json:"change"`
		} `json:"textDocumentSync"`
		CompletionProvider      map[string]any `json:"completionProvider"`
		HoverProvider           bool           `json:"hoverProvider"`
		DefinitionProvider      bool           `json:"definitionProvider"`
		ReferencesProvider      bool           `json:"referencesProvider"`
		DocumentSymbolProvider  bool           `json:"documentSymbolProvider"`
		WorkspaceSymbolProvider bool           `json:"workspaceSymbolProvider"`
		CodeActionProvider      bool           `json:"codeActionProvider"`
		Workspace               struct {
			WorkspaceFolders struct {
				Supported           bool `json:"supported"`
				ChangeNotifications bool `json:"changeNotifications"`
			} `json:"workspaceFolders"`
		} `json:"workspace"`
	} `json:"capabilities"`
	ServerInfo struct {
		Name    string `json:"name"`
		Version string `json:"version"`
	} `json:"serverInfo"`
}

type publishDiagnosticsParams struct {
	URI         string `json:"uri"`
	Diagnostics []struct {
		Code     string `json:"code"`
		Source   string `json:"source"`
		Message  string `json:"message"`
		Severity int    `json:"severity"`
	} `json:"diagnostics"`
}

func main() {
	server := flag.String("server", "", "path to mergeway-lsp")
	root := flag.String("root", "", "path to a known-invalid Mergeway repository")
	flag.Parse()
	if *server == "" || *root == "" {
		fatalf("--server and --root are required")
	}

	resolvedRoot, err := filepath.Abs(*root)
	if err != nil {
		fatalf("resolve root: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, *server)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		fatalf("open server stdin: %v", err)
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		fatalf("open server stdout: %v", err)
	}
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Start(); err != nil {
		fatalf("start server: %v", err)
	}

	reader := bufio.NewReader(stdout)
	rootURI := (&url.URL{Scheme: "file", Path: filepath.ToSlash(resolvedRoot)}).String()
	writeFrame(stdin, map[string]any{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "initialize",
		"params": map[string]any{
			"workspaceFolders": []map[string]string{{"uri": rootURI, "name": "invalid-fixture"}},
		},
	})

	var initialized initializeResult
	var diagnostics []publishDiagnosticsParams
	for {
		frame := readFrame(reader)
		var msg message
		if err := json.Unmarshal(frame, &msg); err != nil {
			fatalf("decode server frame: %v", err)
		}
		if msg.Method == "textDocument/publishDiagnostics" {
			var params publishDiagnosticsParams
			if err := json.Unmarshal(msg.Params, &params); err != nil {
				fatalf("decode diagnostics: %v", err)
			}
			diagnostics = append(diagnostics, params)
		}
		if msg.ID == 1 {
			if err := json.Unmarshal(msg.Result, &initialized); err != nil {
				fatalf("decode initialize result: %v", err)
			}
			break
		}
	}

	assertCapabilities(initialized)
	codes := diagnosticCodes(diagnostics)
	if len(codes) == 0 {
		fatalf("server published no diagnostics for %s", resolvedRoot)
	}

	writeFrame(stdin, map[string]any{"jsonrpc": "2.0", "id": 2, "method": "shutdown"})
	for {
		frame := readFrame(reader)
		var msg message
		if err := json.Unmarshal(frame, &msg); err != nil {
			fatalf("decode shutdown frame: %v", err)
		}
		if msg.ID == 2 {
			break
		}
	}
	writeFrame(stdin, map[string]any{"jsonrpc": "2.0", "method": "exit"})
	_ = stdin.Close()
	if err := cmd.Wait(); err != nil {
		fatalf("server exit: %v; stderr: %s", err, stderr.String())
	}
	if stderr.Len() != 0 {
		fatalf("server contaminated diagnostics channel: %s", stderr.String())
	}

	fmt.Printf("server=%s version=%s sync=full workspace-folders=static\n", initialized.ServerInfo.Name, initialized.ServerInfo.Version)
	fmt.Println("capabilities=completion,hover,definition,references,document-symbols,workspace-symbols,code-actions")
	fmt.Printf("diagnostic-codes=%s\n", strings.Join(codes, ","))
}

func assertCapabilities(result initializeResult) {
	c := result.Capabilities
	if result.ServerInfo.Name != "mergeway-lsp" || !c.TextDocumentSync.OpenClose || c.TextDocumentSync.Change != 1 || c.CompletionProvider == nil ||
		!c.HoverProvider || !c.DefinitionProvider || !c.ReferencesProvider || !c.DocumentSymbolProvider || !c.WorkspaceSymbolProvider || !c.CodeActionProvider ||
		!c.Workspace.WorkspaceFolders.Supported || c.Workspace.WorkspaceFolders.ChangeNotifications {
		fatalf("unexpected initialize result: %+v", result)
	}
}

func diagnosticCodes(all []publishDiagnosticsParams) []string {
	seen := make(map[string]struct{})
	for _, params := range all {
		for _, diagnostic := range params.Diagnostics {
			if diagnostic.Source != "mergeway" || diagnostic.Severity != 1 {
				fatalf("unexpected diagnostic source/severity: %+v", diagnostic)
			}
			if diagnostic.Code != "" {
				seen[diagnostic.Code] = struct{}{}
			}
		}
	}
	result := make([]string, 0, len(seen))
	for code := range seen {
		result = append(result, code)
	}
	sort.Strings(result)
	return result
}

func writeFrame(w io.Writer, payload any) {
	body, err := json.Marshal(payload)
	if err != nil {
		fatalf("encode client frame: %v", err)
	}
	if _, err := fmt.Fprintf(w, "Content-Length: %d\r\n\r\n", len(body)); err != nil {
		fatalf("write client header: %v", err)
	}
	if _, err := w.Write(body); err != nil {
		fatalf("write client body: %v", err)
	}
}

func readFrame(r *bufio.Reader) []byte {
	contentLength := -1
	for {
		line, err := r.ReadString('\n')
		if err != nil {
			fatalf("read server header: %v", err)
		}
		line = strings.TrimSpace(line)
		if line == "" {
			break
		}
		parts := strings.SplitN(line, ":", 2)
		if len(parts) == 2 && strings.EqualFold(strings.TrimSpace(parts[0]), "Content-Length") {
			contentLength, err = strconv.Atoi(strings.TrimSpace(parts[1]))
			if err != nil {
				fatalf("parse Content-Length: %v", err)
			}
		}
	}
	if contentLength < 0 {
		fatalf("server frame has no Content-Length")
	}
	body := make([]byte, contentLength)
	if _, err := io.ReadFull(r, body); err != nil {
		fatalf("read server body: %v", err)
	}
	return body
}

func fatalf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}
