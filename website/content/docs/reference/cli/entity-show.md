---
title: mergeway-cli entity show
description: Print one normalized entity definition.
weight: 50
aliases: [/cli-reference/entity-show.html, /cli-reference/entity-show/]
---
# `mergeway-cli entity show`

## Synopsis
```sh
mergeway-cli [global flags] entity show <name>
```

`<name>` is required. There are no local flags; inherited `--format` defaults to YAML and may be `json`.

## Output, effects, and exits

Prints the effective normalized entity, including inherited fields and source metadata. It changes no files. A known entity exits `0`; a missing name, unknown entity, or load error exits `1`.

```sh
mergeway-cli --root examples/full --format yaml entity show Post
```

Use [`entity list`](../entity-list/) to discover names or [`config export`](../config-export/) for JSON Schema.
