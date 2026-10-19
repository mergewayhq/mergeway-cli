---
title: mergeway-cli files
description: List YAML files included by entity definitions.
weight: 110
aliases: [/cli-reference/files.html, /cli-reference/files/]
---
# `mergeway-cli files`

## Synopsis
```sh
mergeway-cli [global flags] files [--type <entity>] [--group]
```

`--type` optionally restricts one entity; `--group` (default `false`) collapses one-record-per-file wildcard storage to its container pattern. There are no arguments. `--format` defaults to YAML.

## Output, effects, and exits

Prints sorted `{type, file}` entries for YAML includes only; inline and JSON sources are excluded. It changes no files. Success exits `0`; unknown entities, ambiguous configured files, or load/filesystem errors exit `1`.

```sh
mergeway-cli --root examples/full --format json files --group
```

See [split schemas and records](../../../how-to/cli/split-files/) and [`list`](../list/).
