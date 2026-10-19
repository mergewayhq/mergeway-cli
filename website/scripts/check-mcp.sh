#!/bin/sh

set -eu

repo_root=$(CDPATH= cd -- "$(dirname -- "$0")/../.." && pwd)
reference=$repo_root/website/content/docs/reference/mcp/_index.md
tools_reference=$repo_root/website/content/docs/reference/mcp/tools.md
transport_reference=$repo_root/website/content/docs/reference/mcp/transports.md
expected=$repo_root/website/testdata/mcp/examples.json
tmp_dir=$(mktemp -d)
trap 'rm -rf "$tmp_dir"' EXIT HUP INT TERM
mcp_bin=$tmp_dir/mergeway-mcp

go build -o "$mcp_bin" "$repo_root/cmd/mergeway-mcp"

"$mcp_bin" --help > "$tmp_dir/help.stdout" 2> "$tmp_dir/help.stderr"
test ! -s "$tmp_dir/help.stdout"
for entry in \
  '--root|Root folder of the Mergeway repository (default ".")' \
  '--transport|Transport protocol (stdio|http) (default "stdio")' \
  '--http-listen|Listen address for HTTP transport (default "127.0.0.1:8080")' \
  '--http-base-path|Base path for HTTP transport (default "/")' \
  '--entity|Allow only the named entity (repeatable)'
do
  flag=${entry%%|*}
  help_text=${entry#*|}
  grep -Fq -- "$flag" "$reference"
  grep -Fq -- "$help_text" "$tmp_dir/help.stderr"
done

for tool in entity_list entity_show object_list object_get repository_export files_list; do
  grep -Fq -- "## \`$tool\`" "$tools_reference"
done

go run "$repo_root/website/scripts/mcp-example-client.go" \
  --server "$mcp_bin" --root "$repo_root/examples/full" > "$tmp_dir/examples.json"
cmp "$expected" "$tmp_dir/examples.json"

go test ./internal/mcp ./internal/mcpcmd \
  -run 'Test(NewServerRegistersInitialReadOnlyTools|ServerExposesOnlyReadOnlyToolSet|ServerReturnsProtocolErrorsForUnknownEntityAndBlockedEntity|ServiceEntityListFiltersAllowedEntitiesExactly|ServiceObjectGetRejectsBlockedEntityEvenWhenConfigContainsIt|ServiceReloadsRepositoryStatePerRequest|NewHTTPHandlerSupportsDefaultBasePath|NewHTTPHandlerSupportsMountedBasePath|NewHTTPHandlerRejectsRequestsOutsideConfiguredMount|RunSupportsStdioAndHTTP|RunHelp|RunRejectsHTTPOnlyFlagForStdio)' \
  -count=1

grep -Fq 'http://127.0.0.1:8080/' "$transport_reference"
grep -Fq 'http://127.0.0.1:9090/mcp' "$transport_reference"

printf 'Verified mergeway-mcp help, six registered read-only tools, deterministic examples, exact allow-list behavior, protocol-only stdio, and default/custom HTTP mounts.\n'
