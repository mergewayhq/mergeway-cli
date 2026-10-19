# Documentation migration ledger

This ledger records the Phase 10 disposition of every file beneath `docs/src`, the relevant `docs/arch` sources, and the legacy style guide. “Move” means one canonical Hugo destination supersedes the source. “Split” means distinct durable concerns moved to named destinations. “Merge” means useful content was reconciled into existing canonical material. “Archive” means the file remains only in the recoverable mdBook history because it is generated navigation, an asset already replaced by an accessible source-backed rendering, or stale/proposed behavior.

The mdBook files are not deleted during the transition. This ledger assigns ownership; production removal remains gated by the cutover and observation phases.

| Legacy source | Disposition | Canonical destination or archive rationale |
| --- | --- | --- |
| `docs/src/README.md` | Move | `website/content/docs/_index.md` |
| `docs/src/SUMMARY.md` | Archive | mdBook-only generated navigation; Hugo derives navigation from the content tree. |
| `docs/src/getting-started/README.md` | Merge | `website/content/docs/tutorials/first-repository.md` and `website/content/docs/explanation/repository-model.md` |
| `docs/src/getting-started/installation.md` | Move | `website/content/docs/how-to/install.md` |
| `docs/src/getting-started/language-server.md` | Split | `website/content/docs/reference/lsp/_index.md` and `website/content/docs/how-to/lsp/debug-logs.md` |
| `docs/src/getting-started/schema-spec.md` | Merge | `website/content/docs/reference/specification/` after reconciliation with released code/tests. |
| `docs/src/getting-started/workspace-setup.md` | Move | `website/content/docs/tutorials/first-repository.md` |
| `docs/src/guides/README.md` | Move | `website/content/docs/how-to/_index.md` |
| `docs/src/guides/communicate-changes-with-diff.md` | Split | `website/content/docs/tutorials/review-semantic-change.md` and `website/content/docs/how-to/diff/` |
| `docs/src/guides/example-repo.md` | Merge | `website/content/docs/tutorials/first-repository.md` using `examples/full`. |
| `docs/src/guides/json-schema.md` | Split | `website/content/docs/how-to/repositories/reuse-json-schema.md` and `website/content/docs/how-to/repositories/load-jsonpath.md` |
| `docs/src/guides/setup-mergeway-github.md` | Move | `website/content/docs/how-to/cli/github-actions.md` |
| `docs/src/guides/setup-mergeway-lsp-editors.md` | Split | `website/content/docs/how-to/lsp/neovim.md`, `website/content/docs/how-to/lsp/generic-client.md`, and the gated VS Code component page. |
| `docs/src/guides/setup-mergeway-pre-commit.md` | Move | `website/content/docs/how-to/cli/pre-commit.md` |
| `docs/src/guides/use-mergeway-mcp-with-claude-and-codex.md` | Split | `website/content/docs/tutorials/inspect-with-ai-client.md` and `website/content/docs/how-to/mcp/` |
| `docs/src/cli-reference/README.md` | Move | `website/content/docs/reference/cli/_index.md` |
| `docs/src/cli-reference/config-export.md` | Move | `website/content/docs/reference/cli/config-export.md` |
| `docs/src/cli-reference/config-lint.md` | Move | `website/content/docs/reference/cli/config-lint.md` |
| `docs/src/cli-reference/create.md` | Move | `website/content/docs/reference/cli/create.md` |
| `docs/src/cli-reference/delete.md` | Move | `website/content/docs/reference/cli/delete.md` |
| `docs/src/cli-reference/diff.md` | Move | `website/content/docs/reference/diff/_index.md` |
| `docs/src/cli-reference/entity-list.md` | Move | `website/content/docs/reference/cli/entity-list.md` |
| `docs/src/cli-reference/entity-show.md` | Move | `website/content/docs/reference/cli/entity-show.md` |
| `docs/src/cli-reference/export.md` | Move | `website/content/docs/reference/cli/export.md` |
| `docs/src/cli-reference/files.md` | Move | `website/content/docs/reference/cli/files.md` |
| `docs/src/cli-reference/fmt.md` | Move | `website/content/docs/reference/cli/fmt.md` |
| `docs/src/cli-reference/gen-erd.md` | Move | `website/content/docs/reference/cli/gen-erd.md` |
| `docs/src/cli-reference/get.md` | Move | `website/content/docs/reference/cli/get.md` |
| `docs/src/cli-reference/init.md` | Move | `website/content/docs/reference/cli/init.md` |
| `docs/src/cli-reference/list.md` | Move | `website/content/docs/reference/cli/list.md` |
| `docs/src/cli-reference/lsp.md` | Move | `website/content/docs/reference/lsp/_index.md` |
| `docs/src/cli-reference/mcp.md` | Split | `website/content/docs/reference/mcp/_index.md`, `transports.md`, and `tools.md` |
| `docs/src/cli-reference/update.md` | Move | `website/content/docs/reference/cli/update.md` |
| `docs/src/cli-reference/validate.md` | Move | `website/content/docs/reference/cli/validate.md` |
| `docs/src/cli-reference/version.md` | Move | `website/content/docs/reference/cli/version.md` |
| `docs/src/cli-reference/assets/erd-example-full.png` | Archive | Raster CLI screenshot replaced by generated ERD examples with descriptive alt text; retain only with mdBook rollback history. |
| `docs/arch/cli-behavior.md` | Merge | Released behavior is reconciled into `website/content/docs/reference/cli/`; stale optional proposals remain historical. |
| `docs/arch/diff-merge-readiness.md` | Merge | Durable semantics moved to `website/content/docs/reference/diff/_index.md` and `website/content/docs/explanation/semantic-diff.md`. |
| `docs/arch/go-cli-build-test.md` | Merge | Current contributor workflow is maintained in `DEVELOPING.md`; stale platform/tooling proposals remain historical. |
| `docs/arch/good-documentation.md` | Merge | Durable guidance consolidated into `website/content/docs/project/documentation.md`. |
| `docs/arch/schema-spec.md` | Merge | Released semantics reconciled into `website/content/docs/reference/specification/`; stale syntax is not copied. |
| `docs/arch/storage-layout.md` | Split | `website/content/docs/explanation/repository-model.md`, `schema-design.md`, and `website/content/docs/how-to/cli/split-files.md` |
| `docs/STYLEGUIDE.md` | Merge | Hugo and Diátaxis conventions consolidated into `website/content/docs/project/documentation.md`. |

## Explicitly deferred content

- **Scaling a repository (P2):** deferred until measured operational experience can support useful repository-size, latency, memory, and file-granularity guidance. Implementation structure alone is not evidence for capacity claims.
- **VS Code visual-editor workflows:** remain gated by Phase 9's unmet source-merge entry condition; they are not reclassified as Phase 10 content.
