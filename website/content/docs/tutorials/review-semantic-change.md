---
title: Review your first semantic change
description: Modify, add, and remove records in a temporary Git copy of examples/full, then validate and interpret the semantic diff.
type: docs
weight: 20
bookToC: true
---

# Review your first semantic change

In this tutorial you will create a disposable Git repository from Mergeway's `examples/full`, make three kinds of record change, validate them, and review the result with `mergeway-diff`. Nothing writes to your source checkout.

## Prerequisites

- Git.
- `mergeway-cli` and `mergeway-diff` v0.10.0 on `PATH`; use [Install Mergeway tools](../../how-to/install/) if needed.
- A downloaded v0.10.0 source archive, which supplies the fixed `examples/full` starting state used by this tutorial.

## 1. Create a temporary repository

From the extracted source archive, copy the fixture and initialize only the copy:

```sh
TUTORIAL_DIR="$(mktemp -d)"
cp -R examples/full "$TUTORIAL_DIR/repository"
cd "$TUTORIAL_DIR/repository"
git init
git config user.name "Mergeway Tutorial"
git config user.email "tutorial@example.com"
git add .
git commit -m "Initial Mergeway repository"
```

`git status --short` should print nothing. If Git reports an unknown identity, repeat the two local `git config` commands; they intentionally do not change your global configuration.

## 2. Modify an existing record

Change the title in `data/posts/launch.yaml` from `Launch Day` to `Launch Day reviewed`. For example:

```sh
sed 's/title: Launch Day/title: Launch Day reviewed/' \
  data/posts/launch.yaml > data/posts/launch.yaml.new
mv data/posts/launch.yaml.new data/posts/launch.yaml
```

The logical identity remains `Post[post-001]`; only its `title` value changes.

## 3. Add and remove records

Create `data/posts/retrospective.yaml`:

```yaml
id: post-002
title: Launch retrospective
status: DRAFT
author: user-alice
tags:
  - tag-update
body: Follow-up notes from the launch.
```

Then remove `data/comments/launch-comment.json`:

```sh
rm data/comments/launch-comment.json
```

These changes add `Post[post-002]` and remove `Comment[comment-001]`. If your editor saved the new file elsewhere, move it to the exact path above so the configured `data/posts/*.yaml` include finds it.

## 4. Validate the complete working tree

```sh
mergeway-cli validate
```

Expected result:

```text
status: validation succeeded
```

Validation is a separate prerequisite: Diff loads enough data to compare records but does not replace full schema and relationship validation.

## 5. Compare the working tree with the baseline

Use one explicit revision so the right side is the complete working tree, including the untracked new post:

```sh
mergeway-diff HEAD
```

Expected semantic facts are:

```text
REMOVED Comment[comment-001]
  from: data/comments/launch-comment.json

MODIFIED Post[post-001]
  title: "Launch Day" -> "Launch Day reviewed"

ADDED Post[post-002]
  at: data/posts/retrospective.yaml
```

The executable output is ordered by entity and identifier, so these entries may not follow the order in which you edited the files. `REMOVED`, `MODIFIED`, and `ADDED` describe logical record identity, not Git line operations.

If the command prints `No changes.`, confirm that you are inside the temporary repository and that `git status --short` lists the three paths. If it reports a data error, rerun validation and fix the named file.

## 6. Contrast the file and semantic views

```sh
git diff -- data/posts/launch.yaml data/comments/launch-comment.json
git status --short
```

Git shows patches and identifies the untracked post separately. Diff loads all three paths through their entity definitions and presents one record-level review. Neither view replaces the other.

## 7. Clean up

Leave the temporary directory, then remove it when you no longer need it:

```sh
cd -
rm -rf "$TUTORIAL_DIR"
```

Only delete the directory created by `mktemp -d`. Continue with [Compare the working tree with `HEAD`](../../how-to/diff/working-tree/) for snapshot choices, or read [Semantic diff versus file diff](../../explanation/semantic-diff/) for the underlying model.
