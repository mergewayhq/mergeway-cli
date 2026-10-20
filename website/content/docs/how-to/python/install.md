---
title: Install the Python SDK
description: Install mergeway-python and verify its Python and CLI prerequisites.
type: docs
mode: how-to
weight: 10
bookToC: true
aliases:
  - /python/installation/
---

# Install the Python SDK

The published package is named `mergeway`. It requires Python 3.13 or newer and runs `mergeway-cli` as a subprocess, so install both before using the SDK.

## Install the CLI dependency

Follow [Install Mergeway tools](../../install/) if `mergeway-cli` is not already available. Verify that the Python process will be able to find it:

```sh
mergeway-cli version
```

If the binary is not on `PATH`, pass its path as `cli_binary` when you construct a `Database`.

## Add the Python package

With `uv`, add the package to a project:

```sh
uv add mergeway
```

With `pip`, install it into the active environment:

```sh
python -m pip install mergeway
```

The package also installs the `mergeway-python` model-generation command.

## Verify the installation

Check that Python can import the SDK and that the generator command is available:

```sh
python -c 'from mergeway import Database; print(Database.__name__)'
mergeway-python --help
```

The first command should print `Database`. The second should show the `generate` subcommand.

Continue with [Get started with `mergeway-python`](../getting-started/) to connect to a repository and generate typed models.

