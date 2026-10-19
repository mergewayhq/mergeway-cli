---
title: Connect Codex over stdio
description: Configure Codex to launch mergeway-mcp locally with an absolute repository root.
type: docs
weight: 10
bookToC: true
aliases:
  - /guides/use-mergeway-mcp-with-claude-and-codex.html
  - /guides/use-mergeway-mcp-with-claude-and-codex/
---

# Connect Codex over stdio

Install `mergeway-mcp`, validate the repository, and resolve both paths:

```sh
command -v mergeway-mcp
cd /path/to/repository
mergeway-cli validate
pwd
```

Add the local server. Replace both example paths with the absolute values above:

```sh
codex mcp add mergeway -- /absolute/path/to/mergeway-mcp \
  --root /absolute/path/to/repository
codex mcp list
```

Codex CLI and the Codex IDE extension share MCP configuration. The equivalent `~/.codex/config.toml` entry is:

```toml
[mcp_servers.mergeway]
command = "/absolute/path/to/mergeway-mcp"
args = ["--root", "/absolute/path/to/repository"]
```

Restart a running client so it reloads configuration. Ask it to list the Mergeway entities, then compare the answer with `mergeway-cli entity list`. If the server is missing, use `codex mcp --help`, `codex mcp list`, and the [troubleshooting guide](../troubleshoot/).

This syntax follows the current [official Codex MCP configuration](https://learn.chatgpt.com/docs/extend/mcp). Client commands can evolve independently of the Mergeway protocol and tool contract.
