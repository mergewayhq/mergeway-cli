---
title: Explanation
description: Understand Mergeway's concepts, design choices, and component relationships.
type: docs
weight: 50
params:
  bookFlatSection: true
  bookIcon: owl
---

# Explanation

Build a mental model before choosing commands or designing a repository:

- [Why metadata as code?](metadata-as-code/) connects file-backed records to normal review and automation practices.
- [The Mergeway repository model](repository-model/) explains repositories, workspaces, entities, and records.
- [The validation model](validation-model/) explains the checks shared by Mergeway tools.
- [Semantic diff versus file diff](semantic-diff/) explains logical records, snapshots, and why line changes are not always data changes.
- [MCP access and security](mcp-access-security/) explains read-only tools, entity scoping, and the surrounding trust boundaries.
- [Editor architecture](editor-architecture/) explains the LSP process, workspace ownership, in-memory overlays, and validation feedback loop.
- [Schema design choices](schema-design/) compares schema formats, record granularity, inheritance, unions, and identity strategies.
- [Component architecture](component-architecture/) traces the shared repository core through each runnable interface.

The [component overview](../components/) shows how the released tools relate to the shared specification. A scaling guide is deferred until measured operational evidence can support useful thresholds and trade-offs; the site does not invent capacity guidance from implementation structure alone.
