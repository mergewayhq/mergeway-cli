---
title: Export repository data
description: Produce deterministic YAML or JSON for another tool.
weight: 60
---
# Export repository data

Validate first so the snapshot is not built from known-invalid records:

```sh
mergeway-cli validate
mergeway-cli --format json export --output repository.json
```

To limit the output, list entity names after `export`:

```sh
mergeway-cli --format json export User Post --output authors-and-posts.json
```

The document maps each entity to an ID-sorted record array. Entity ordering in JSON objects is not a semantic contract; consumers should address keys by name. `--output` replaces an existing destination, so use a temporary path when reviewing:

```sh
tmp_file=$(mktemp)
mergeway-cli --format yaml export User --output "$tmp_file"
less "$tmp_file"
```

See [`export`](../../../reference/cli/export/) for failure behavior and [records/data sources](../../../reference/specification/records/) for the loaded model.
