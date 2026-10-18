---
title: VS Code
description: Connect Visual Studio Code to Mergeway language features through the local LSP server.
type: docs
weight: 60
bookToC: true
---

# VS Code

The Mergeway VS Code extension connects Visual Studio Code to `mergeway-lsp`. It is the suite's editor-specific integration and depends on the language server for diagnostics and navigation.

## Availability

The extension source is included in this repository and can be built into a local VSIX. There is not yet a documented public extension distribution channel, and the VSIX does not bundle `mergeway-lsp`; users must install or build the server and configure its path. This site does not yet present local contributor packaging as an end-user installation path.

## Planned documentation

After a public distribution route and the editor surface are ready for review, later phases will add installation, editing, diagnostics, navigation, settings, commands, and troubleshooting documentation.
