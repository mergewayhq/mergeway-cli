---
title: mergeway-cli reference
description: Invocation, global behavior, and released command reference for mergeway-cli v0.10.0.
type: docs
weight: 20
aliases:
  - /cli-reference/index.html
  - /cli-reference/
  - /cli-reference/README.html
params:
    bookCollapseSection: true
---

# `mergeway-cli` reference

This reference describes `mergeway-cli` v0.10.0, which operates on repositories using `mergeway.version: 1`.

## Invocation and repository discovery

```sh
mergeway-cli [global flags] <command> [arguments] [local flags]
```

Long flags use two dashes and may appear before or after the command. `--root` defaults to the current directory. Unless `--config` is set, the entry configuration is `<root>/mergeway.yaml`; paths in configuration use the resolution rules in the [repository specification](../specification/repository-configuration/). The CLI does not search parent directories.

## Global flags

| Flag | Default | Effect |
| --- | --- | --- |
| `--root <path>` | `.` | Repository root containing configuration and data. |
| `--config <path>` | `<root>/mergeway.yaml` | Configuration entry file. |
| `--format yaml\|json` | `yaml` | Structured output encoding where a command supports it. |
| `--fail-fast` | `false` | Stop `validate` after the first error. |
| `--yes` | `false` | Confirm destructive prompts, notably `delete`. |
| `--verbose` | `false` | Enable additional diagnostic logging where emitted. |
| `-h`, `--help` | — | Print help and exit successfully. |

## Output, exits, and side effects

Structured results use standard output; diagnostics and prompts use standard error. `--format` affects structured results from `validate`, `version`, `entity show`, `config export`, `files`, `get`, and `export`, but not plain identifier or mutation-status lines. Success exits `0`. Invalid arguments, loading or validation errors, failed writes, rejected confirmation, and `fmt --lint` differences exit `1`. In particular, a repository containing validation errors makes `validate` exit `1` after writing those errors.

Read-only commands are `validate`, `version`, `entity list`, `entity show`, `config lint`, `config export`, `list`, `files`, `get`, and `export` (unless `export --output` creates or truncates a destination). `init`, default/`--in-place` `fmt`, `create`, `update`, and `delete` modify files. `gen-erd` creates or replaces its output image through Graphviz.

## Commands

- Repository: [`init`](init/), [`validate`](validate/), [`version`](version/).
- Schema: [`entity list`](entity-list/), [`entity show`](entity-show/), [`fmt`](fmt/), [`config lint`](config-lint/), [`config export`](config-export/), [`gen-erd`](gen-erd/).
- Records: [`list`](list/), [`files`](files/), [`get`](get/), [`create`](create/), [`update`](update/), [`delete`](delete/), [`export`](export/).

Every command inherits the global flags above. `completion` and `help` are Cobra utility commands, not Mergeway repository operations, and are therefore intentionally outside this released operation inventory.

## Related documentation

Start with the [record lifecycle guide](../../how-to/cli/manage-records/) for a task-oriented workflow, or consult the [CLI component overview](../../components/cli/) and [normative specification](../specification/).
