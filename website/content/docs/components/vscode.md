---
title: VS Code
description: Connect Visual Studio Code to Mergeway language features through the local LSP server.
type: docs
weight: 60
bookToC: true
---

# VS Code

The Mergeway VS Code extension connects Visual Studio Code to `mergeway-lsp`. It is the suite's editor-specific integration and depends on the language server for diagnostics and navigation.

## Who it is for

VS Code users are the intended audience for the extension. Extension contributors can build and test it from this repository.

## Purpose and boundaries

The extension connects editor documents and user actions to Mergeway language features. It does not define repository behavior and does not bundle the language server. The CLI remains the explicit final validation and automation interface.

## Relationship to the suite

The extension launches `mergeway-lsp`, which shares the specification and validation behavior used by the CLI. It is the editor-specific layer, while LSP is the reusable protocol layer.

## Availability and compatibility

The extension source is in this repository and contributors can build a local VSIX. There is no documented public Marketplace or Open VSX distribution channel, so this site does not present contributor packaging as an end-user installation route. A locally built VSIX also requires a separately installed `mergeway-lsp` and configuration of its path.

## Start here

- Read the [LSP component overview](../lsp/) to understand the required server.
- Learn [the repository model](../../explanation/repository-model/) and [validation model](../../explanation/validation-model/) that drive editor feedback.
- Use [the CLI tutorial](../../tutorials/first-repository/) for the supported first-use path today.

Public installation, editing, settings, commands, and troubleshooting documentation wait for the reviewed extension distribution route and the VS Code documentation phase.

Contributors running the source-built LSP client can use the focused [extension troubleshooting notes](https://github.com/mergewayhq/mergeway-cli/blob/main/vscode/README.md#troubleshooting). They are contributor instructions, not a public installation claim.
