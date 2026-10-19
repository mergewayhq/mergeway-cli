---
title: mergeway-cli config lint
description: Load and validate configuration without loading records.
weight: 70
aliases: [/cli-reference/config-lint.html, /cli-reference/config-lint/]
---
# `mergeway-cli config lint`

## Synopsis
```sh
mergeway-cli [global flags] config lint
```

There are no arguments or local flags; it inherits repository-selection flags.

## Output, effects, and exits

Prints `configuration valid` when the entry file, includes, and normalized definitions load. It does not load record data and changes no files. Success exits `0`; syntax, include, version, or normalization errors exit `1`.

```sh
mergeway-cli --root examples/full config lint
```

Use [`validate`](../validate/) for records and references, or read [which validation to choose](../../../how-to/cli/choose-validation/).
