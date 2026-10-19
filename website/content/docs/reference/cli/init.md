---
title: mergeway-cli init
description: Scaffold a Mergeway repository entry file.
weight: 10
aliases: [/cli-reference/init.html, /cli-reference/init/]
---
# `mergeway-cli init`

Creates the configured entry file without overwriting an existing file.

## Synopsis
```sh
mergeway-cli [global flags] init
```

There are no arguments or local flags. It inherits all [global flags](../); `--root` selects the directory and `--config` may select a different entry filename.

## Output and effects

The command creates missing parent directories and a starter v1 configuration, then prints `Initialized repository at <root>`. Re-running it leaves an existing entry file unchanged. Success exits `0`; positional arguments or filesystem failures exit `1`.

## Example
```sh
mkdir blog-metadata
mergeway-cli --root blog-metadata init
```

Next, [split schemas and records across files](../../../how-to/cli/split-files/) or run [`config lint`](../config-lint/).
