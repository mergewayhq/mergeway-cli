# `mw type list`

Last updated: 2025-10-22

> **Synopsis:** Show every type Mergeway discovered from your configuration.

## Usage

```bash
mw [global flags] type list
```

No command-specific flags. Add the global `--root` flag if you need to inspect another workspace.

## Example

List types for the `examples/` workspace bundled with the repository:

```bash
mw --root examples type list
```

Output:

```
Comment
Post
Tag
User
```

Types are listed alphabetically.

## Related Commands

- [`mw type show`](type-show.md) — inspect an individual schema definition.
- [`mw config lint`](config-lint.md) — verify the configuration if a type is missing.
