# `mw list`

> **Synopsis:** List object identifiers for a given type, optionally filtered by a field.

## Usage

```bash
mw [global flags] list --type <type> [--filter key=value]
```

| Flag | Description |
| --- | --- |
| `--type` | Required. Type identifier to query. |
| `--filter` | Optional `key=value` string used to filter objects before listing their IDs. The comparison is a simple string equality check. |

## Example

Run the command from the workspace root. If you need to operate on another directory, add the global `--root` flag.

List all posts in the quickstart workspace:

```bash
mw list --type Post
```

Output:

```
post-001
```

Filter by author:

```bash
mw list --type Post --filter author=user-alice
```

Output:

```
post-001
```

## Related Commands

- [`mw get`](get.md) — inspect a specific object.
- [`mw create`](create.md) — add a new object when an ID is missing.
