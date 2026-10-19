---
title: Compare two Git revisions
description: Produce a stable semantic report between two committed repository states.
type: docs
weight: 20
bookToC: true
---

# Compare two Git revisions

Resolve the comparison endpoints, then pass both to Diff:

```sh
LEFT="$(git rev-parse HEAD~1)"
RIGHT="$(git rev-parse HEAD)"
mergeway-diff "$LEFT" "$RIGHT"
```

Both arguments must resolve to commits. Tags, branches, hashes, and Git revision expressions work. The command reads configuration and data from those commits and ignores every current staged, unstaged, and untracked file, making the result reproducible from the same repository history.

To compare a pull-request range locally, use the merge base rather than assuming the branch point:

```sh
BASE="$(git merge-base origin/main HEAD)"
mergeway-diff "$BASE" HEAD
```

Run `mergeway-cli validate` on the checked-out state when it is the right endpoint. Validating an arbitrary historical endpoint requires checking it out in a separate temporary worktree; Diff itself is read-only but is not a full validator.

Use [Publish a semantic change summary](../publish-summary/) to package the report, or the [`mergeway-diff` reference](../../../reference/diff/) for errors and limitations.
