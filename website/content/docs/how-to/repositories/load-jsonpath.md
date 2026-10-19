---
title: Load records with JSONPath
description: Select multiple Mergeway records from an array inside a JSON document.
type: docs
mode: how-to
weight: 20
bookToC: true
---

# Load records with JSONPath

Use a selector when a JSON file wraps the records you want beneath another property. This procedure targets Mergeway v0.10.0 and `examples/jsonpath`.

## Add a structured include

Copy the example, then inspect its `User` entity:

```yaml
entities:
  User:
    identifier: id
    fields:
      id: string
      name: string
    include:
      - path: data/users.json
        selector: "$.users[*]"
```

The selector runs against the decoded JSON document. Each selected object becomes one `User` record and must provide a usable `id`.

## Check the result

Run these commands from the copied repository:

```sh
mergeway-cli validate
mergeway-cli list --type User
mergeway-cli get --type User User-001
```

The list contains `User-001` and `User-002`; the source remains the single `data/users.json` file. Use [`mergeway-cli files`](../../../reference/cli/files/) when you need to see that backing-file relationship.

Selectors cannot be combined with top-level `items` extraction, and every match must be an object. See [Record sources](../../../reference/specification/records/) for the exact JSONPath boundary and [schema design choices](../../../explanation/schema-design/) for the file-granularity trade-off.
