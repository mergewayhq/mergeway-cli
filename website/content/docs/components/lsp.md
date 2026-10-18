---
title: LSP
description: Add Mergeway diagnostics and repository-aware navigation to compatible editors.
type: docs
weight: 50
bookToC: true
---

# LSP

`mergeway-lsp` is a Language Server Protocol server that gives editors repository-aware diagnostics and navigation. It reuses Mergeway's configuration, data, and validation behavior so editor feedback aligns with command-line validation.

## Availability

The language server is released in GitHub release archives and can be built from this repository. It communicates over stdio and must be started by an editor or LSP client with stdout reserved for protocol traffic.

## Planned documentation

Later phases will add editor configuration and troubleshooting guides, supported-capability reference, logging details, and an explanation of workspace and document ownership.
