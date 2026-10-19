---
title: mergeway-cli validate
description: Validate repository formats, schemas, and references.
weight: 20
aliases: [/cli-reference/validate.html, /cli-reference/validate/]
---
# `mergeway-cli validate`

## Synopsis
```sh
mergeway-cli [global flags] validate [--phase format|schema|references]...
```

`--phase` is repeatable; with no phase, the order is format, schema, references. Requesting references also runs schema. The inherited `--fail-fast` default is `false`; `--format` defaults to YAML. There are no arguments.

## Output, effects, and exits

Success writes `status: validation succeeded`. Failures write structured errors with phase, entity, identifier, file, and message where available. It never changes repository files. A valid repository exits `0`; any validation or loading error exits `1`.

## Example
```sh
mergeway-cli --root examples/full --format json validate --phase schema --phase references
```

See [configuration versus repository validation](../../../how-to/cli/choose-validation/) and [`config lint`](../config-lint/).
