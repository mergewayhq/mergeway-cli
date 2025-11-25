# `mw entity show`

> **Synopsis:** Print the normalized schema for a given entity.

## Usage

```bash
mw [global flags] entity show <entity>
```

No additional flags. Use `--format json` if you prefer JSON output, and add the global `--root` flag when working outside the workspace root.

## Example

Show the `Post` entity in YAML form:

```bash
mw --root examples/full --format yaml entity show Post
```

Output (abridged):

```yaml
name: Post
source: .../examples/full/entities/Post.yaml
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

- [`mw entity list`](entity-list.md) — find available entities.
- [`mw config export`](config-export.md) — generate a JSON Schema from an entity definition.
