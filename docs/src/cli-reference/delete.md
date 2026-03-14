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
| `<id>`   | Required positional argument identifying the object to delete. For entities that use `identifier: $path`, this is the workspace-relative file path. |

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

For path-based identifiers, delete by file path:

```bash
mergeway-cli --yes delete --type Note data/notes/alpha.yaml
```

Like `create` and `update`, `delete` only operates on files inside the workspace root. If a `$path` record comes from an external include such as `../secondary/products/widget.yaml`, you can inspect and export it, but `delete` will reject that identifier.

## Related Commands

- [`mergeway-cli list`](list.md) — confirm an object’s identifier before deleting.
- [`mergeway-cli create`](create.md) — recreate an object if you delete the wrong one.
