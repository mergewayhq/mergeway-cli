# Mergeway VS Code Extension

This extension connects VS Code to a local `mergeway-lsp` binary.

## Requirements

You must build or install `mergeway-lsp` yourself and configure its path manually.

Example workspace setting:

```json
{
  "mergeway.lsp.path": "${workspaceFolder}/bin/mergeway-lsp"
}
```

`${workspaceFolder}` expands to the first open workspace folder before the extension validates the path.
You can also use a plain absolute path.

The extension activates only when the opened workspace contains `mergeway.yaml` or `mergeway.yml`.

## Development

Install dependencies:

```bash
npm install
```

Compile:

```bash
npm run compile
```

Run the extension:

1. Open the `vscode/` directory in VS Code.
2. Press `F5`.
3. In the Extension Development Host, open a workspace containing `mergeway.yaml` or `mergeway.yml`.
4. Configure `mergeway.lsp.path` in that workspace.
5. Open a YAML or JSON file managed by Mergeway.

## Packaging

Build a local VSIX package:

```bash
npm run package
```

This packages only the extension code. The VSIX does not bundle `mergeway-lsp`.

## Troubleshooting

### The extension does not activate

Make sure the opened workspace contains either:

- `mergeway.yaml`
- `mergeway.yml`

### The LSP does not start

Check that `mergeway.lsp.path`:

- is configured
- resolves to an absolute path
- points to an existing file

### The LSP starts but behaves strangely

Make sure the LSP server does not write logs to stdout. stdout is reserved for LSP protocol messages.
Use stderr or a log file instead.

### I need to restart the server after changing settings

Run `Mergeway: Restart Language Server` from the Command Palette.
