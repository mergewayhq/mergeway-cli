---
title: "Run mergeway-lsp manually"
linkTitle: "Run mergeway-lsp manually"
description: "Start the Mergeway language server, turn on logging, and understand its current capabilities."
weight: 18
---

`mergeway-lsp` is a Language Server Protocol server that communicates over stdio. In normal use, your editor starts it for you and keeps stdout reserved for protocol traffic.

## Install or Build the Binary

Use one of these paths:

- extract `mergeway-lsp` from a release archive
- run `make build` and use `./bin/mergeway-lsp`

The published Docker image does not include the language server.

## Start the Server Manually

If `mergeway-lsp` is on your `PATH`:

```bash
mergeway-lsp
```

From a local source checkout:

```bash
./bin/mergeway-lsp
```

The process will wait for framed LSP messages on stdin.

## Enable Debug Logging

Use stderr logging while testing locally:

```bash
mergeway-lsp --log-stderr --log-level=debug
```

Or write logs to a file:

```bash
mergeway-lsp --log-file /tmp/mergeway-lsp.log --log-level=debug
```

Equivalent environment variables:

- `MERGEWAY_LSP_LOG_STDERR=1`
- `MERGEWAY_LSP_LOG_FILE=/tmp/mergeway-lsp.log`
- `MERGEWAY_LSP_LOG_LEVEL=debug`

Keep stdout protocol-only. Human-readable logging on stdout will break editor integration.

## Current Capabilities

- workspace-aware diagnostics backed by the same validation core as `mergeway-cli validate`
- completion for fields, entity references, and close enum values
- hover, go-to-definition, find references, document symbols, and workspace symbols
- conservative quick fixes for a small set of unambiguous schema and data mistakes

## Current Limitations

- The VS Code extension still requires manual configuration of the local `mergeway-lsp` binary path.
- The server currently uses full-document sync.
- Missing-required-field quick fixes are limited to YAML and YML files.
- Features only apply to files owned by a detected Mergeway root.

For editor configuration examples, see [Set up mergeway-lsp in VS Code and Neovim](../guides/setup-mergeway-lsp-editors.md).
