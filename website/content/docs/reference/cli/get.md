---
title: mergeway-cli get
description: Print one record by entity and identifier.
weight: 120
aliases: [/cli-reference/get.html, /cli-reference/get/]
---
# `mergeway-cli get`

## Synopsis
```sh
mergeway-cli [global flags] get --type <entity> <id>
```

`--type` and `<id>` are required. Inherited `--format` defaults to YAML. `$path` identifiers use the workspace-relative file path and may include `../` for external read-only records.

## Output, effects, and exits

Prints record fields, including declared path-derived fields, and changes no files. Parent types may resolve descendants. Success exits `0`; missing arguments, unknown records/types, or load failures exit `1`.

```sh
mergeway-cli --root examples/full --format json get --type Post post-001
```

Discover IDs with [`list`](../list/) or modify local records with [`update`](../update/).
