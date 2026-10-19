---
title: Restrict MCP to selected entities
description: Apply and verify an exact-name entity allow-list for a Mergeway MCP client.
type: docs
weight: 20
bookToC: true
---

# Restrict MCP to selected entities

Choose the minimum exact entity set the task needs. For example, expose users and posts but not comments or tags:

```sh
mergeway-mcp --root /absolute/path/to/repository \
  --entity User --entity Post
```

Put the same repeated arguments in a stdio client's configuration:

```toml
[mcp_servers.mergeway]
command = "/absolute/path/to/mergeway-mcp"
args = [
  "--root", "/absolute/path/to/repository",
  "--entity", "User",
  "--entity", "Post",
]
```

Restart the client. `entity_list` should return only `Post` and `User`; `entity_show` for `Comment` should fail with `entity_not_allowed`. Names are case-sensitive. Inheritance does not widen the rule: allowing a parent does not allow its descendants.

The allow-list filters all six tool results, including export and backing paths. It does not redact fields inside allowed records, hide referenced identifiers or schema relationships, limit the server process's filesystem permissions, authenticate an HTTP client, validate prompt content, or prevent the AI client from using unrelated capabilities. See the full [security boundary](../../../explanation/mcp-access-security/).
