---
title: Create, inspect, update, and delete records
description: Safely complete a file-backed record lifecycle.
weight: 10
aliases: [/guides/manage-records.html, /guides/manage-records/]
---
# Create, inspect, update, and delete records

Work in a clean Git branch and confirm the target entity has a writable one-record-per-file include. This example uses `User` in `examples/full`; use a copy so the destructive final step cannot affect source data.

```sh
cp -R examples/full /tmp/mergeway-records
mergeway-cli --root /tmp/mergeway-records create --type User --id user-carol <<'YAML'
name: Carol Example
email: carol@example.com
roles: [author]
YAML
```

The command prints `User user-carol created` and writes `data/users/user-carol.yaml`. Inspect it:

```sh
mergeway-cli --root /tmp/mergeway-records list --type User
mergeway-cli --root /tmp/mergeway-records get --type User user-carol
```

Merge only the changed fields; omit `--merge` only when you intend to replace the entire record:

```sh
printf '%s\n' 'name: Carol Updated' | \
  mergeway-cli --root /tmp/mergeway-records update --type User --id user-carol --merge
mergeway-cli --root /tmp/mergeway-records validate
```

Review references before deletion, then opt out of the interactive prompt in deliberate automation:

```sh
mergeway-cli --root /tmp/mergeway-records --yes delete --type User user-carol
mergeway-cli --root /tmp/mergeway-records validate
```

Create, update, and delete reject external-root records. See the [`create`](../../../reference/cli/create/), [`update`](../../../reference/cli/update/), and [`delete`](../../../reference/cli/delete/) contracts.
