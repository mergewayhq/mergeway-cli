---
title: Inspect a repository with an AI client
description: Connect Codex to a scoped read-only Mergeway server and inspect the fixed examples/full repository.
type: docs
weight: 30
bookToC: true
---

# Inspect a repository with an AI client

In this tutorial, Codex launches `mergeway-mcp` over stdio and inspects the fixed `examples/full` repository. The server cannot modify records, and the allow-list exposes only `User` and `Post`.

## Prerequisites

- Codex CLI with local stdio MCP support.
- `mergeway-cli` and `mergeway-mcp` v0.10.0 on `PATH`.
- A v0.10.0 source archive containing `examples/full`.

## 1. Validate and locate the repository

From the extracted source archive:

```sh
mergeway-cli --root examples/full validate
cd examples/full
pwd
command -v mergeway-mcp
```

Expect `status: validation succeeded`. Save the two absolute paths printed by `pwd` and `command -v`.

## 2. Add the scoped server

Replace the paths below:

```sh
codex mcp add mergeway-tutorial -- /absolute/path/to/mergeway-mcp \
  --root /absolute/path/to/examples/full \
  --entity User \
  --entity Post
codex mcp list
```

Restart Codex if a running session does not show the new server.

## 3. Discover the visible model

Ask:

> Use the Mergeway tools to list the visible entities. Do not use shell commands.

The answer should identify only `Post` and `User`. `Comment` and `Tag` exist in the repository but are hidden by the allow-list.

## 4. Inspect records and relationships

Ask:

> List the User records, get user-alice, then inspect Post post-001. Explain the author relationship using only tool results.

The tool results establish these facts:

- users `user-alice` and `user-bob` exist;
- Alice's name is `Alice Example` and her roles are `admin` and `author`;
- post `post-001` is titled `Launch Day` and its `author` is `user-alice`.

The post also contains `tag-product`, but `Tag` is not visible. The allow-list does not redact reference values inside an allowed record.

## 5. Test the boundary

Ask:

> Show the Comment entity schema with the Mergeway tools.

The call should fail because `Comment` is not allowed. This verifies scope, not isolation from every Codex capability: the client may have other tools, and the server process still has its operating-system permissions.

## 6. Remove the tutorial connection

Use `codex mcp --help` to find the removal command supported by your installed client, then remove `mergeway-tutorial` if you no longer need it. Client configuration syntax is maintained separately; the stable Mergeway behavior is in the [server](../../reference/mcp/) and [tool](../../reference/mcp/tools/) references.

Continue with [MCP access and security](../../explanation/mcp-access-security/) before connecting a real repository.
