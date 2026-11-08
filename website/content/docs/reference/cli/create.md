---
title: mergeway-cli create
description: Create a file-backed record.
weight: 130
aliases: [/cli-reference/create.html, /cli-reference/create/]
---
# `mergeway-cli create`

## Synopsis
```sh
mergeway-cli [global flags] create --type <entity> [--file <payload>] [--id <id>] [--target <path>]
```

`--type` is required. `--file` defaults to standard input. `--id` overrides a field identifier. `--target` optionally selects a workspace-relative output path and must match a writable, non-selector include for the entity. For `$path` entities, either `--id` or `--target` is required; when both are present, they must resolve to the same path.

## Output, effects, and exits

Creates a YAML record through the entity's writable include pattern and prints `<Entity> <id> created`. Writes are restricted to the repository root; inline-only, grouped, ambiguous, duplicate, invalid, or external targets fail. Success exits `0`; input, validation, resolution, or write errors exit `1`.

```sh
printf '%s\n' 'name: Bob Example' | mergeway-cli create --type User --id user-bob
```

Choose a specific configured output location with `--target`:

```sh
printf '%s\n' 'name: Bob Example' | mergeway-cli create --type User --id user-bob --target data/users/user-bob.json
```

Absolute paths, parent traversal, paths outside the entity's includes, and escapes through symlinks are rejected. Without `--target`, the destination continues to be inferred from the entity's includes.

Use the [record lifecycle guide](../../../how-to/cli/manage-records/) and [`validate`](../validate/) after mutation.
