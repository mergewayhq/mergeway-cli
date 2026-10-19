---
title: Publish a semantic change summary
description: Turn a dated Git baseline and committed semantic diff into a shareable Markdown report.
type: docs
weight: 30
bookToC: true
aliases:
  - /guides/communicate-changes-with-diff.html
  - /guides/communicate-changes-with-diff/
---

# Publish a semantic change summary

Choose the most recent commit before the reporting window and verify that it exists:

```sh
BASELINE="$(git log -1 --before="7 days ago" --format="%H")"
test -n "$BASELINE" || { echo "No baseline commit found" >&2; exit 1; }
CURRENT="$(git rev-parse HEAD)"
```

Compare two explicit revisions so local work cannot leak into a published report:

```sh
mergeway-diff "$BASELINE" "$CURRENT"
```

Then assemble the result with its reproducible endpoints:

```sh
{
  echo "# Weekly semantic data update"
  echo
  echo "Baseline: $BASELINE"
  echo "Current:  $CURRENT"
  echo
  echo '```text'
  mergeway-diff "$BASELINE" "$CURRENT"
  echo '```'
} > weekly-data-update.md
```

Review `weekly-data-update.md` before publishing it. The report describes record additions, removals, modifications, and relocations; it deliberately omits commit authorship and configuration-only edits. Pair it with your normal Git review when those details matter.

For a pipeline-friendly document, use the [JSON automation guide](../json-automation/). See [Compare two Git revisions](../revisions/) for merge-base selection.
