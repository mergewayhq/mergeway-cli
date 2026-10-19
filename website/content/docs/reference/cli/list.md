---
title: mergeway-cli list
description: List record identifiers for an entity.
weight: 100
aliases: [/cli-reference/list.html, /cli-reference/list/]
---
# `mergeway-cli list`

## Synopsis
```sh
mergeway-cli [global flags] list --type <entity> [--filter key=value]
```

`--type` is required. `--filter` has no default and performs simple field equality; path-derived fields are eligible. There are no arguments.

## Output, effects, and exits

Prints sorted identifiers, one per line; parent-entity queries include descendants. It changes no files. Success (including no matches) exits `0`; missing/unknown types, malformed filters, or load failures exit `1`.

```sh
mergeway-cli --root examples/full list --type Post --filter author=user-alice
```

Fetch a result with [`get`](../get/) or follow the [record lifecycle](../../../how-to/cli/manage-records/).
