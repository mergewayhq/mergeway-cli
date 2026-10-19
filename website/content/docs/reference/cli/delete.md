---
title: mergeway-cli delete
description: Delete a file-backed record after confirmation.
weight: 150
aliases: [/cli-reference/delete.html, /cli-reference/delete/]
---
# `mergeway-cli delete`

## Synopsis
```sh
mergeway-cli [global flags] delete --type <entity> <id>
```

`--type` and `<id>` are required. The command prompts on standard error; inherited `--yes` (default `false`) skips confirmation.

## Output, effects, and exits

Deletes the local backing file and prints `<Entity> <id> deleted`. It will not delete an external-root record. Success exits `0`; rejection (`aborted`), missing/unknown records, or filesystem failures exit `1`.

```sh
mergeway-cli --yes delete --type User user-bob
```

Destructive examples belong on temporary copies; follow [manage records](../../../how-to/cli/manage-records/) and validate references afterward.
