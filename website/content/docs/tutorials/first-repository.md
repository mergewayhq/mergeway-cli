---
title: Build your first Mergeway repository
description: Install one released CLI, explore a fixed blog repository, create a related record, validate it, and export the result.
type: docs
mode: tutorial
weight: 10
bookToC: true
aliases:
  - /getting-started/workspace-setup.html
  - /getting-started/workspace-setup/
  - /guides/example-repo.html
  - /guides/example-repo/
---

# Build your first Mergeway repository

In this tutorial you will install `mergeway-cli` v0.10.0 for Linux x86_64, copy the tag-matched `examples/full` fixture, inspect its entities and records, create a blog post related to existing records, validate the repository, and export it as JSON.

The commands work in a temporary learning directory. They do not modify the source fixture.

## Before you begin

You need a Linux x86_64 shell with `curl` and `tar`. For another supported platform, follow [Install Mergeway tools](../../how-to/install/) and return after `mergeway-cli version` works.

## 1. Install the fixed release

Create a learning directory and install the CLI locally inside it:

```sh
mkdir mergeway-first-repository
cd mergeway-first-repository
mkdir bin
curl -fLO https://github.com/mergewayhq/mergeway-cli/releases/download/v0.10.0/mergeway-cli_Linux_x86_64.tar.gz
tar -xzf mergeway-cli_Linux_x86_64.tar.gz
install -m 0755 mergeway-cli bin/mergeway-cli
export PATH="$PWD/bin:$PATH"
mergeway-cli version
```

The version output begins with `version: 0.10.0-dev` for the published v0.10.0 archive and also identifies its release commit. If the shell cannot find the command, confirm that you ran the `export PATH` line in the same shell.

## 2. Copy the fixed starting repository

Download the source archive for the same tag and copy its full example:

```sh
curl -fL https://github.com/mergewayhq/mergeway-cli/archive/refs/tags/v0.10.0.tar.gz -o source.tar.gz
tar -xzf source.tar.gz
cp -R mergeway-cli-0.10.0/examples/full repository
cd repository
```

Your working copy now has this shape:

```text
repository/
├── mergeway.yaml
├── entities/
│   ├── Comment.yaml
│   ├── Post.yaml
│   ├── Tag.yaml
│   └── User.yaml
└── data/
    ├── comments/
    ├── posts/
    ├── tags/
    └── users/
```

`mergeway.yaml` is the repository entry point. It includes the four entity schemas from `entities/`; those schemas in turn locate YAML and JSON records under `data/`.

## 3. Inspect the model and records

List the entities:

```sh
mergeway-cli entity list
```

Expected output:

```text
Comment
Post
Tag
User
```

Now list the `User` record identifiers and inspect the existing post:

```sh
mergeway-cli list --type User
mergeway-cli get --type Post post-001
```

The first command prints `user-alice` and `user-bob`. The post output includes `author: user-alice` and `tag-product`: these values form relationships from the `Post` record to `User` and `Tag` records.

## 4. Create a related post

Create a payload outside the repository, then ask the CLI to store it using the `Post` entity's configured include path:

```sh
cat > ../new-post.yaml <<'YAML'
id: post-002
title: Documentation Tour
status: DRAFT
author: user-bob
tags:
  - tag-update
body: |
  Learning Mergeway one record at a time.
YAML

mergeway-cli create --type Post --file ../new-post.yaml
```

Expected output:

```text
Post post-002 created
```

The new `data/posts/post-002.yaml` record relates to the existing `User` record `user-bob` and `Tag` record `tag-update`. If the CLI reports that the record already exists, remove this learning directory and begin again from step 1.

## 5. Validate the repository

```sh
mergeway-cli validate
```

Expected output:

```text
status: validation succeeded
```

Validation checks that files parse, records satisfy their entity schemas, identifiers are unique, and the `author` and `tags` relationships resolve. See [The validation model](../../explanation/validation-model/) for how these checks fit together.

## 6. Export the result

```sh
mergeway-cli --format json export --output ../repository.json
grep -F '"id": "post-002"' ../repository.json
```

The `grep` command prints the exported identifier:

```text
      "id": "post-002",
```

The export contains all four entities and their records in one JSON document. Your source files remain the repository of record; the export is a snapshot for another tool.

## What you built

You started from a fixed file-based database, found its entities and records, added a record with two valid relationships, validated the whole repository, and exported a machine-readable snapshot.

Next, read [Why metadata as code?](../../explanation/metadata-as-code/) for the collaboration model, [The repository model](../../explanation/repository-model/) for vocabulary, or the [CLI component overview](../../components/cli/) to choose another operation.
