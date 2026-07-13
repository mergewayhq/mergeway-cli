---
title: "Use mergeway-mcp with Claude and Codex"
linkTitle: "Use mergeway-mcp with Claude and Codex"
description: "Configure mergeway-mcp as a local stdio MCP server for Claude and Codex."
weight: 36
---

This guide shows how to launch `mergeway-mcp` over stdio from a local MCP client.

## Before You Start

- Put `mergeway-mcp` on your `PATH`, or use an absolute path to the binary.
- Decide which Mergeway repository root the server should expose.
- Prefer an absolute path for `--root` so the client does not depend on its own working directory.
- Keep stdout reserved for MCP protocol traffic. Do not wrap `mergeway-mcp` in a script that prints extra output to stdout.

For example, if the repository you want to inspect lives at `/abs/path/to/repo`, this guide uses:

```bash
mergeway-mcp --root /abs/path/to/repo
```

Add repeated `--entity` flags if the client should only see specific exact entity names:

```bash
mergeway-mcp --root /abs/path/to/repo --entity User --entity Post
```

## Claude

If you use Claude Code, add the server from the CLI:

```bash
claude mcp add --transport stdio mergeway -- mergeway-mcp --root /abs/path/to/repo
```

Run `claude mcp list` or open `/mcp` inside Claude Code to confirm that the server is connected.

If you use Claude Desktop, add a stdio server entry to your `claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "mergeway": {
      "type": "stdio",
      "command": "mergeway-mcp",
      "args": [
        "--root",
        "/abs/path/to/repo"
      ],
      "env": {}
    }
  }
}
```

If `mergeway-mcp` is not on `PATH`, replace `command` with the full executable path.

After saving the file, restart Claude Desktop so it reloads the MCP configuration.

## Codex

You can add the server from the CLI:

```bash
codex mcp add mergeway -- mergeway-mcp --root /abs/path/to/repo
```

To keep the configuration in `~/.codex/config.toml` or a trusted project-level `.codex/config.toml`, use:

```toml
[mcp_servers.mergeway]
command = "mergeway-mcp"
args = ["--root", "/abs/path/to/repo"]
```

If you want to narrow the exposed surface, include entity filters in `args`:

```toml
[mcp_servers.mergeway]
command = "mergeway-mcp"
args = ["--root", "/abs/path/to/repo", "--entity", "User", "--entity", "Post"]
```

Run `codex mcp list` to confirm that Codex sees the server.

## Example Prompts

Once the client connects, ask for repository inspection tasks such as:

- `List the Mergeway entities in this repository.`
- `Show the schema for the User entity.`
- `List the User objects and summarize the identifiers.`
- `Export the repository structure and explain the main entity relationships.`

## Related Pages

- [mergeway-mcp Reference](../cli-reference/mcp.md) — flags, transport options, and tool surface.
- [Install Mergeway](../getting-started/installation.md) — installation paths for `mergeway-mcp`.
- [Claude Code MCP docs](https://code.claude.com/docs/en/mcp) — Anthropic MCP client documentation.
- [Codex manual: Model Context Protocol](https://developers.openai.com/codex/codex-manual) — official Codex MCP configuration docs.
