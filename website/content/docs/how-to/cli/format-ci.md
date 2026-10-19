---
title: Format safely in CI
description: Check canonical formatting without rewriting the CI checkout.
weight: 80
---
# Format safely in CI

Use lint mode in automation; plain `fmt` rewrites files and can conceal the failure you meant to detect.

```sh
mergeway-cli fmt --lint
```

A clean repository exits `0` without output. When files differ, their paths are printed and the command exits `1` without modifying them. Contributors can fix the same failure locally:

```sh
mergeway-cli fmt
git diff --check
mergeway-cli fmt --lint
```

To inspect a single file without writing it, use `--stdout` and do not combine it with `--lint` or `--in-place`:

```sh
mergeway-cli fmt --stdout data/posts/launch.yaml > /tmp/launch.formatted.yaml
diff -u data/posts/launch.yaml /tmp/launch.formatted.yaml
```

Only configured YAML data is formatted. Pair this check with [full validation](../choose-validation/) and consult the [`fmt` reference](../../../reference/cli/fmt/).
