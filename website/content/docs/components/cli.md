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

The CLI creates, reads, updates, and deletes records; inspects entities and files; formats and validates repositories; and exports data. It is not an interactive editor, semantic Git diff engine, MCP server, language server, or Python API; the other components cover those roles.

## Relationship to the suite

The CLI is the most direct interface to the specification and shares configuration, loading, and validation code with other Go components. The Python SDK invokes it and consumes its JSON output. Validate a repository before handing it to Diff, MCP, LSP, or downstream automation.

## Availability and compatibility

`mergeway-cli` is released for Linux, macOS, and Windows on `amd64` and `arm64`. A release archive provides it alongside the matching Diff, MCP, and LSP binaries. Docker and Nix provide only the CLI, while Go and source builds can install binaries individually. Version 0.10.0 supports repositories declaring `mergeway.version: 1`.

## Start here

- [Install Mergeway tools](../../how-to/install/).
- [Build your first Mergeway repository](../../tutorials/first-repository/).
- Look up every released operation in the [`mergeway-cli` reference](../../reference/cli/).
- Follow the [record lifecycle](../../how-to/cli/manage-records/), choose a [validation scope](../../how-to/cli/choose-validation/), or configure [repository automation](../../how-to/cli/github-actions/).
- Understand [the repository model](../../explanation/repository-model/) and [validation model](../../explanation/validation-model/).
- [Troubleshoot root, configuration, validation, mutation, or output failures](../../how-to/cli/troubleshoot/).
- Use the [Python SDK](../python/) when application code needs typed models around these operations.

The reference covers all 16 released repository operations, their flags and defaults, output and exit behavior, and file side effects. The how-to section adds safe mutation, layout, validation, export, ERD, pre-commit, and CI workflows.
