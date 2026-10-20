---
title: Get started with mergeway-python
description: Generate typed models and automate common Mergeway repository operations from Python.
type: docs
mode: how-to
weight: 20
bookToC: true
aliases:
  - /python/getting-started/
---

# Get started with `mergeway-python`

This guide connects the Python SDK to an existing Mergeway repository, generates typed models, and demonstrates its read, write, validation, formatting, and export operations.

## Before you start

You need:

- Python 3.13 or newer with the [`mergeway` package installed](../install/)
- `mergeway-cli` available on `PATH`
- an existing repository with a `mergeway.yaml` file

The mutation examples change repository files. Run them on a disposable branch and replace the example entity, fields, and identifiers with values valid for your schema.

## Generate Python models

Generate a module from a repository root:

```sh
mergeway-python generate \
  --root /path/to/mergeway-repo \
  /path/to/mergeway-repo/generated_models.py
```

To select a configuration file explicitly, use `--config` instead:

```sh
mergeway-python generate \
  --config /path/to/mergeway-repo/mergeway.yaml \
  /path/to/mergeway-repo/generated_models.py
```

The generated module contains one dataclass per entity, nested dataclasses for object fields, enums for native enum fields, and an `ENTITY_REGISTRY` used by the SDK.

## Connect to the repository

Construct a `Database` with the configuration and generated module paths:

```python
from pathlib import Path

from mergeway import Database

repo_root = Path("/path/to/mergeway-repo")
db = Database(
    repo_root / "mergeway.yaml",
    classes_module=repo_root / "generated_models.py",
)
```

Without `classes_module`, the same operations work but reads return dictionaries instead of generated model instances. Use `cli_binary="/path/to/mergeway-cli"` if the binary is not on `PATH`.

## Read repository data

Fetch one record, list an entity, or export selected data:

```python
post = db.get("Post", "post-001")
print(post.title)
print(post.to_payload())

for user in db.list("User"):
    print(user.id, user.email)

for post in db.export("Post"):
    print(post.id, post.status)
```

Calling `db.export()` without entity names returns a dictionary keyed by entity name.

## Create, update, and delete records

Use a generated class for typed creation:

```python
User = db.classes_module.User

created = db.create(
    User,
    User(
        id="user-cara",
        name="Cara Example",
        email="cara@example.com",
        roles=["author"],
    ),
)
print(created.to_payload())
```

You can also pass an entity name and plain dictionary. An update replaces the stored record by default; set `merge=True` to merge only the supplied fields:

```python
updated = db.update(
    "User",
    "user-cara",
    {"roles": ["author", "editor"]},
    merge=True,
)
print(updated.to_payload())

message = db.delete("User", "user-cara")
print(message)
```

These methods inherit the underlying CLI's validation, external-path, confirmation, and file-layout behavior. Review the [`create`](../../../reference/cli/create/), [`update`](../../../reference/cli/update/), and [`delete`](../../../reference/cli/delete/) references before automating destructive changes.

## Validate, format, and export

Run repository-level operations through the same `Database`:

```python
result = db.validate()
print(result)

db.validate(phases=["schema", "references"])
db.format(in_place=True)

snapshot = db.export()
print(snapshot["User"])
```

To check formatting without rewriting files, call:

```python
db.format(lint=True, in_place=False)
```

## Generate models from Python

Model generation is also available through the API:

```python
from pathlib import Path

from mergeway import Database

repo_root = Path("/path/to/mergeway-repo")
db = Database(repo_root / "mergeway.yaml")
generated_path = db.generate_classes(repo_root / "generated_models.py")
print(generated_path)
```

The `Database` loads the new module automatically, so later reads can return typed objects. Commit generated code only if your project's generated-file policy calls for it, and regenerate it when the repository schema changes.

For the SDK's ownership and compatibility boundaries, return to the [Python SDK component overview](../../../components/python/). For exact repository behavior, use the [Mergeway specification](../../../reference/specification/) and [`mergeway-cli` reference](../../../reference/cli/).

