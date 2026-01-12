---
title: "mergeway-cli fmt"
linkTitle: "fmt"
---

> **Synopsis:** Format one or more data/config files using Mergeway’s canonical ordering.

## Usage

```bash
mw [global flags] fmt [--in-place|--lint] [<file>...]
```

| Flag         | Description                                                                            |
| ------------ | -------------------------------------------------------------------------------------- |
| `--in-place` | Rewrite each file on disk with the formatted content (default when no other flag set). |
| `--stdout`   | Print formatted content to stdout instead of touching files.                           |
| `--lint`     | Do not rewrite files; exit `1` if any file would change and print the offending paths. |

You can't combine `--stdout` with `--lint` or `--in-place`. When neither flag is supplied, `mw fmt` rewrites files in place and prints a line for each path it touched.

If you omit file arguments entirely, the command formats every file referenced by the `include` directives in `mergeway.yaml`. Supplying explicit files narrows the scope, but each file needs to belong to the configured data set—`mw fmt` fails fast when a path is not declared in the config.

When formatting entity data, field order follows the definition in your schema so that e.g. `id`, `title`, and other properties always appear in the same sequence you specified.

## Examples

Preview a single file without modifying disk:

```bash
mw fmt --stdout data/posts/posts.yaml > /tmp/posts.yaml
```

Rewrite files in place (default behavior, useful before committing):

```bash
mw fmt data/posts/posts.yaml data/users.yaml
```

Use lint mode in CI to ensure working tree files are already formatted:

```bash
mw fmt --lint data/posts/posts.yaml
```

If any file requires formatting, the command prints each relative path and exits with status `1`. Clean runs exit with status `0` and no output. Passing a file that is not listed in the configuration returns an error, which helps CI avoid accidentally mutating out-of-scope files.

## Related Commands

- [`mw validate`](validate.md) — validate schemas and data after formatting.
- [`mw list`](list.md) — inspect records that were just formatted.
