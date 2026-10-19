---
title: mergeway-cli entity list
description: List normalized entity names.
weight: 40
aliases: [/cli-reference/entity-list.html, /cli-reference/entity-list/]
---
# `mergeway-cli entity list`

## Synopsis
```sh
mergeway-cli [global flags] entity list
```

There are no arguments or local flags. Repository-selection flags are inherited.

## Output, effects, and exits

Prints one entity name per line in alphabetical order and changes no files. A successful load exits `0`; configuration errors exit `1` on standard error.

```sh
mergeway-cli --root examples/full entity list
# Comment
# Post
# Tag
# User
```

Inspect one definition with [`entity show`](../entity-show/).
