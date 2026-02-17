---
title: "Install Mergeway CLI"
linkTitle: "Installation"
description: "Install the Mergeway CLI using a release download or Go install."
weight: 10
---

Pick the method that fits your setup. Each installs a single binary named `mergeway-cli`.

## Option 1 – Download a Release (macOS, Linux)

```bash
curl -L https://github.com/mergewayhq/mergeway-cli/releases/download/v0.11.0/mergeway-cli-
$(uname | tr '[:upper:]' '[:lower:]')-amd64.tar.gz \
  | tar -xz
sudo mv mergeway-cli /usr/local/bin/
```

Check the published SHA-256 checksum before moving the binary if you operate in a locked-down environment.

## Option 2 – Go Install (for contributors)

```bash
go install github.com/mergewayhq/mergeway-cli@latest
```

This drops the binary in `$GOPATH/bin`. Prefer tagged versions in production.

## Verify

```bash
mergeway-cli --version
```

You should see something similar to:

```
Mergeway CLI v0.11.0 (commit abc1234)
```

If the command is missing, confirm that the installation path is on your `PATH`.

Move on to the [Getting Started](README.md) guide once the binary is available.
