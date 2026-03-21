---
title: "mergeway-cli version"
linkTitle: "version"
description: "Display the CLI build metadata (semantic version, commit, build date)."
---

> **Synopsis:** Display the CLI build metadata (semantic version, commit, build date).

## Usage

```bash
mergeway-cli [global flags] version
```

No additional flags.

This command does not touch workspace files; global flags like `--root` are ignored.

## Example

```bash
mergeway-cli --format json version
```

Output:

```json
{
  "version": "0.4.1",
  "commit": "a713be5",
  "buildDate": "2025-10-22T18:25:03Z"
}
```

Values change with each build; use the command to confirm which binary produced a validation report or data change.

## Related Commands

- [`mergeway-cli validate`](validate.md) — include the CLI version in validation artifacts for traceability.
