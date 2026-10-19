---
title: LSP
description: Add Mergeway diagnostics and repository-aware navigation to compatible editors.
type: docs
weight: 50
bookToC: true
---

# LSP

`mergeway-lsp` is a Language Server Protocol server that gives editors repository-aware diagnostics and navigation. It reuses Mergeway's configuration, data, and validation behavior so editor feedback aligns with command-line validation.

## Who it is for

Repository authors who want feedback while editing and editor/plugin authors who need a standards-based integration use LSP.

## Purpose and boundaries

The server provides diagnostics and repository-aware navigation over stdio. It is not a standalone terminal UI, does not replace final CLI validation, and does not install or configure an editor integration by itself. Standard output is reserved for protocol traffic.

## Relationship to the suite

LSP discovers workspaces and reuses Mergeway configuration, data, and validation behavior. The VS Code extension launches this server; other compatible editors can integrate it directly.

## Availability and compatibility

`mergeway-lsp` is included in release archives for Linux, macOS, and Windows on `amd64` and `arm64`; it can also be installed with Go or built from source. Docker and the Nix flake do not provide it. The server is distributed separately from the VS Code extension.

## Start here

- [Install Mergeway tools](../../how-to/install/) to obtain the server.
- [Configure Neovim](../../how-to/lsp/neovim/) or [another LSP client](../../how-to/lsp/generic-client/).
- Look up the complete [`mergeway-lsp` server contract](../../reference/lsp/).
- Understand [how the editor integration works](../../explanation/editor-architecture/), including roots, in-memory overlays, and validation parity.
- [Capture debug logs](../../how-to/lsp/debug-logs/) or [troubleshoot missing features](../../how-to/lsp/troubleshoot/).

The released server supports diagnostics, completion, hover, definitions, references, document and workspace symbols, and conservative quick fixes. The separate [VS Code component](../vscode/) owns VS Code extension and visual-editor guidance.
