---
title: Semantic diff versus file diff
description: Understand how Mergeway compares logical records across repository snapshots rather than changed lines.
type: docs
weight: 40
bookToC: true
---

# Semantic diff versus file diff

A Git file diff answers “which lines changed?” A Mergeway semantic diff answers “which records changed, and how?” Both views are useful, but they describe different layers of a repository.

## Logical records outlive files

Mergeway loads YAML, JSON, selected arrays, and other configured sources into one logical database. An entity name and identifier—such as `Post[post-001]`—form a record's identity. Its source file is provenance, not identity.

That distinction matters when storage changes without changing the model. Moving an unchanged record from one file to another is a relocation. Reformatting YAML may produce a Git hunk but no semantic change. Conversely, one edit to a JSON array can become several added, modified, or removed record entries.

## A snapshot is a logical database at one state

Before comparison, Diff resolves each side to a snapshot. A snapshot can be a committed Git revision or a selected view of files in the working tree. It loads that state using the configuration present in that state, then normalizes every included record.

The invocation chooses the pair:

- no revisions: `HEAD` and the unstaged working-tree view;
- one revision: that commit and the complete working tree on disk;
- two revisions: two committed states.

The [Diff reference](../../reference/diff/) defines the exact staged, unstaged, untracked, and committed boundaries.

## Semantic equality is not textual equality

Maps are compared by key rather than source order. Repeated values are compared without regard to order. Nested maps produce focused paths such as `profile.name`. Values use normalized decoded data, so quoting or indentation alone does not constitute a record change.

This makes the output useful to a reviewer asking whether data meaning changed. It also means Diff is not a replacement for `git diff`: use Git when whitespace, comments, key ordering, configuration text, or an exact patch matters.

## Loading is not full validation

Diff must parse configuration and data and derive unique logical identities, so malformed data can stop a comparison. It does not promise the full schema, format, and relationship checks of [`mergeway-cli validate`](../validation-model/). A dependable review flow is therefore:

1. validate the intended repository state;
2. inspect the semantic diff for record meaning;
3. inspect the Git diff when file-level provenance matters.

Try the flow in [Review your first semantic change](../../tutorials/review-semantic-change/) or use the [working-tree comparison guide](../../how-to/diff/working-tree/).
