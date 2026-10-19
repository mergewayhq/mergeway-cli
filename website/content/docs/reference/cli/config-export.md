---
title: mergeway-cli config export
description: Export one entity definition as JSON Schema.
weight: 80
aliases: [/cli-reference/config-export.html, /cli-reference/config-export/]
---
# `mergeway-cli config export`

## Synopsis
```sh
mergeway-cli [global flags] config export --type <entity>
```

`--type` is required and has no default. There are no arguments. Inherited `--format` controls whether the schema document is emitted as YAML (default) or JSON.

## Output, effects, and exits

Prints a draft 2020-12 JSON Schema representation and changes no files. Reference fields use `x-reference-type`. Success exits `0`; missing/unknown entities, load failures, and unsupported reference unions exit `1`.

```sh
mergeway-cli --root examples/full --format json config export --type Post
```

See [JSON Schema interoperability](../../specification/json-schema/) and [`entity show`](../entity-show/).
