---
title: "mergeway-cli list"
linkTitle: "list"
description: "List object identifiers for a given type, optionally filtered by a field."
---

> **Synopsis:** List object identifiers for a given type, optionally filtered by a field.

## Usage

```bash
mergeway-cli [global flags] list --type <type> [--filter key=value]
```

| Flag       | Description                                                                                                                    |
| ---------- | ------------------------------------------------------------------------------------------------------------------------------ |
| `--type`   | Required. Type identifier to query. When the type has descendants, the list also includes objects from those descendant types. |
| `--filter` | Optional `key=value` string used to filter objects before listing their IDs. The comparison is a simple string equality check. Declared read-only fields derived from file paths can also be used here. |

## Example

Run the command from the workspace root. If you need to operate on another directory, add the global `--root` flag.

List all posts in the quickstart workspace:

```bash
mergeway-cli list --type Post
```

Output:

```
post-001
```

If an entity uses `identifier: $path`, the output contains workspace-relative file paths such as `data/notes/alpha.yaml`. When the entity reads files from outside the workspace root, the identifier may contain `../...`, for example `../secondary/products/widget.yaml`.

Filter by author:

```bash
mergeway-cli list --type Post --filter author=user-alice
```

Output:

```
post-001
```

Filter by a declared path-derived field:

```bash
mergeway-cli list --type Page --filter 'section=guides'
```

For inherited entities, querying the parent includes descendant objects:

```bash
mergeway-cli --root examples/inheritance list --type Animal
```

Output:

```
dog-1
```

## Related Commands

- [`mergeway-cli get`](get.md) — inspect a specific object.
- [`mergeway-cli create`](create.md) — add a new object when an ID is missing.
