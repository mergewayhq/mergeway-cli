---
title: mergeway-cli version
description: Display CLI build metadata.
weight: 30
aliases: [/cli-reference/version.html, /cli-reference/version/]
---
# `mergeway-cli version`

## Synopsis
```sh
mergeway-cli [global flags] version
```

There are no arguments or local flags. `--format` (`yaml` by default, or `json`) controls the result; other inherited repository flags have no effect.

## Output, effects, and exits

Prints `version`, `commit`, and `buildDate`. Values depend on the build. It changes no files and exits `0` unless output serialization fails.

```sh
mergeway-cli --format json version
```

Use this output to identify the binary behind validation or automation results. See the [CLI overview](../).
