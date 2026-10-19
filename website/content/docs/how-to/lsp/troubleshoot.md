---
title: Troubleshoot diagnostics and navigation
description: Diagnose LSP startup, root ownership, stale diagnostics, and empty navigation results.
type: docs
weight: 40
bookToC: true
---

# Troubleshoot diagnostics and navigation

## The server does not start

Run `mergeway-lsp --help` in the same environment as the editor and confirm the executable is found. Then enable a [debug log](../debug-logs/). If the log path cannot be opened or the level is invalid, startup fails before protocol initialization.

## The client attaches but reports nothing

Inspect the client's initialization data. Each `rootUri` or `workspaceFolders` entry must be the exact repository directory containing `mergeway.yaml` or `mergeway.yml`; the server does not walk upward. Dynamic `workspace/didChangeWorkspaceFolders` notifications are not supported, so restart the client after adding or removing folders.

The target must be the entry configuration, a sourced configuration file, or a file matched by an entity `include`. A YAML or JSON file merely located below the directory is not automatically owned. Unrelated files intentionally return empty completion, hover, navigation, symbol, and action results.

## Diagnostics are missing or stale

The client must send `didOpen` and full-text `didChange` notifications. Incremental range-only changes are not supported. Reloads are debounced for 75 ms. Compare saved state with:

```sh
mergeway-cli --root /path/to/repository validate
```

Saved repository diagnostics use the same validator and messages. Editor-only syntax errors and unsaved content can differ until saved. A temporarily incomplete YAML or JSON buffer can prevent a usable workspace snapshot; finish or undo the partial edit and wait for the next reload.

## Definition or references are empty

Definition works only on a reference value that resolves to exactly one record. References work on reference values and record identifiers, within the same detected root. Duplicate identifiers can make definition ambiguous. Files and records in another root are isolated even when the identifier text matches.

Document symbols apply to configured data files. Workspace symbols search record type, identifier, and primary display text; they do not index arbitrary YAML keys. The server does not watch unopened disk files, so saved external edits require an editor notification or server restart before the in-memory index changes.

## A quick fix is absent

Quick fixes are deliberately narrow: a close unknown key may be renamed to a missing required field, a missing required field may be inserted in YAML/YML, and close invalid enum or reference values may be replaced. The request must include the corresponding diagnostic and allow `quickfix`. No speculative fix is returned when there is no safe candidate.

For exact boundaries, use the [capability reference](../../../reference/lsp/#feature-behavior-and-limitations).
