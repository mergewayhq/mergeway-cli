---
title: Validate a repository in GitHub Actions
description: Pin mergeway-cli and reject invalid pull requests.
weight: 40
aliases: [/guides/setup-mergeway-github.html, /guides/setup-mergeway-github/]
---
# Validate a repository in GitHub Actions

Commit this workflow as `.github/workflows/mergeway.yml`. It pins v0.10.0 rather than following `latest` and verifies the downloaded archive before running validation.

```yaml
name: Mergeway
on:
  pull_request:
  push:
    branches: [main]

jobs:
  validate:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - name: Install Mergeway CLI v0.10.0
        shell: bash
        run: |
          curl -fsSLO https://github.com/mergewayhq/mergeway-cli/releases/download/v0.10.0/mergeway-cli_Linux_x86_64.tar.gz
          tar -xzf mergeway-cli_Linux_x86_64.tar.gz mergeway-cli
          sudo install -m 0755 mergeway-cli /usr/local/bin/mergeway-cli
          mergeway-cli version
      - name: Validate configuration and repository
        run: |
          mergeway-cli config lint
          mergeway-cli validate
```

For stronger supply-chain controls, add the release checksum published by your release process and verify it before extraction. Update the pinned version deliberately and review its release notes. Validation exits `1` for repository errors, so no output parsing is needed. Add [`fmt --lint`](../format-ci/) when canonical formatting is also required.
