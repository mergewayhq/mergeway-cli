---
title: Component architecture
description: Understand how the specification and shared Go core support the CLI, Diff, MCP, LSP, and editor integrations.
type: docs
mode: explanation
weight: 80
bookToC: true
---

# Component architecture

Mergeway's components are separate interfaces over one file-backed repository model. The specification defines the contract; shared Go packages load and validate it; each runnable component adds a workflow-specific boundary.

```text
repository files
      │
      ▼
configuration + data + validation core
      │
      ├── mergeway-cli  operations, validation, formatting, export
      ├── mergeway-diff snapshots and semantic comparison
      ├── mergeway-mcp  read-only MCP tools
      └── mergeway-lsp  diagnostics and language features
                              ▲
                              │ stdio LSP
                       editor integrations
```

## Shared repository core

`internal/config` loads and normalizes configuration. `internal/data` discovers sources and represents records. `internal/validation` applies configuration, format, schema, and reference checks. This layering keeps identity, inheritance, path derivation, and reference semantics consistent across interfaces.

## Workflow boundaries

The [CLI](../../components/cli/) is the explicit mutation and automation interface. [Diff](../../components/diff/) builds logical snapshots rather than comparing lines. [MCP](../../components/mcp/) exposes a deliberately read-only tool surface. [LSP](../../components/lsp/) maintains workspace indexes and unsaved overlays, then maps shared validation into editor diagnostics.

The checked-in VS Code extension is an LSP client and does not redefine repository behavior. The separate visual editor assumed by the migration plan is not present in this repository, so no visual-editor architecture is claimed here. [Editor architecture](../editor-architecture/) covers the shipped LSP boundary.

## Compatibility follows the repository contract

Repository compatibility is anchored by `mergeway.version: 1`. Executable compatibility still matters: use suite binaries from the same release because their command, protocol, and output surfaces can evolve independently. The [component overview](../../components/) records current distribution boundaries, while the [specification](../../reference/specification/) defines portable repository behavior.

Contributors can find package ownership and build procedures in [Contributing](../../project/contributing/).
