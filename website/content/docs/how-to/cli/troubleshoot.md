---
title: Troubleshoot the CLI
description: Diagnose repository roots, configuration loading, validation, mutation, and output failures.
type: docs
mode: how-to
weight: 90
bookToC: true
---

# Troubleshoot the CLI

## The repository is not found

Run commands from the directory containing `mergeway.yaml`, or pass `--root` before the command. Use `--config` when the entry file has another path. The CLI does not use the LSP client's workspace-folder negotiation.

## Configuration does not load

Run [`mergeway-cli config lint`](../../../reference/cli/config-lint/) before full validation. Check include paths relative to the file that declares them, entity names, inheritance cycles, and mutually exclusive `fields`/`json_schema` members.

## Validation fails

Run [`mergeway-cli validate`](../../../reference/cli/validate/) without `--fail-fast` to collect all errors, then isolate `format`, `schema`, or `references` with repeated `--phase` flags. The [validation model](../../../explanation/validation-model/) explains why a later phase waits for earlier phases.

## A write command refuses a record

Inline records and path-derived fields are read-only. Mutation also refuses unsafe targets outside the repository root. Use [`files`](../../../reference/cli/files/) to find a record's source, preview formatting with `fmt --stdout`, and keep a Git diff before accepting an in-place change.

## Output is hard to automate

Place the global `--format json` option before the command and capture standard output. Diagnostics and command failures use standard error. Look up the exact flags, output, and side effects in the [`mergeway-cli` reference](../../../reference/cli/).
