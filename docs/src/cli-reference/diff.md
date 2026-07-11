---
title: "mergeway-diff"
linkTitle: "mergeway-diff"
description: "Compare Mergeway-managed data between logical repository snapshots."
---

> **Synopsis:** Compare Mergeway-managed data between logical repository snapshots.

## Usage

```bash
mergeway-diff [flags]
mergeway-diff [flags] <left>
mergeway-diff [flags] <left> <right>
mergeway-diff --format json [<left>] [<right>]
```

`mergeway-diff` is a standalone binary. It does not accept `mergeway-cli` subcommands.

## Flags

| Flag       | Description                                                            |
| ---------- | ---------------------------------------------------------------------- |
| `--root`   | Path to the workspace (defaults to `.`).                               |
| `--config` | Explicit path to `mergeway.yaml` (defaults to `<root>/mergeway.yaml`). |
| `--format` | Output format (`yaml` or `json`, default `yaml`).                      |

This command is a data-only diff. It compares Mergeway-managed records across the repository and excludes configuration entirely.

Snapshot interpretation:

- `mergeway-diff` compares `HEAD` data against current working tree data using unstaged changes only.
- `mergeway-diff <left>` compares `<left>` against the current working tree state including unstaged changes.
- `mergeway-diff <left> <right>` compares `<left>` against `<right>`.

Passing more than two positional arguments is an error.

## Examples

Compare the current `HEAD` data against unstaged local changes:

```bash
mergeway-diff
```

Compare an earlier revision against the full current working tree:

```bash
mergeway-diff HEAD~1
```

Emit machine-readable output for automation:

```bash
mergeway-diff --format json HEAD~1 HEAD
```

## Notes

- The command reports semantic record changes, not path-based Git file diffs.
- Output is intentionally simple for now and may be refined in a later phase.
- `mergeway-diff <left>` includes both staged and unstaged working tree changes on the right-hand side.
- `mergeway-diff <left> <right>` compares two explicit revisions and ignores local working tree noise.

## Related Commands

- [`mergeway-cli export`](export.md) — inspect repository data in a serialized form.
- [`mergeway-cli validate`](validate.md) — validate the current repository state before comparing revisions.
