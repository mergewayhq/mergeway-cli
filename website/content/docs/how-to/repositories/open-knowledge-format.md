---
title: Load Open Knowledge Format documents
description: Map OKF or frontmatter Markdown concept files to Mergeway records.
type: docs
mode: how-to
weight: 50
bookToC: true
---

# Load Open Knowledge Format documents

Mergeway v0.10.0 supports the checked `examples/okf` fixture, so this workflow is part of the launch documentation rather than deferred content.

## Configure the mapping

Copy `examples/okf`. Its `Knowledge` entity includes both supported extensions and derives `id` from the filename:

```yaml
Knowledge:
  identifier: id
  include:
    - data/concepts/*.md
    - data/concepts/*.okf
  fields:
    id:
      type: string
      required: true
      source: { path_segment_rev: 0 }
    title: { type: string, required: true }
    description: string
    tags: { type: string, repeated: true }
    body: { type: string, required: true }
```

Each document begins with YAML frontmatter whose `type` is exactly `Knowledge`. Mergeway removes `type`, maps the other frontmatter keys to fields, and exposes the content after the closing delimiter as `body`.

## Validate and inspect

```sh
mergeway-cli --root examples/okf validate
mergeway-cli --root examples/okf --format json get --type Knowledge customer-orders.md
mergeway-cli --root examples/okf --format json get --type Knowledge order-returns.okf
```

One document produces one record. A frontmatter `body` value is replaced by document content, and the external OKF vocabulary does not add implicit fields. See the normative [Open Knowledge Format mapping](../../../reference/specification/open-knowledge-format/) and [Path-derived data](../../../reference/specification/path-derived-data/).
