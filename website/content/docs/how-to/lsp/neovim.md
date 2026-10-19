---
title: Configure Neovim
description: Start mergeway-lsp for YAML and JSON files owned by a Mergeway repository.
type: docs
weight: 10
bookToC: true
---

# Configure Neovim

This setup uses Neovim's built-in LSP client. It starts one `mergeway-lsp` process for files below the nearest directory containing `mergeway.yaml` or `mergeway.yml`.

## Prerequisites

- Install `mergeway-lsp` and confirm `vim.fn.exepath("mergeway-lsp")` is non-empty.
- Use a Neovim release that provides `vim.lsp.start` and `vim.fs.root`.
- Open the repository directory, not an individual unrelated file.

## Add the client configuration

Put this in `init.lua` or a sourced Lua module:

```lua
vim.api.nvim_create_autocmd("FileType", {
  pattern = { "yaml", "json" },
  callback = function(args)
    local root = vim.fs.root(args.buf, { "mergeway.yaml", "mergeway.yml" })
    if not root then
      return
    end

    vim.lsp.start({
      name = "mergeway-lsp",
      cmd = { "mergeway-lsp" },
      root_dir = root,
    })
  end,
})
```

Restart Neovim and open a configured data file such as `data/users/alice.yaml`. Run `:LspInfo`; the attached client should be named `mergeway-lsp`.

## Verify diagnostics

Temporarily replace a required string with a number, for example `name: 7`. After the short reload debounce, Neovim should display a Mergeway error. Undo the edit and confirm the diagnostic clears. The unsaved buffer is used; saving is not required.

Completion, hover, definition, references, symbols, and quick fixes are exposed through normal Neovim LSP commands and mappings. If the client attaches but a feature is empty, follow [Troubleshoot diagnostics and navigation](../troubleshoot/).

For several independent repositories in one Neovim session, the root callback starts or reuses a client per detected root. A client that instead sends all folders in one `workspaceFolders` array may use the server's static multi-root support.
