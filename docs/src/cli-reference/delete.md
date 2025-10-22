# `mw delete`

Last updated: 2025-10-22

> **Synopsis:** Remove an object file from the workspace.

## Usage

```bash
mw [global flags] delete --type <type> <id>
```

| Flag | Description |
| --- | --- |
| `--type` | Required. Type identifier. |
| `<id>` | Required positional argument identifying the object to delete. |

The command prompts for confirmation unless you pass the global `--yes` flag.

Place global flags (like `--yes` or `--root`) before the command name.

## Example

Run the command from the workspace root (or add `--root` to target another workspace). Delete a user without prompting:

```bash
mw --yes delete --type User user-bob
```

Output:

```
User user-bob deleted
```

## Related Commands

- [`mw list`](list.md) — confirm an object’s identifier before deleting.
- [`mw create`](create.md) — recreate an object if you delete the wrong one.
