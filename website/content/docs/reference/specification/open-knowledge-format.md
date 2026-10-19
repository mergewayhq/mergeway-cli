---
title: Open Knowledge Format mapping
description: Normative mapping of OKF Markdown concept documents to Mergeway records in specification v1.
type: docs
mode: reference
weight: 120
bookToC: true
---

# Open Knowledge Format mapping

> **Normative v1.** This page applies to repositories declaring `mergeway.version: 1` and is normative.

## Recognized sources

An entity include that matches a `.md` or `.okf` file reads it as an Open Knowledge Format concept document. The document MUST begin with a line containing exactly `---` and MUST contain a later closing delimiter line. Content between the delimiters is decoded as a YAML map.

Frontmatter MUST contain a non-empty `type` whose value exactly equals the including entity. `type` associates the source and is removed from record fields. Other frontmatter entries become fields. All content after the closing delimiter becomes the string field `body`, preserving its Markdown text.

The checked example is:

{{< spec-example path="testdata/specification-v1/data/knowledge/orders.md" language="markdown" >}}

## Identity and fields

The resulting record follows ordinary [field validation](../fields/) and [reference validation](../identifiers-references/). It MUST obtain an identifier from a frontmatter field, `identifier: $path`, or a path-derived field that is also the configured identifier.

One OKF document represents exactly one record. Selector and `items` multi-record semantics do not apply to its Markdown body. A `body` frontmatter key, if supplied, is replaced by the Markdown content.

This mapping is the complete normative OKF behavior in version 1; the external OKF vocabulary does not implicitly create Mergeway fields or constraints.
