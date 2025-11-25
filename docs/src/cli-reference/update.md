# `mw update`

> **Synopsis:** Modify an existing object. You can replace the object entirely or merge in a subset of fields.

## Usage

```bash
mw [global flags] update --type <type> --id <id> [--file path] [--merge]
```

| Flag | Description |
| --- | --- |
| `--type` | Required. Type identifier. |
| `--id` | Required. Object identifier to update. |
| `--file` | Optional path to a YAML/JSON payload (defaults to STDIN). |
| `--merge` | Merge fields into the existing object instead of replacing it. |

## Example

Run the command from the workspace root (or add `--root` to target another workspace). Update a post title by merging in a tiny payload:

```bash
cat <<'PAYLOAD' > post-update.yaml
title: Launch Day (Updated)
PAYLOAD

mw update --type Post --id post-001 --file post-update.yaml --merge
```

Output:

```
Post post-001 updated
```

Without `--merge`, the payload replaces the entire object.

Run `mw validate` after significant updates to confirm references still resolve.
Delete the temporary payload file once you are done with the update.

## Related Commands

- [`mw create`](create.md) — add new objects.
- [`mw delete`](delete.md) — remove objects that are no longer needed.
