---
title: Editor architecture
description: Understand how LSP roots, owned files, overlays, validation, and editor-specific layers fit together.
type: docs
weight: 60
bookToC: true
---

# Editor architecture

Mergeway separates repository intelligence from editor presentation. `mergeway-lsp` is the editor-neutral process: an LSP client launches it over stdio, supplies workspace folders and full document text, and renders the returned diagnostics and language features. An editor extension may add commands or a visual editor, but those are separate consumers of the same repository model.

```text
editor UI and client
        │ stdio LSP
        ▼
mergeway-lsp
        │ roots + unsaved overlays
        ▼
configuration, data loader, validator, record index
        │
        ▼
repository files on disk
```

## Roots establish isolation

At initialization the server takes `workspaceFolders`, or falls back to `rootUri` and then legacy `rootPath`. Each candidate is checked directly for `mergeway.yaml` and then `mergeway.yml`. A valid candidate gets its own configuration, records, and identifier index. This isolation means equal identifiers in two roots neither collide nor navigate across roots.

The initial folder list is static: the server advertises workspace-folder support but not change notifications. A client must restart or create another server after its folder set changes.

## Configuration establishes ownership

A root owns its entry configuration, sourced configuration files, and the concrete files matched by entity includes. Location alone is insufficient. Ownership keeps unrelated YAML from receiving Mergeway semantics and lets external include paths participate even when they sit outside the root directory.

## Unsaved text is an overlay

Open and changed documents replace the corresponding disk content in an in-memory overlay. After a 75 ms debounce, each affected repository is reloaded and validated as a whole. Completion and navigation therefore see related unsaved records, not only the current line. Closing a document removes its overlay and restores disk content.

This favors consistency over incremental parsing. The client sends full documents; very large repositories are reloaded and validated after edits. The implementation establishes no formal repository-size, latency, or memory guarantee, so teams should measure their own largest workspace rather than rely on a speculative limit.

## Validation is shared, presentation differs

For saved, parseable repositories, LSP diagnostics are produced from the same validation package and messages as `mergeway-cli validate`: format, schema, reference, and configuration failures become error-severity diagnostics with source `mergeway`. The editor maps errors to document ranges and adds buffer versions. Unsaved syntax failures and temporary partial documents exist only in the editor view; final automation should still run the CLI.

Language features then use the same configuration and record index. That shared core is why completion can propose fields and identifiers, hover can describe schemas and records, and navigation can resolve relationships without an editor-specific database.

Continue with the [`mergeway-lsp` reference](../../reference/lsp/) or a [client setup guide](../../how-to/lsp/).
