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
- Follow [Review your first semantic change](../../tutorials/review-semantic-change/) for a complete guided review.
- Read [Semantic diff versus file diff](../../explanation/semantic-diff/) to understand logical records and snapshots.
- Look up the complete [`mergeway-diff` reference](../../reference/diff/).

## Common tasks

- [Compare the working tree with `HEAD`](../../how-to/diff/working-tree/).
- [Compare two Git revisions](../../how-to/diff/revisions/).
- [Publish a semantic change summary](../../how-to/diff/publish-summary/).
- [Consume JSON output in automation](../../how-to/diff/json-automation/).

Validate a repository with [`mergeway-cli validate`](../../reference/cli/validate/) before publishing a diff. Diff parses repository data so it can compare records, but it does not perform the full schema and relationship validation contract.
