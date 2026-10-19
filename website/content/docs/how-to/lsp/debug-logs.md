---
title: Capture LSP debug logs
description: Enable mergeway-lsp debug logging safely on stderr or in a file.
type: docs
weight: 30
bookToC: true
---

# Capture LSP debug logs

Prefer a log file because an editor may hide its child process's standard error:

```sh
mergeway-lsp --log-file /tmp/mergeway-lsp.log --log-level debug
```

Configure those two arguments in the client launcher, reproduce the problem, and inspect the file. Existing content is retained because the server opens the file in append mode. The log can contain absolute paths and error details; review it before sharing and remove it when finished.

For a client that exposes child stderr, use:

```sh
mergeway-lsp --log-stderr --log-level debug
```

Never redirect stderr into stdout. Standard output must contain only framed JSON-RPC traffic.

Environment variables are useful when a client cannot supply arguments:

```sh
MERGEWAY_LSP_LOG_FILE=/tmp/mergeway-lsp.log \
MERGEWAY_LSP_LOG_LEVEL=debug \
mergeway-lsp
```

Flags override the matching non-empty file and level variables. `--log-stderr` and a truthy `MERGEWAY_LSP_LOG_STDERR` are additive, but a selected log file takes priority over stderr. Valid truthy values are `1`, `true`, `yes`, and `on`, case-insensitively. Logging defaults to `info`; without a file or stderr selection, logs are discarded.

If the process rejects an invalid level or cannot open the file, it writes `mergeway-lsp: ...` to stderr and exits `1`. Unknown flags exit `2`.
