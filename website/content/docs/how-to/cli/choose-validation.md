---
title: Validate configuration or a repository
description: Choose config lint or full, phased repository validation.
weight: 20
---
# Validate configuration or a repository

Use configuration lint for fast feedback while editing `mergeway.yaml` or included entity files:

```sh
mergeway-cli config lint
```

It proves that configuration parses and normalizes, but it does not open record files. Before review or export, validate the full repository:

```sh
mergeway-cli validate
```

For machine-readable diagnostics and an early stop:

```sh
mergeway-cli --format json --fail-fast validate
```

To isolate a problem, repeat `--phase` as needed:

```sh
mergeway-cli validate --phase format
mergeway-cli validate --phase schema --phase references
```

Reference validation automatically includes schema validation. Any reported validation error exits `1`, so CI can use the command directly. See the [validation model](../../../explanation/validation-model/) and [`validate` reference](../../../reference/cli/validate/).
