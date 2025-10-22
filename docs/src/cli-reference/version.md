# `mw version`

Last updated: 2025-10-22

> **Synopsis:** Display the CLI build metadata (semantic version, commit, build date).

## Usage

```bash
mw [global flags] version
```

No additional flags.

This command does not touch workspace files; global flags like `--root` are ignored.

## Example

```bash
mw --format json version
```

Output:

```json
{
  "version": "0.1.0",
  "commit": "a713be5",
  "buildDate": "2025-10-22T18:25:03Z"
}
```

Values change with each build; use the command to confirm which binary produced a validation report or data change.

## Related Commands

- [`mw validate`](validate.md) — include the CLI version in validation artifacts for traceability.
