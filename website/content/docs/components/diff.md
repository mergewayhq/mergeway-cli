---
title: Diff
description: Compare Mergeway repository states as records and relationships rather than raw text.
type: docs
weight: 30
bookToC: true
---

# Diff

`mergeway-diff` compares two repository snapshots and reports semantic record changes. It builds on the same repository model as the CLI so reviewers and automation can reason about added, modified, and removed data rather than only changed lines.

## Who it is for

Reviewers, repository maintainers, and CI authors use Diff when a text diff does not clearly express the logical data change.

## Purpose and boundaries

Diff compares committed revisions or a revision and working-tree state, then describes record-level additions, modifications, and removals. It does not modify records, replace repository validation, or serve as a general text-diff tool.

## Relationship to the suite

`mergeway-diff` loads the same repository model as the CLI. A useful review flow validates with `mergeway-cli` and then uses Diff to communicate semantic changes.

## Availability and compatibility

`mergeway-diff` is included in release archives for Linux, macOS, and Windows on `amd64` and `arm64`; it can also be installed with Go or built from source. Docker and the Nix flake do not provide it. Use binaries from the same release when combining suite components.

## Start here

- [Install Mergeway tools](../../how-to/install/).
- Learn [why metadata as code](../../explanation/metadata-as-code/) makes semantic review useful.
- Build and validate [the tutorial repository](../../tutorials/first-repository/) before comparing revisions.

Guided review workflows, output contracts, and complete command reference will be added in the Diff documentation phase.
