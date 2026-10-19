---
title: Derive identifiers and fields from paths
description: Use record paths as identifiers or expose selected path segments as read-only fields.
type: docs
mode: how-to
weight: 40
bookToC: true
---

# Derive identifiers and fields from paths

Use path-derived data when the repository layout already carries stable identity or classification. This procedure targets Mergeway v0.10.0 and `examples/path-identifier` plus `examples/path-segments`.

## Use the complete path as identity

`examples/path-identifier` configures:

```yaml
Note:
  identifier: $path
  include:
    - data/notes/*.yaml
  fields:
    title: { type: string, required: true }
    body: string
```

The records do not store an ID. Confirm the normalized repository-relative identifiers:

```sh
mergeway-cli --root examples/path-identifier validate
mergeway-cli --root examples/path-identifier list --type Note
mergeway-cli --root examples/path-identifier get --type Note data/notes/alpha.yaml
```

## Expose selected path values

`examples/path-segments` keeps a field identifier but derives three read-only fields:

```yaml
section:
  type: string
  required: true
  source: { path_segment: 2 }
filename:
  type: string
  required: true
  source: { path_segment_rev: 0 }
relative_path:
  type: string
  required: true
  source: { path: true }
```

Inspect the computed values and filter on one:

```sh
mergeway-cli --root examples/path-segments get --type Page guide-install
mergeway-cli --root examples/path-segments list --type Page --filter 'section=guides'
```

Mergeway derives these fields for reading, filtering, exporting, and validation; it does not write them into source files. Review the exact type, index, inline-data, and multi-record restrictions in [Path-derived data](../../../reference/specification/path-derived-data/). Use [schema design choices](../../../explanation/schema-design/) before coupling an identifier to a movable path.
