---
title: MCP access and security
description: Understand the read-only promise, entity allow-list, process and network boundaries, and AI-content risks.
type: docs
weight: 50
bookToC: true
---

# MCP access and security

`mergeway-mcp` separates mutation authority from inspection. Its catalog contains six query tools and no create, update, delete, shell, or arbitrary file-read tool. Calls reload saved repository state but never persist changes.

## What read-only guarantees

Through this server's tools, a client can discover schemas, record fields, logical identifiers, configured backing paths, and structured exports. It cannot ask this server to modify them. This is narrower than saying the entire AI client is read-only: the client may have shell, editor, network, or other MCP tools with separate permissions.

Repository content is untrusted input. A record can contain text that looks like an instruction. Clients should treat returned fields as data, not authority, and should not let data override system or user intent.

## What the entity allow-list guarantees

With no `--entity`, every configured entity is visible. With one or more flags, each tool permits only those exact names. Listing and whole-repository operations filter to that set; direct calls for another existing entity fail. Inheritance does not implicitly expose descendants.

The allow-list does not:

- redact individual fields or values within an allowed entity;
- remove identifiers that reference a blocked entity, or necessarily hide its name in schema relationships;
- constrain which files the operating-system process can read;
- turn an external include into a sandboxed path;
- authenticate clients, encrypt HTTP, or distinguish users;
- validate repository correctness or sanitize prompt-like content;
- restrict capabilities supplied by the client, another server, or the host application.

Use a purpose-built repository or data projection when field-level secrecy is required. Run the server as a minimally privileged OS identity, expose the smallest entity set, validate first, and prefer a local stdio child process.

## Transport trust

Stdio ties the server lifecycle to one launching client and avoids a listening socket. HTTP supports separately managed clients but ships without authentication or TLS. Its loopback default reduces accidental network exposure; it is not authorization. A remote deployment needs a trusted authenticated TLS proxy, network controls, lifecycle supervision, and its own audit policy.

The result is a useful least-authority layer, not a complete security boundary. Combine the [restriction procedure](../../how-to/mcp/restrict-entities/) with the host client's permissions and deployment controls.
