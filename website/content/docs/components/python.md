---
title: Python SDK
description: Automate Mergeway repositories from Python with generated typed models.
type: docs
weight: 40
bookToC: true
aliases:
  - /python/
---

# Python SDK

`mergeway-python` is the Python automation layer for Mergeway repositories. The published [`mergeway` package on PyPI](https://pypi.org/project/mergeway/) wraps `mergeway-cli`, uses the CLI's JSON output as its data boundary, and can generate typed Python models from a repository schema.

## Who it is for

Python application developers, data engineers, and automation authors use the SDK when shelling out to the CLI directly would make typed application code difficult to maintain.

## Purpose and boundaries

The SDK can read, create, update, and delete records; validate and format a repository; export repository data; and generate Python dataclasses from `mergeway.yaml`. It is an automation wrapper, not a separate storage engine or independent implementation of the repository specification.

## Relationship to the suite

Each `Database` operation invokes `mergeway-cli` against the configured repository. The CLI remains the source of truth for loading, mutation, formatting, validation, and JSON output; the SDK adds Python objects and generated types around that interface.

Because the SDK starts the CLI as a subprocess, `mergeway-cli` must be installed and available on `PATH`. You can provide another executable name or path through the `cli_binary` argument when constructing `Database`.

## Availability and compatibility

The SDK is published separately from the Go binaries as the [`mergeway` Python package](https://pypi.org/project/mergeway/) and requires Python 3.13 or newer. Its source is maintained in the [`mergewayhq/mergeway-python` repository](https://github.com/mergewayhq/mergeway-python). Match the SDK to a CLI version whose commands and JSON output it supports; installing the Python package does not install `mergeway-cli`.

## Start here

- [Install the Python SDK](../../how-to/python/install/) and its CLI dependency.
- [Get started with `mergeway-python`](../../how-to/python/getting-started/) to generate models and automate repository operations.
- [Build your first repository](../../tutorials/first-repository/) before adding Python automation if you do not already have one.
- Use the [`mergeway-cli` reference](../../reference/cli/) to understand the underlying operations and their side effects.

