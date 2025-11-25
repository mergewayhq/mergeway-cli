# `mw entity list`

> **Synopsis:** Show every entity Mergeway discovered from your configuration.

## Usage

```bash
mw [global flags] entity list
```

No command-specific flags. Add the global `--root` flag if you need to inspect another workspace.

## Example

List entities for the `examples/` workspace bundled with the repository:

```bash
mw --root examples/full entity list
```

Output:

```
Comment
Post
Tag
User
```

Entities are listed alphabetically.

## Related Commands

- [`mw entity show`](entity-show.md) — inspect an individual schema definition.
- [`mw config lint`](config-lint.md) — verify the configuration if an entity is missing.
