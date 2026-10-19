---
title: Enforce canonical formatting with pre-commit
description: Rewrite configured records before each commit.
weight: 50
aliases: [/guides/setup-mergeway-pre-commit.html, /guides/setup-mergeway-pre-commit/]
---
# Enforce canonical formatting with pre-commit

Install `mergeway-cli` and [pre-commit](https://pre-commit.com/), then add `.pre-commit-config.yaml`:

```yaml
repos:
  - repo: local
    hooks:
      - id: mergeway-fmt
        name: Canonicalize Mergeway records
        entry: mergeway-cli fmt
        language: system
        pass_filenames: false
        files: ^data/.*\.(ya?ml)$
```

`pass_filenames: false` lets Mergeway derive its complete configured data set instead of receiving an arbitrary staged subset. Install and test the hook:

```sh
pre-commit install
pre-commit run mergeway-fmt --all-files
git diff --check
```

The default formatter rewrites files; review and stage those changes. If a non-mutating local policy is preferred, set `entry: mergeway-cli fmt --lint`. See the [`fmt` modes](../../../reference/cli/fmt/) and [CI formatting guide](../format-ci/).
