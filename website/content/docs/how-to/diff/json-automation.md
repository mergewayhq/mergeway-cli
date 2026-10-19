---
title: Consume JSON diff output in automation
description: Parse the versioned mergeway-diff JSON document and handle empty results and errors correctly.
type: docs
weight: 40
bookToC: true
---

# Consume JSON diff output in automation

Capture a committed comparison as JSON:

```sh
mergeway-diff --format json HEAD~1 HEAD > semantic-diff.json
```

Check the schema version before consuming entries. With `jq`:

```sh
jq -e '.version == 1' semantic-diff.json >/dev/null
jq -r '.entries[] | [.kind, .type, .object_id] | @tsv' semantic-diff.json
```

Diff exits `0` for both empty and non-empty comparisons. Test the array when a workflow needs to branch:

```sh
if jq -e '.entries | length > 0' semantic-diff.json >/dev/null; then
  echo "Semantic changes found"
else
  echo "No semantic changes"
fi
```

Keep command failure separate from a non-empty result:

```sh
if ! mergeway-diff --format json HEAD~1 HEAD > semantic-diff.json; then
  echo "Diff could not load or compare the repository" >&2
  exit 1
fi
```

On failure, the command writes a categorized diagnostic to standard error and does not emit a partial JSON document. Consumers should reject a top-level version they do not support, tolerate additional fields, and switch on `kind` before reading kind-specific fields. The [JSON schema reference](../../../reference/diff/#json-output-schema) defines `value`, `changes`, and source arrays.
