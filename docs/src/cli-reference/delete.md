---
title: "mergeway-cli delete"
linkTitle: "delete"
description: "Remove an object file from the workspace."
---

> **Synopsis:** Remove an object file from the workspace.

## Usage

```bash
mergeway-cli [global flags] delete --type <type> <id>
```

| Flag     | Description                                                    |
| -------- | -------------------------------------------------------------- |
| `--type` | Required. Type identifier.                                     |
| `<id>`   | Required positional argument identifying the object to delete. |

The command prompts for confirmation unless you pass the global `--yes` flag.

Global flags (like `--yes` or `--root`) can appear before or after the command name.

## Example

Run the command from the workspace root (or add `--root` to target another workspace). Delete a user without prompting:

```bash
mergeway-cli --yes delete --type User user-bob
```

Output:

```
User user-bob deleted
```

## Related Commands

- [`mergeway-cli list`](list.md) — confirm an object’s identifier before deleting.
- [`mergeway-cli create`](create.md) — recreate an object if you delete the wrong one.
