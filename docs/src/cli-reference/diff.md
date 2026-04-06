---
title: "mergeway-cli diff"
linkTitle: "diff"
description: "Compare Mergeway-managed data between logical repository snapshots."
---

> **Synopsis:** Compare Mergeway-managed data between logical repository snapshots.

## Usage

```bash
mergeway-cli [global flags] diff
mergeway-cli [global flags] diff <left>
mergeway-cli [global flags] diff <left> <right>
mergeway-cli [global flags] --format json diff [<left>] [<right>]
```

This command is a data-only diff. It compares Mergeway-managed records across the repository and excludes configuration entirely.

Snapshot interpretation:

- `diff` compares `HEAD` data against current working tree data using unstaged changes only.
- `diff <left>` compares `<left>` against the current working tree state including unstaged changes.
- `diff <left> <right>` compares `<left>` against `<right>`.

Passing more than two positional arguments is an error.

## Notes

- The command reports semantic record changes, not path-based Git file diffs.
- Output is intentionally simple for now and may be refined in a later phase.
- Pass the global `--format json` flag to emit a machine-readable semantic diff document.

## Related Commands

- [`mergeway-cli export`](export.md) — inspect repository data in a serialized form.
- [`mergeway-cli validate`](validate.md) — validate the current repository state before comparing revisions.
