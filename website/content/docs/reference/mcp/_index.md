---
title: mergeway-mcp reference
description: Invocation, lifecycle, compatibility, transport, error, and tool reference for the read-only Mergeway MCP server.
type: docs
weight: 40
bookToC: true
aliases:
  - /cli-reference/mcp.html
  - /cli-reference/mcp/
params:
    bookCollapseSection: true
---

# `mergeway-mcp` reference

`mergeway-mcp` v0.10.0 exposes a repository declaring `mergeway.version: 1` through a read-only Model Context Protocol server. It advertises tools only: no MCP resources, prompts, mutation tools, or subscriptions.

## Invocation

```sh
mergeway-mcp [flags]
```

| Flag | Default | Effect |
| --- | --- | --- |
| `--root <path>` | `.` | Repository root. The server makes it absolute and looks directly inside it for `mergeway.yaml`, then `mergeway.yml`; it does not search parents. |
| `--transport stdio\|http` | `stdio` | Select newline-delimited MCP over standard input/output or stateless streamable HTTP. Values are case-insensitive. |
| `--http-listen <address>` | `127.0.0.1:8080` | TCP listen address for HTTP. Supplying this flag is invalid unless `--transport=http`. |
| `--http-base-path <path>` | `/` | Exact HTTP mount path. It must start with `/`; supplying it is invalid unless `--transport=http`. |
| `--entity <name>` | all entities | Expose only this exact entity. Repeat for several entities. Names are case-sensitive; duplicates are removed. |
| `-h`, `--help` | — | Write help to standard error and exit `0`. |

The root must be an existing directory with a loadable configuration. Every `--entity` must name an entity present at startup. Unexpected positional arguments, unknown flags, unsupported transports, malformed addresses and base paths, and HTTP-only flags in stdio mode are rejected before the server starts.

## Lifecycle and reload behavior

Configuration is loaded once to validate startup and the allow-list. Each tool request then reloads the workspace, configuration, and records from disk. A later request sees saved changes without restarting the server. If the changed configuration no longer loads or removes an allow-listed entity, the request fails; the startup allow-list itself does not change.

The server loads repository data but does not run the complete validation pipeline before every result. Validate publishable state separately with [`mergeway-cli validate`](../cli/validate/).

## Output, errors, and exit codes

In stdio mode, standard output contains MCP protocol messages only. Help, startup diagnostics, and runtime failures go to standard error. Wrappers must also keep stdout protocol-only.

| Exit | Meaning |
| --- | --- |
| `0` | Help completed, the client closed a stdio session cleanly, or HTTP shutdown completed after cancellation. |
| `1` | Repository/server initialization or runtime failed. |
| `2` | Command-line parsing or validation failed. |

Tool argument decoding failures are MCP tool-error results. Known lookup failures use JSON-RPC invalid-parameters errors with `data.kind` set to `unknown_entity`, `entity_not_allowed`, or `object_not_found`. Repository loading failures use an internal error with `data.kind: repository_error`. Messages can include repository paths and must not be treated as a stable machine API; use the kind where present.

## Compatibility

The v0.10.0 server uses `github.com/modelcontextprotocol/go-sdk` v1.6.1. It negotiates MCP `2025-11-25`, `2025-06-18`, `2025-03-26`, or `2024-11-05`; its current default negotiation is `2025-11-25`. Clients must support MCP tools plus either stdio or streamable HTTP. HTTP is stateless JSON-response mode; the legacy standalone SSE transport is not exposed. SDK compatibility does not imply that Mergeway implements every optional MCP feature.

Continue with the [transport contract](transports/), [complete tool contract](tools/), or [access and security model](../../explanation/mcp-access-security/).
