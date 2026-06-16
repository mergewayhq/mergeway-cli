---
title: "Set up mergeway-lsp in VS Code and Neovim"
linkTitle: "Set up mergeway-lsp in VS Code and Neovim"
description: "Configure the Mergeway VS Code extension or a generic LSP client to launch mergeway-lsp."
weight: 35
---

This guide covers the current editor setup paths for `mergeway-lsp`.

## Before You Start

- Put `mergeway-lsp` on your `PATH`, or use an absolute path to the binary.
- Open a folder that contains `mergeway.yaml` or `mergeway.yml`.
- Keep stdout reserved for protocol traffic. Send logs to stderr or a log file instead.

## VS Code

Use the extension in `vscode/` and follow `vscode/README.md`, then point it at your local `mergeway-lsp` binary.

Example workspace settings:

```json
{
  "mergeway.lsp.path": "/absolute/path/to/mergeway-lsp",
  "mergeway.lsp.trace.server": "off"
}
```

The extension activates only when the opened workspace contains `mergeway.yaml` or `mergeway.yml`.

## VS Code Generic LSP Client

If you prefer a generic LSP launcher extension, configure it to run `mergeway-lsp`.

Example settings shape:

```json
{
  "mergeway.server.command": "mergeway-lsp",
  "mergeway.server.args": [
    "--log-file",
    "/tmp/mergeway-lsp.log"
  ],
  "mergeway.server.filetypes": [
    "yaml",
    "json"
  ],
  "mergeway.server.rootPatterns": [
    "mergeway.yaml",
    "mergeway.yml"
  ]
}
```

Adjust the exact setting names to match the generic LSP client extension you are using. The important pieces are the command, optional debug args, filetypes, and root markers.

## Neovim

With `nvim-lspconfig`, you can register the server directly:

```lua
local lspconfig = require("lspconfig")
local util = require("lspconfig.util")

lspconfig.mergeway_lsp = {
  default_config = {
    cmd = { "mergeway-lsp", "--log-file", "/tmp/mergeway-lsp.log" },
    filetypes = { "yaml", "json" },
    root_dir = util.root_pattern("mergeway.yaml", "mergeway.yml"),
    single_file_support = false,
  },
}

lspconfig.mergeway_lsp.setup({})
```

If you keep the binary outside `PATH`, replace `mergeway-lsp` with the absolute path to the executable.

## Troubleshooting

- If the server does not start, run `mergeway-lsp --log-stderr --log-level=debug` directly in a terminal and inspect the error.
- If the editor connects but features are missing, confirm the opened folder contains `mergeway.yaml` or `mergeway.yml`.
- If diagnostics look wrong, compare the same workspace with `mergeway-cli validate`.
- If you need persistent logs, use `--log-file /tmp/mergeway-lsp.log`.
- If an editor integration writes logs to stdout, disable that behavior. LSP transport requires stdout to stay protocol-only.
