---
title: mergeway-cli create
description: Create a file-backed record.
weight: 130
aliases: [/cli-reference/create.html, /cli-reference/create/]
---
# `mergeway-cli create`

## Synopsis
```sh
mergeway-cli [global flags] create --type <entity> [--file <payload>] [--id <id>]
```

`--type` is required. `--file` defaults to standard input. `--id` overrides a field identifier and is required for `$path` entities.

## Output, effects, and exits

Creates a YAML record through the entity's writable include pattern and prints `<Entity> <id> created`. Writes are restricted to the repository root; inline-only, grouped, ambiguous, duplicate, invalid, or external targets fail. Success exits `0`; input, validation, resolution, or write errors exit `1`.

```sh
printf '%s\n' 'name: Bob Example' | mergeway-cli create --type User --id user-bob
```

Use the [record lifecycle guide](../../../how-to/cli/manage-records/) and [`validate`](../validate/) after mutation.
