---
title: "mergeway-cli update"
linkTitle: "update"
description: "Modify an existing object by replacing it or merging in fields."
---

> **Synopsis:** Modify an existing object. You can replace the object entirely or merge in a subset of fields.

## Usage

````

bash
mergeway-cli [global flags] update --type <type> --id <id> [--file path] [--merge]
```| Flag | Description |
| --- | --- |
| `--file` | Optional path to a YAML/JSON payload (defaults to STDIN). |
| `--merge` | Merge fields into the existing object instead of replacing it. |
type` | Required. Type identifier. |
| `--id` | Required. Object identifier to update. |
| `--
```bash
## Example

Run the command from the workspace root (or add `--root` to target another workspace). Update a post title by merging in a tiny payload:

cat <<'PAYLOAD' > post-update.yaml
title: Launch Day (Updated)
PAYLOAD

mergeway-cli update --type Post --id post-001 --file post-update.yaml --merge
````

```
Output:

Post post-001 updated
```

Run `mergeway-cli validate` after significant updates to confirm references still resolve.
Without `--merge`, the payload replaces the entire object.

Delete the temporary payload file once you are done with the update.

## Related Commands

- [`mergeway-cli create`](create.md) — add new objects.
- [`mergeway-cli delete`](delete.md) — remove objects that are no longer needed.
