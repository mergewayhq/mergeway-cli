---
title: MCP
description: Give MCP clients read-only, scoped access to a Mergeway repository.
type: docs
weight: 40
bookToC: true
---

# MCP

`mergeway-mcp` is a Model Context Protocol server for read-only inspection of a Mergeway repository. It lets compatible clients discover entities and records while using the suite's shared loading and validation behavior.

## Availability

The MCP server is released in GitHub release archives and can be built from this repository. It supports local stdio use and an HTTP transport; access can be restricted to selected entities. It does not provide record mutation tools.

## Planned documentation

Later phases will add client setup guides, transport and security guidance, troubleshooting, and a complete reference for flags, resources, and tools.
