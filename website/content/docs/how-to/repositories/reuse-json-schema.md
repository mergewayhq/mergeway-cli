---
title: Reuse a JSON Schema
description: Back a Mergeway entity with an existing JSON Schema and validate YAML records against it.
type: docs
mode: how-to
weight: 10
bookToC: true
aliases:
  - /guides/json-schema.html
---

# Reuse a JSON Schema

Use `json_schema` when another tool already owns the record contract or when draft 2020-12 constraints are more suitable than native Mergeway fields. This procedure targets Mergeway v0.10.0 and the release-equivalent `examples/json-schema` fixture.

## Point an entity at the schema

Copy `examples/json-schema` to a working directory. Its entry configuration resolves the schema relative to `mergeway.yaml`:

```yaml
entities:
  Customer:
    description: Customers defined via a JSON Schema file.
    identifier: id
    include:
      - data/customers/*.yaml
    json_schema: schemas/customer.json
```

Keep `fields` absent: `json_schema` and native fields are mutually exclusive. The JSON Schema must describe an object and expose the configured identifier in `properties`.

## Validate and inspect

From the copied repository, run:

```sh
mergeway-cli validate
mergeway-cli --format json entity show Customer
mergeway-cli get --type Customer cust-001
```

Validation applies the imported required fields, nested objects, formats, and string enum/constant choices. `entity show` displays the normalized Mergeway view used by the other tools.

When conversion fails, compare the schema with the supported [JSON Schema interoperability subset](../../../reference/specification/json-schema/). JSON-Schema-backed entities cannot participate in inheritance in specification v1.

For the design trade-off, see [schema design choices](../../../explanation/schema-design/). For general validation failures, use [Troubleshoot the CLI](../../cli/troubleshoot/).
