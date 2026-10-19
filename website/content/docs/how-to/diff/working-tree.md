---
title: Compare the working tree with HEAD
description: Choose whether a semantic review includes unstaged changes only or the complete current working tree.
type: docs
weight: 10
bookToC: true
---

# Compare the working tree with `HEAD`

Run validation and inspect Git state first:

```sh
mergeway-cli validate
git status --short
```

To review unstaged tracked changes and untracked data while excluding staged-only content, use zero arguments:

```sh
mergeway-diff
```

This mode compares `HEAD` with an unstaged working-tree view. It is useful while continuing to edit after staging a known baseline. Untracked, non-ignored records matching configured includes are part of the view.

To review everything currently on disk against `HEAD`, pass `HEAD` explicitly:

```sh
mergeway-diff HEAD
```

The one-revision mode includes staged, unstaged, and untracked files. Prefer it before committing because it matches the complete state that `mergeway-cli validate` just read.

Neither command modifies files. Both exit `0` whether changes exist or not; use [JSON output](../json-automation/) when automation must distinguish an empty result. The [reference snapshot table](../../../reference/diff/#invocation-and-snapshot-modes) covers combined staged and unstaged edits precisely.
