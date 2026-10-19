---
title: Contributing
description: Find the stable contribution workflow and detailed repository development instructions.
weight: 40
type: docs
mode: how-to
bookToC: true
---

# Contributing

Mergeway accepts changes through its public GitHub repository. Before starting a substantial behavior change, open or find an issue so the intended specification and compatibility impact are visible.

## Prepare a change

1. Enter the pinned development environment with `devenv shell`.
2. Make a focused change with tests at the layer that owns the behavior.
3. Run the relevant focused checks, then `make fmt-check` and `go test ./...`.
4. Run `make website-check` when user behavior, examples, commands, protocols, or documentation change.
5. Explain the user-documentation impact in the pull request, including “no documentation impact” when appropriate.

Detailed prerequisites, package ownership, build/debug commands, release mechanics, and the temporary mdBook/Hugo dual-site workflow live in [`DEVELOPING.md`](https://github.com/mergewayhq/mergeway-cli/blob/main/DEVELOPING.md). Keeping volatile commands there lets this page remain a concise public entry point.

## Documentation contributions

Use [Documentation authoring](../documentation/) for page modes, front matter, links, and validation. User-facing claims must describe released behavior or name the fixture and unreleased boundary explicitly. Do not remove legacy mdBook content until the cutover and observation phases authorize it.

For the codebase's component boundaries, see [Component architecture](../../explanation/component-architecture/).
