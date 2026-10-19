---
title: Split schemas and records across files
description: Replace inline definitions and records with deterministic includes.
weight: 30
aliases: [/guides/split-files.html, /guides/split-files/]
---
# Split schemas and records across files

Create predictable directories from the repository root:

```sh
mkdir -p entities data/users
```

Move each entity into a v1 configuration fragment such as `entities/User.yaml`, and make the entry file include those fragments:

```yaml
mergeway:
  version: 1
include:
  - entities/*.yaml
```

In the `User` fragment, point the entity at its record files:

```yaml
mergeway:
  version: 1
entities:
  User:
    identifier: id
    include:
      - data/users/*.yaml
    fields:
      id: {type: string, required: true}
      name: {type: string, required: true}
```

Write one record per file, for example `data/users/user-alice.yaml`, then verify both layers:

```sh
mergeway-cli config lint
mergeway-cli files --type User
mergeway-cli list --type User
mergeway-cli validate
```

Configuration includes resolve from their declaring configuration file; entity data includes resolve from the repository root. See [repository configuration](../../../reference/specification/repository-configuration/) and [records/data sources](../../../reference/specification/records/).
