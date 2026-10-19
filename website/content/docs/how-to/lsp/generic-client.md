---
title: Configure a generic LSP client
description: Map a client launcher to mergeway-lsp stdio, file types, and root initialization.
type: docs
weight: 20
bookToC: true
---

# Configure a generic LSP client

Use these settings in any client that can launch a Language Server Protocol process. Setting names differ by client; the values are the portable contract.

| Client setting | Value |
| --- | --- |
| Command | `mergeway-lsp` |
| Arguments | none, or logging flags only |
| Transport | stdio with standard LSP `Content-Length` framing |
| File types | YAML and JSON files used by the repository |
| Root markers | `mergeway.yaml`, `mergeway.yml` |
| Initialization root | the directory containing the root marker |
| Synchronization | full document, with open and close notifications |

Configure the client to send a `file://` URI in `rootUri`, or send one or more file-URI entries in `workspaceFolders`. When `workspaceFolders` is non-empty it is authoritative and `rootUri`/legacy `rootPath` are ignored. The server checks each supplied directory directly; the client must find the marker rather than expecting the server to search parent directories.

Do not configure a TCP port, socket, or HTTP endpoint. Do not merge the server's standard error into standard output: stdout carries protocol frames only.

## Verify the connection

Open a configured record and confirm the initialization result names `mergeway-lsp`. The client should see these providers:

```text
completion, hover, definition, references,
document symbols, workspace symbols, code actions
```

Change a required string field to a number without saving. A working client publishes an error with source `mergeway` and code `schema`; undoing the edit clears it. The repository's saved errors are also published during initialization.

The server accepts normal LSP lifecycle and feature messages but does not claim every optional method or a particular LSP specification revision. Integrate against the [advertised capabilities](../../../reference/lsp/#advertised-capabilities), not the dependency version.
