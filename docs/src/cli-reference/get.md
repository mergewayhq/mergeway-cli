# `mw get`

Last updated: 2025-10-22

> **Synopsis:** Print the fields of one object.

## Usage

```bash
mw [global flags] get --type <type> <id>
```

| Flag | Description |
| --- | --- |
| `--type` | Required. Type identifier that owns the object. |
| `<id>` | Required positional argument representing the object identifier. |

Use `--format json` if you prefer JSON output.

## Example

Run the command from the workspace root. Use `--root` if you need to target another workspace.

Fetch the `post-001` record as YAML:

```bash
mw --format yaml get --type Post post-001
```

Output:

```yaml
author: user-alice
body: |
    We are excited to announce the product launch.
id: post-001
title: Launch Day
```

## Related Commands

- [`mw list`](list.md) — discover identifiers before calling `get`.
- [`mw update`](update.md) — change object fields.
