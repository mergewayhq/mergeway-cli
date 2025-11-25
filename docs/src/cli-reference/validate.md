# `mw validate`

> **Synopsis:** Check schemas, records, and references, emitting formatted errors when something is wrong.

## Usage

```bash
mw [global flags] validate [--phase format|schema|references]... [--fail-fast]
```

| Flag | Description |
| --- | --- |
| `--phase` | Optional. Repeat to run a subset of phases. By default all phases run (`format`, `schema`, then `references`). |
| `--fail-fast` | Stop after the first error. Defaults to the global `--fail-fast` flag. |

When you request the `references` phase, Mergeway automatically includes the `schema` phase so reference checks have the information they need.

## Examples

Run the command from the workspace root (or add `--root` to point elsewhere).

Validate the current workspace:

```bash
mw validate
```

Add `--format json` when you need machine-readable output.

Output:

```
validation succeeded
```

Run validation after introducing a breaking schema change:

```bash
mw validate
```

Output when the `Post` schema requires an `author` but the record is missing it:

```yaml
- phase: schema
  type: Post
  id: post-001
  file: data/posts/launch.yaml
  message: missing required field "author"
```

The command writes errors to standard output and still exits with status `0`, so automation should check whether any errors were returned.

## Related Commands

- [`mw config lint`](config-lint.md) — validate configuration without loading data.
- [`mw list`](list.md) — locate the objects mentioned in validation errors.
