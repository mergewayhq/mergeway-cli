---
title: "mergeway-mcp"
linkTitle: "mergeway-mcp"
description: "Run the Mergeway read-only MCP server over stdio or HTTP."
---

> **Synopsis:** Run the Mergeway read-only MCP server for repository inspection.

## Usage

```bash
mergeway-mcp [flags]
```

`mergeway-mcp` is a standalone binary. It does not run through `mergeway-cli`, and it does not expose any mutation workflow.

## Flags

| Flag               | Description                                                                          |
| ------------------ | ------------------------------------------------------------------------------------ |
| `--root`           | Root folder of the Mergeway repository (defaults to `.`).                            |
| `--transport`      | MCP transport: `stdio` or `http` (defaults to `stdio`).                             |
| `--http-listen`    | Listen address for HTTP transport (defaults to `127.0.0.1:8080`).                   |
| `--http-base-path` | Base path for HTTP transport (defaults to `/`). Only valid with `--transport=http`. |
| `--entity`         | Allow only the named exact entity. Repeat the flag to build an allow-list.          |

## Behavior

- The server is read-only. It exposes repository inspection tools and does not support create, update, or delete operations.
- `--entity` uses exact entity names. Allowing `Animal` does not implicitly expose `Dog`.
- Repository state reloads on each request, so changes made after server startup become visible on later calls.
- In `stdio` mode, keep stdout protocol-only. Any extra text written to stdout will break the MCP session.

## Examples

Start the default stdio server rooted at the current repository:

```bash
mergeway-mcp
```

Point the server at a different repository root:

```bash
mergeway-mcp --root examples/full
```

Expose only the `User` and `Post` entities to the client:

```bash
mergeway-mcp --root examples/full --entity User --entity Post
```

Serve the MCP endpoint over HTTP on the default local address:

```bash
mergeway-mcp --root examples/full --transport=http
```

Serve HTTP from a custom mount path:

```bash
mergeway-mcp --root examples/full --transport=http --http-listen 127.0.0.1:9090 --http-base-path /mcp
```

## Tool Surface

The initial tool set is intentionally narrow and read-only:

- `entity_list`
- `entity_show`
- `object_list`
- `object_get`
- `repository_export`
- `files_list`

Clients should treat these tool names and their structured responses as the supported inspection surface.

## Notes

- Use `stdio` for MCP clients that launch `mergeway-mcp` as a subprocess.
- Use `http` when the client expects a streamable HTTP endpoint.
- Bad flag combinations fail before the server starts. For example, `--http-listen` and `--http-base-path` require `--transport=http`.

## Related Pages

- [Install Mergeway](../getting-started/installation.md) — release, Go, and source installation paths.
- [Use mergeway-mcp with Claude and Codex](../guides/use-mergeway-mcp-with-claude-and-codex.md) — stdio client setup examples.
- [mergeway-cli Reference](README.md) — repository management and CRUD commands.
- [mergeway-lsp Reference](lsp.md) — compare the MCP server with the editor-facing LSP binary.
