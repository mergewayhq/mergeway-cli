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

This command is a data-only diff. It compares Mergeway-managed records across the repository and excludes configuration entirely.

Snapshot interpretation:

- `mergeway-diff` compares `HEAD` data against current working tree data using unstaged changes only.
- `mergeway-diff <left>` compares `<left>` against the current working tree state including unstaged changes.
- `mergeway-diff <left> <right>` compares `<left>` against `<right>`.

Passing more than two positional arguments is an error.

## Notes

- The command reports semantic record changes, not path-based Git file diffs.
- Output is intentionally simple for now and may be refined in a later phase.
- Pass `--format json` to emit a machine-readable semantic diff document.
- Supported flags:
  - `--root` sets the workspace root and defaults to `.`
  - `--config` sets an explicit `mergeway.yaml`
  - `--format` chooses `yaml` or `json`

## Related Commands

- [`mergeway-cli export`](export.md) — inspect repository data in a serialized form.
- [`mergeway-cli validate`](validate.md) — validate the current repository state before comparing revisions.
