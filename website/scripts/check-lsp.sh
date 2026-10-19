#!/bin/sh

set -eu

repo_root=$(CDPATH= cd -- "$(dirname -- "$0")/../.." && pwd)
reference=$repo_root/website/content/docs/reference/lsp/_index.md
tmp_dir=$(mktemp -d)
trap 'rm -rf "$tmp_dir"' EXIT HUP INT TERM
lsp_bin=$tmp_dir/mergeway-lsp

go build -o "$lsp_bin" "$repo_root/cmd/mergeway-lsp"

if "$lsp_bin" --help > "$tmp_dir/help.stdout" 2> "$tmp_dir/help.stderr"; then
  echo 'mergeway-lsp --help unexpectedly exited 0; released behavior is exit 2' >&2
  exit 1
else
  status=$?
fi
test "$status" -eq 2
test ! -s "$tmp_dir/help.stdout"
for entry in \
  '--log-file|Write LSP logs to a file' \
  '--log-level|Log level (debug|info|warn|error)' \
  '--log-stderr|Write LSP logs to stderr'
do
  flag=${entry%%|*}
  help_text=${entry#*|}
  grep -Fq -- "$flag" "$reference"
  grep -Fq -- "$help_text" "$tmp_dir/help.stderr"
done

for variable in MERGEWAY_LSP_LOG_FILE MERGEWAY_LSP_LOG_LEVEL MERGEWAY_LSP_LOG_STDERR; do
  grep -Fq -- "$variable" "$reference"
done

go test ./cmd/mergeway-lsp ./internal/lsp ./internal/workspace \
  -run 'Test(RunHelpUsesStderrAndLeavesStdoutProtocolOnly|NewLogger|RunInitializeShutdownExit|RunStdioTranscriptUsesProtocolFramesOnly|HandleInitializeBuildsRootsFromWorkspaceFolders|HandleInitializeSupportsNoConfigRootURI|OpenRootsDetectsMissingAndLoadsIndexes|OpenRootsKeepsMultiRootIndexesIsolated|RootIndexOwnsOnlyConfiguredFiles|HandleInitializePublishesSavedWorkspaceDiagnosticsParity|HandleCompletionProvidesContextAwareSuggestions|HandleHoverReturnsSchemaAndEntitySummaries|HandleDefinitionResolvesSavedAndUnsavedReferences|HandleReferencesUsesSemanticTargets|HandleDocumentSymbolMapsEntitiesAndFields|HandleWorkspaceSymbolFindsByNameAndKeepsRootsSeparate|HandleCodeActionOffersConservativeQuickFixes)' \
  -count=1

go run "$repo_root/website/scripts/lsp-example-client.go" \
  --server "$lsp_bin" \
  --root "$repo_root/internal/validation/testdata/schema_error" > "$tmp_dir/client.out"

grep -Fq 'server=mergeway-lsp' "$tmp_dir/client.out"
grep -Fq 'sync=full workspace-folders=static' "$tmp_dir/client.out"
grep -Fq 'capabilities=completion,hover,definition,references,document-symbols,workspace-symbols,code-actions' "$tmp_dir/client.out"
grep -Fq 'diagnostic-codes=schema' "$tmp_dir/client.out"

printf 'Verified mergeway-lsp help and environment precedence, static single/multi-root ownership, advertised capabilities, protocol-only stdio, validation parity, and a real-client diagnostic session.\n'
