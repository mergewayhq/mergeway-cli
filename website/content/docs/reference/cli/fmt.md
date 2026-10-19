---
title: mergeway-cli fmt
description: Canonically format configured YAML data files.
weight: 60
aliases: [/cli-reference/fmt.html, /cli-reference/fmt/]
---
# `mergeway-cli fmt`

## Synopsis
```sh
mergeway-cli [global flags] fmt [paths...] [--in-place|--stdout|--lint]
```

| Local flag | Default | Effect |
| --- | --- | --- |
| `--in-place` | effective default | Rewrite target files. |
| `--stdout` | `false` | Print one formatted file; do not write. |
| `--lint` | `false` | Check only; list files that differ. |

The three modes are mutually exclusive. With no paths, all configured YAML data files are targets. Explicit paths may be files, directories, or globs, but every result must belong to the configured data set.

## Output, effects, and exits

Default/`--in-place` rewrites canonical YAML and reports paths. `--stdout` changes nothing. `--lint` exits `1` when formatting differs and `0` when clean. Invalid combinations, unmatched/out-of-scope paths, or I/O failures exit `1`.

```sh
mergeway-cli --root examples/full fmt --lint
```

See [format safely in CI](../../../how-to/cli/format-ci/) and [pre-commit formatting](../../../how-to/cli/pre-commit/).
