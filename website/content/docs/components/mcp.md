---
title: MCP
description: Give MCP clients read-only, scoped access to a Mergeway repository.
type: docs
weight: 40
bookToC: true
---

# MCP

`mergeway-mcp` is a Model Context Protocol server for read-only inspection of a Mergeway repository. It lets compatible clients discover entities and records while using the suite's shared loading and validation behavior.

## Who it is for

AI-tool users and client integrators use MCP when a compatible client needs structured, controlled access to repository data.

## Purpose and boundaries

The server exposes read-only discovery and inspection through MCP over stdio or HTTP. Entity allow-listing can narrow the visible model. It does not create, update, or delete records, and read-only tools do not by themselves make an untrusted transport or client safe.

## Relationship to the suite

MCP loads the specification's entities and records using the shared repository code. Use the CLI for mutation and explicit validation; use LSP for editor features.

## Availability and compatibility

`mergeway-mcp` v0.10.0 is included in release archives for Linux, macOS, and Windows on `amd64` and `arm64`; it can also be installed with Go or built from source. Docker and the Nix flake do not provide it. The released Go SDK negotiates MCP `2025-11-25`, `2025-06-18`, `2025-03-26`, and `2024-11-05` over stdio or streamable HTTP.

## Start here

- [Install Mergeway tools](../../how-to/install/).
- Follow [Inspect a repository with an AI client](../../tutorials/inspect-with-ai-client/).
- Use the [MCP guides](../../how-to/mcp/) to connect, scope, serve, and troubleshoot the server.
- Look up the [server and tool contract](../../reference/mcp/) or understand [MCP access and security](../../explanation/mcp-access-security/).
