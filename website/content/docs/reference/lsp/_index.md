---
title: mergeway-lsp reference
description: Invocation, logging, initialization, workspaces, diagnostics, capabilities, and limitations of the Mergeway language server.
type: docs
weight: 50
bookToC: true
aliases:
  - /cli-reference/lsp.html
  - /cli-reference/lsp/
  - /getting-started/language-server.html
  - /getting-started/language-server/
  - /guides/setup-mergeway-lsp-editors.html
  - /guides/setup-mergeway-lsp-editors/
---

# `mergeway-lsp` reference

`mergeway-lsp` v0.10.0 is an editor-neutral Language Server Protocol process for repositories declaring `mergeway.version: 1`. It communicates only over standard input/output using JSON-RPC messages with LSP `Content-Length` framing.

## Invocation and exit behavior

```sh
mergeway-lsp [flags]
```

| Flag | Default | Effect |
| --- | --- | --- |
| `--log-file <path>` | unset | Append structured text logs to a file. A selected file takes priority over stderr. |
| `--log-level <level>` | `info` | Minimum `debug`, `info`, `warn`, or `error` level. |
| `--log-stderr` | false | Write logs to stderr when no log file is selected. |
| `-h`, `--help` | — | Write flag help to stderr and exit `2`. |

There is no `--root` flag: the LSP client supplies roots during initialization. Help, unknown flags, and malformed command lines exit `2` because the released command uses Go's `flag` error path for help. An invalid log level, an unusable log file, or a server/transport failure exits `1`. A `shutdown` request followed by `exit` exits `0`; `exit` without prior shutdown exits `1`.

Stdout is reserved for protocol frames. Help, command errors, and optional logs go to stderr; wrappers must not print banners or combine stderr with stdout.

## Environment and precedence

| Variable | Meaning |
| --- | --- |
| `MERGEWAY_LSP_LOG_FILE` | File path used when `--log-file` is empty. |
| `MERGEWAY_LSP_LOG_LEVEL` | Level used when `--log-level` is empty. |
| `MERGEWAY_LSP_LOG_STDERR` | Enables stderr for `1`, `true`, `yes`, or `on`, ignoring case and whitespace. |

Non-empty file and level flags override their variables. `--log-stderr` and the truthy stderr variable are additive rather than a true/false override. After resolving both, a non-empty log file wins over stderr. With no destination, logs are discarded.

## Initialization and roots

The server accepts the standard `initialize` fields below:

1. A non-empty `workspaceFolders` array is authoritative. Every non-empty file URI becomes a separate root candidate.
2. Otherwise `rootUri` supplies one candidate.
3. Otherwise legacy `rootPath` supplies one candidate.
4. With none of these, the server starts with no Mergeway roots.

Each candidate may be a directory or file path. A file candidate is reduced to its parent. The resulting directory is checked directly for `mergeway.yaml`, then `mergeway.yml`; parent directories are not searched. Missing markers do not fail initialization, but that candidate has no features.

Multi-root indexes are isolated. The server advertises `workspace.workspaceFolders.supported: true` and `changeNotifications: false`: initialization-time folders work, but `workspace/didChangeWorkspaceFolders` is not registered. Restart when the folder set changes.

## File ownership and synchronization

A detected root owns:

- its `mergeway.yaml` or `mergeway.yml` entry file;
- every sourced configuration file resolved by that configuration;
- concrete data files matched by an entity include, including absolute/external paths.

An arbitrary YAML or JSON file under the directory is not owned. Feature requests for unrelated files return empty results. Supported repository content is YAML/YML and JSON as admitted by the configured data sources; the server's semantic analysis uses YAML-compatible document nodes. Missing-field insertion fixes are YAML/YML-only.

The server advertises open/close tracking and `textDocumentSync.change: 1` (`Full`). `didOpen` stores the complete buffer. For `didChange`, only the last supplied content change is used as the complete new text; incremental range patches are unsupported. Reload and diagnostics are debounced by 75 ms. `didClose` discards the overlay and restores disk content.

## Advertised capabilities

The initialization result names the server `mergeway-lsp`, reports its build version, and advertises:

| LSP capability | Advertised value |
| --- | --- |
| `textDocumentSync` | `{ openClose: true, change: 1 }` |
| `completionProvider` | `{}` |
| `hoverProvider` | `true` |
| `definitionProvider` | `true` |
| `referencesProvider` | `true` |
| `documentSymbolProvider` | `true` |
| `workspaceSymbolProvider` | `true` |
| `codeActionProvider` | `true` |
| `workspace.workspaceFolders` | `{ supported: true, changeNotifications: false }` |

The implementation uses `go.lsp.dev/protocol` v0.12.0 and accepts the implemented lifecycle, synchronization, trace, and feature methods. That library version is not a promise of complete support for a particular LSP revision. Clients should use capability negotiation and standard `file://` URIs.

## Feature behavior and limitations

### Diagnostics

Initialization and each overlay reload validate all detected roots and publish errors with source `mergeway`. Diagnostic codes are the validation phases `format`, `schema`, and `references`; configuration load failures use `config`. Severity is Error. The server maps messages to the narrowest known line, field, identifier, object, or configuration range and clears previously published paths when errors disappear.

For a saved, parseable repository, message/category results have parity with [`mergeway-cli validate`](../cli/validate/), because both call the same validator. LSP adds ranges, source, severity, and open-document versions. Unsaved syntax errors and overlays can differ from the saved CLI view.

### Completion

Configuration `type:` values complete the primitive types `string`, `integer`, `number`, `boolean`, `enum`, and `object`, plus configured entity names. Data records complete absent fields, enum values, and identifiers from referenced entity types. Results are prefix-filtered and repository-root-local. There are no trigger characters or resolve requests.

### Hover

Hover describes configuration field types, requirements, descriptions, and allowed enum/reference values. On reference values it summarizes the resolved record; on identifiers it summarizes the current record. Unknown, ambiguous, malformed, or unowned positions return no hover.

### Go to definition

Definition resolves a reference value to the identifier range of exactly one matching record in the same root. Missing or duplicate targets return no location. Unsaved overlays participate after reload.

### Find references

References operate from record identifiers and reference values. Results include matching reference fields in the same root; the declaration is included only when the client sets `includeDeclaration`. Inherited/descendant types participate when compatible with the declared reference type. Results are sorted and deduplicated.

### Document and workspace symbols

Document symbols describe each record and its configured fields in an owned data file; configuration files and unrelated files return none. Workspace symbols search record entity name, identifier, and primary display text case-insensitively across all initialized roots. Roots remain isolated for semantic lookup even though one response can contain symbols from several roots.

### Code actions and quick fixes

Only `quickfix` actions are returned. The server can rename a close unknown key to a missing required field, insert a missing required YAML/YML field with a placeholder, replace a close invalid enum value, or replace a close missing reference with an existing identifier. It derives actions from diagnostics supplied in the request and omits a fix when no conservative candidate exists. It does not format whole files or apply edits itself.

## Scaling and reload boundaries

Each edit reloads and validates each detected repository after the debounce; analysis is not incremental. Open buffers are held in memory, and workspace-symbol searches walk loaded records. No formal size, latency, or memory ceiling is established by the implementation or tests. Saved external filesystem changes are not watched independently; clients should notify through document lifecycle events or restart the server.

See [editor architecture](../../explanation/editor-architecture/) for the data flow, [configure a generic client](../../how-to/lsp/generic-client/) for integration, and [troubleshooting](../../how-to/lsp/troubleshoot/) for empty results.
