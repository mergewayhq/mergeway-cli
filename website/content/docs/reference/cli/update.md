---
title: mergeway-cli update
description: Replace or merge fields in a file-backed record.
weight: 140
aliases: [/cli-reference/update.html, /cli-reference/update/]
---
# `mergeway-cli update`

## Synopsis
```sh
mergeway-cli [global flags] update --type <entity> --id <id> [--file <payload>] [--merge]
```

`--type` and `--id` are required. `--file` defaults to standard input. `--merge` defaults to `false`: without it, the payload replaces the record; with it, supplied fields overlay existing fields.

## Output, effects, and exits

Rewrites the backing YAML file and prints `<Entity> <id> updated`. External-root records are read-only. Success exits `0`; missing/invalid input, unknown records, validation, or write failures exit `1`.

```sh
printf '%s\n' 'title: Updated title' | mergeway-cli update --type Post --id post-001 --merge
```

See [manage records](../../../how-to/cli/manage-records/) and [`get`](../get/).
