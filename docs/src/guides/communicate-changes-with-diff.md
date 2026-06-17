---
title: "Communicate Repository Changes with mergeway-diff"
linkTitle: "Communicate changes"
description: "Use mergeway-diff with a time-based Git baseline to explain semantic data changes over time."
weight: 30
---

Goal: produce a shareable summary of Mergeway-managed data changes over time by comparing the current repository state to a commit selected with a Git time offset.

We will keep using the fictional GrainBox Market repository from the other guides. Product records live under `data/products/`, category lookups live under `data/categories/`, and the Data Platform team wants a weekly update that operations, merchandising, and leadership can all understand without reading raw Git file diffs.

## Prerequisites

- The repository is already initialized as a Git repository and contains a valid Mergeway workspace.
- `mergeway-diff` is installed and available on your `PATH`.
- You know the reporting window you want to communicate, such as the last 7 days.

## 1. Pick a Baseline Commit from a Time Offset

Use Git to find the most recent commit that existed before your reporting window started:

```bash
BASELINE="$(git log -1 --before="7 days ago" --format="%H")"
printf 'Baseline commit: %s\n' "$BASELINE"
git show --stat --oneline --no-patch "$BASELINE"
```

This command asks Git for the latest commit older than 7 days and stores its full commit identifier in `BASELINE`. That gives you a stable starting point for a weekly summary without manually searching through history.

If `BASELINE` is empty, your repository may not have any commits that old yet. In that case, shorten the time window or choose a specific revision manually.

## 2. Compare the Baseline to the Current Committed State

For a report you want to share broadly, compare two revisions explicitly:

```bash
mergeway-diff "$BASELINE" HEAD
```

This produces a semantic, data-only diff between the baseline commit and the current `HEAD` commit. Unlike a raw Git diff, `mergeway-diff` groups changes by Mergeway object identity and field values, so the result is easier to explain to non-specialists.

Example output:

```text
ADDED Product[prod-104]
  at: data/products/prod-104.yaml

MODIFIED Category[cat-bulk]
  description: "Warehouse items" -> "Warehouse and bulk items"

REMOVED Product[prod-017]
  from: data/products/prod-017.yaml
```

This is usually the right mode for weekly or monthly updates because it ignores local, uncommitted work and reports only the committed state of the repository.

## 3. Know When to Include the Working Tree

`mergeway-diff` has three useful snapshot modes:

- `mergeway-diff` compares `HEAD` against unstaged changes only.
- `mergeway-diff "$BASELINE"` compares the baseline revision against your full current working tree, including staged and unstaged changes.
- `mergeway-diff "$BASELINE" HEAD` compares two revisions and ignores local worktree noise.

For communication with a broader audience, prefer `mergeway-diff "$BASELINE" HEAD`. Use `mergeway-diff "$BASELINE"` only when you intentionally want to preview in-progress changes before they are committed.

## 4. Validate Before You Share the Diff

Run validation first so your report is based on a clean, parseable dataset:

```bash
mergeway-cli validate
mergeway-diff "$BASELINE" HEAD
```

If the workspace contains invalid YAML, duplicate identifiers, or broken references, fix those issues before publishing the summary. That keeps the diff focused on real data changes instead of repository problems.

## 5. Capture JSON Output for Automation

If you want to feed the result into a script, dashboard, or release-note generator, use JSON output:

```bash
mergeway-diff --format json "$BASELINE" HEAD > weekly-diff.json
```

This is useful when the Data Platform team wants to post the same weekly change set into Slack, an internal changelog, or a reporting pipeline.

## 6. Assemble a Shareable Weekly Update

You can combine the time-based baseline, the current commit, and the semantic diff into a short report file:

```bash
CURRENT="$(git rev-parse HEAD)"

{
  echo "# GrainBox weekly data update"
  echo
  echo "Baseline: $BASELINE"
  echo "Current:  $CURRENT"
  echo
  mergeway-diff "$BASELINE" HEAD
} > weekly-data-update.md
```

The resulting file gives stakeholders a compact summary of what changed over the last week without requiring them to inspect every modified YAML or JSON file.

## Next Steps

- Review the [`mergeway-diff` reference](../cli-reference/diff.md) for the exact snapshot rules and output modes.
- Pair this workflow with [`mergeway-cli validate`](../cli-reference/validate.md) in CI if you want every published change summary to come from a valid repository state.
- If you want teams to review the underlying changes before the summary is shared, combine this guide with [Set up Mergeway with GitHub Actions](setup-mergeway-github.md).
