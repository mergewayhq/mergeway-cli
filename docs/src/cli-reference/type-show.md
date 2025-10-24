# `mw type show`

Last updated: 2025-10-22

> **Synopsis:** Print the normalized schema for a given type.

## Usage

```bash
mw [global flags] type show <type>
```

No additional flags. Use `--format json` if you prefer JSON output, and add the global `--root` flag when working outside the workspace root.

## Example

Show the `Post` type in YAML form:

```bash
mw --root examples --format yaml type show Post
```

Output (abridged):

```yaml
name: Post
source: .../examples/types/Post.yaml
identifier:
  field: id
filepatterns:
  - data/posts/*.yaml
fields:
  title:
    type: string
    required: true
  author:
    type: User
    required: true
  body:
    type: string
```

## Related Commands

- [`mw type list`](type-list.md) — find available types.
- [`mw config export`](config-export.md) — generate a JSON Schema from an entity definition.
