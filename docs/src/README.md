---
title: "Mergeway CLI"
linkTitle: "CLI"
description: "Overview of Mergeway CLI, what it does, and why it keeps schemas and data consistent."
cascade:
  type: docs
---

# Mergeway Overview

Mergeway is a lightweight CLI that keeps metadata honest by treating schemas as code. Instead of juggling spreadsheets or custom scripts, you describe entities in YAML/JSON, run a quick validation, and catch broken references before they reach production.

## What the CLI Does

- Stores entity definitions and relationships in version-controlled files.
- Validates schemas and records so required fields and references stay consistent.
- Generates simple reports you can attach to pull requests or issues.

## Key Features

- **Workspace scaffolding**: `mw init` writes a starter `mergeway.yaml` into your working directory so you can begin defining entities immediately.
- **Dual schema sources**: Author entity fields inline in YAML or reference existing JSON Schema documents (`json_schema`) so teams can reuse specs.
- **Object lifecycle commands**: `list`, `get`, `create`, `update`, and `delete` operate on local YAML/JSON files, respecting identifier fields defined in schemas and inline data.
- **Deterministic formatting**: `mw fmt` emits canonical structure and rewrites files in place (use `--stdout` to preview changes) to keep diffs clean.
- **Layered validation**: Format, schema, and reference phases catch structural, typing, and cross-entity errors before they land in main.
- **Schema introspection**: `mw entity show` and `mw config export` surface normalized schemas or derived JSON Schema for documentation and automation.

## Why Teams Use Mergeway

- **Fast feedback:** One command surfaces missing fields, enum mismatches, or invalid references.
- **Git-native:** Changes live in branches and pull requests, making reviews trivial.
- **Lightweight:** No server component—just a binary that runs locally or in CI.

## Where to Go Next

1. [Install Mergeway](getting-started/installation.md) (or build from source).
2. Follow the [Workspace set-up](getting-started/workspace-setup.md).
3. Review the [Basic Concepts](getting-started/README.md) and [Schema Format](getting-started/schema-spec.md) when you define entities.
4. Browse through the [CLI Reference](cli-reference/README.md) for command syntax.

Updates land in the [Changelog](https://github.com/mergewayhq/mergeway-cli/releases). File GitHub issues for questions, bugs, or requests.
