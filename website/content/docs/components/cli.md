---
title: CLI
description: Operate on, validate, format, inspect, and export a Mergeway repository.
type: docs
weight: 20
bookToC: true
---

# CLI

`mergeway-cli` is the general-purpose command-line interface for creating and inspecting records, validating and formatting repositories, and exporting repository data. It implements the Mergeway specification and supplies the validation foundation used by other components.

## Who it is for

Repository maintainers, data and platform engineers, schema authors, and automation use the CLI for direct repository operations.

## Purpose and boundaries

The CLI creates, reads, updates, and deletes records; inspects entities and files; formats and validates repositories; and exports data. It is not an interactive editor, semantic Git diff engine, MCP server, or language server; the other binaries cover those roles.

## Relationship to the suite

The CLI is the most direct interface to the specification and shares configuration, loading, and validation code with other components. Validate a repository before handing it to Diff, MCP, LSP, or downstream automation.

## Availability and compatibility

`mergeway-cli` is released for Linux, macOS, and Windows on `amd64` and `arm64`. A release archive provides it alongside the matching Diff, MCP, and LSP binaries. Docker and Nix provide only the CLI, while Go and source builds can install binaries individually. Version 0.10.0 supports repositories declaring `mergeway.version: 1`.

## Start here

- [Install Mergeway tools](../../how-to/install/).
- [Build your first Mergeway repository](../../tutorials/first-repository/).
- Understand [the repository model](../../explanation/repository-model/) and [validation model](../../explanation/validation-model/).

A complete command, flag, output, and exit-code reference will be migrated in the CLI documentation phase.
