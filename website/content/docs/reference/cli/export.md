---
title: mergeway-cli export
description: Export repository records as one YAML or JSON document.
weight: 160
aliases: [/cli-reference/export.html, /cli-reference/export/]
---
# `mergeway-cli export`

## Synopsis
```sh
mergeway-cli [global flags] export [entities...] [--output <path>]
```

Entity arguments are optional; omitting them exports all entities. `--output` defaults to standard output. Inherited `--format` defaults to YAML.

## Output, effects, and exits

Produces a map from entity names to ID-sorted record arrays, including derived fields. `--output` creates or truncates its destination; otherwise repository files are untouched. Success exits `0`; unknown entities, load/serialization errors, or output-file failures exit `1`.

```sh
mergeway-cli --root examples/full --format json export User Post --output snapshot.json
```

See [export repository data](../../../how-to/cli/export-data/) and [`get`](../get/).
