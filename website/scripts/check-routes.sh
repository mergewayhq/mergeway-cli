#!/bin/sh

set -eu

public_dir=${1:-website/public}

routes='
/ index.html
/docs/ docs/index.html
/docs/components/ docs/components/index.html
/docs/components/specification/ docs/components/specification/index.html
/docs/components/cli/ docs/components/cli/index.html
/docs/components/diff/ docs/components/diff/index.html
/docs/components/mcp/ docs/components/mcp/index.html
/docs/components/lsp/ docs/components/lsp/index.html
/docs/components/vscode/ docs/components/vscode/index.html
/docs/tutorials/ docs/tutorials/index.html
/docs/tutorials/first-repository/ docs/tutorials/first-repository/index.html
/docs/tutorials/review-semantic-change/ docs/tutorials/review-semantic-change/index.html
/docs/tutorials/inspect-with-ai-client/ docs/tutorials/inspect-with-ai-client/index.html
/docs/how-to/ docs/how-to/index.html
/docs/how-to/install/ docs/how-to/install/index.html
/docs/reference/ docs/reference/index.html
/docs/reference/cli/ docs/reference/cli/index.html
/docs/reference/diff/ docs/reference/diff/index.html
/docs/reference/mcp/ docs/reference/mcp/index.html
/docs/reference/mcp/transports/ docs/reference/mcp/transports/index.html
/docs/reference/mcp/tools/ docs/reference/mcp/tools/index.html
/docs/reference/specification/ docs/reference/specification/index.html
/docs/reference/specification/overview/ docs/reference/specification/overview/index.html
/docs/reference/specification/terminology/ docs/reference/specification/terminology/index.html
/docs/reference/specification/conformance/ docs/reference/specification/conformance/index.html
/docs/reference/specification/repository-configuration/ docs/reference/specification/repository-configuration/index.html
/docs/reference/specification/entities/ docs/reference/specification/entities/index.html
/docs/reference/specification/fields/ docs/reference/specification/fields/index.html
/docs/reference/specification/records/ docs/reference/specification/records/index.html
/docs/reference/specification/identifiers-references/ docs/reference/specification/identifiers-references/index.html
/docs/reference/specification/validation/ docs/reference/specification/validation/index.html
/docs/reference/specification/json-schema/ docs/reference/specification/json-schema/index.html
/docs/reference/specification/path-derived-data/ docs/reference/specification/path-derived-data/index.html
/docs/reference/specification/open-knowledge-format/ docs/reference/specification/open-knowledge-format/index.html
/docs/reference/specification/changelog/ docs/reference/specification/changelog/index.html
/docs/explanation/ docs/explanation/index.html
/docs/explanation/metadata-as-code/ docs/explanation/metadata-as-code/index.html
/docs/explanation/repository-model/ docs/explanation/repository-model/index.html
/docs/explanation/validation-model/ docs/explanation/validation-model/index.html
/docs/explanation/semantic-diff/ docs/explanation/semantic-diff/index.html
/docs/explanation/mcp-access-security/ docs/explanation/mcp-access-security/index.html
/docs/project/ docs/project/index.html'

for cli_page in init validate version entity-list entity-show fmt config-lint config-export gen-erd list files get create update delete export; do
  routes="$routes
/docs/reference/cli/$cli_page/ docs/reference/cli/$cli_page/index.html"
done

for guide_page in manage-records choose-validation split-files github-actions pre-commit export-data generate-erd format-ci; do
  routes="$routes
/docs/how-to/cli/$guide_page/ docs/how-to/cli/$guide_page/index.html"
done

routes="$routes
/docs/how-to/cli/ docs/how-to/cli/index.html"

for diff_page in working-tree revisions publish-summary json-automation; do
  routes="$routes
/docs/how-to/diff/$diff_page/ docs/how-to/diff/$diff_page/index.html"
done

routes="$routes
/docs/how-to/diff/ docs/how-to/diff/index.html"

for mcp_page in connect-stdio restrict-entities serve-http troubleshoot; do
  routes="$routes
/docs/how-to/mcp/$mcp_page/ docs/how-to/mcp/$mcp_page/index.html"
done

routes="$routes
/docs/how-to/mcp/ docs/how-to/mcp/index.html"

echo "$routes" | while read -r route file; do
  [ -n "$route" ] || continue
  if [ ! -s "$public_dir/$file" ]; then
    echo "missing rendered route $route ($public_dir/$file)" >&2
    exit 1
  fi
done

for alias_file in \
  cli-reference/diff.html \
  cli-reference/mcp.html \
  guides/use-mergeway-mcp-with-claude-and-codex.html \
  guides/communicate-changes-with-diff.html
do
  if [ ! -s "$public_dir/$alias_file" ]; then
    echo "missing Diff legacy alias $public_dir/$alias_file" >&2
    exit 1
  fi
done

navigation_file=$public_dir/docs/index.html
navigation=$(sed 's/></>\
</g' "$navigation_file")
previous_line=0

for route in \
  /docs/components/ \
  /docs/tutorials/ \
  /docs/how-to/ \
  /docs/reference/ \
  /docs/explanation/ \
  /docs/project/
do
  line=$(printf '%s\n' "$navigation" | grep -n -m 1 "href=$route" | cut -d: -f1)
  if [ -z "$line" ]; then
    echo "route $route is absent from the rendered navigation" >&2
    exit 1
  fi
  if [ "$line" -le "$previous_line" ]; then
    echo "route $route is out of order in the rendered navigation" >&2
    exit 1
  fi
  previous_line=$line
done

for route in \
  /docs/components/specification/ \
  /docs/components/cli/ \
  /docs/components/diff/ \
  /docs/components/mcp/ \
  /docs/components/lsp/ \
  /docs/components/vscode/
do
  if ! printf '%s\n' "$navigation" | grep -q "href=$route"; then
    echo "component route $route is absent from the rendered navigation" >&2
    exit 1
  fi
done

if grep -R -n -E --include='*.html' "href=[\"']?([^\"' >]*/)?docs/src(/|[\"'#])" "$public_dir"; then
  echo "rendered website contains a link into the legacy docs/src tree" >&2
  exit 1
fi

for alias_file in \
  getting-started/installation.html \
  getting-started/workspace-setup.html \
  guides/example-repo.html
do
  if [ ! -s "$public_dir/$alias_file" ]; then
    echo "missing legacy alias $public_dir/$alias_file" >&2
    exit 1
  fi
done

for alias_file in \
  cli-reference/index.html \
  cli-reference/init.html \
  cli-reference/validate.html \
  cli-reference/version.html \
  cli-reference/entity-list.html \
  cli-reference/entity-show.html \
  cli-reference/fmt.html \
  cli-reference/config-lint.html \
  cli-reference/config-export.html \
  cli-reference/gen-erd.html \
  cli-reference/list.html \
  cli-reference/files.html \
  cli-reference/get.html \
  cli-reference/create.html \
  cli-reference/update.html \
  cli-reference/delete.html \
  cli-reference/export.html \
  guides/setup-mergeway-github.html \
  guides/setup-mergeway-pre-commit.html
do
  if [ ! -s "$public_dir/$alias_file" ]; then
    echo "missing CLI legacy alias $public_dir/$alias_file" >&2
    exit 1
  fi
done

search_index=$(find "$public_dir" -maxdepth 1 -type f -name 'en.search-data*.json' -print -quit)
if [ -z "$search_index" ]; then
  echo "missing generated search index in $public_dir" >&2
  exit 1
fi
for term in \
  'Specification' \
  'mergeway-cli' \
  'mergeway-diff' \
  'mergeway-mcp' \
  'mergeway-lsp' \
  'VS Code'
do
  if ! grep -q "$term" "$search_index"; then
    echo "search index does not contain component term: $term" >&2
    exit 1
  fi
done

for term in \
  entity_list \
  entity_show \
  object_list \
  object_get \
  repository_export \
  files_list \
  'MCP access and security' \
  'streamable HTTP' \
  'entity allow-list'
do
  if ! grep -q "$term" "$search_index"; then
    echo "search index does not contain MCP term: $term" >&2
    exit 1
  fi
done

for term in init validate version 'entity list' 'entity show' fmt 'config lint' 'config export' gen-erd list files get create update delete export; do
  if ! grep -q "$term" "$search_index"; then
    echo "search index does not contain CLI command term: $term" >&2
    exit 1
  fi
done

for term in \
  'Semantic diff versus file diff' \
  'Review your first semantic change' \
  'working tree with HEAD' \
  'two Git revisions' \
  'JSON diff output'
do
  if ! grep -qi "$term" "$search_index"; then
    echo "search index does not contain Diff term: $term" >&2
    exit 1
  fi
done

erd_page=$public_dir/docs/how-to/cli/generate-erd/index.html
grep -q '/images/cli/erd-example-full.svg' "$erd_page" || {
  echo 'ERD guide does not render its example image' >&2
  exit 1
}
grep -q 'alt="Entity-relationship diagram for the full example' "$erd_page" || {
  echo 'ERD guide image is missing useful alt text' >&2
  exit 1
}

spec_source=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)/content/docs/reference/specification
for page in "$spec_source"/*.md; do
  if ! grep -q 'Normative v1' "$page"; then
    echo "specification page does not identify normative v1 applicability: $page" >&2
    exit 1
  fi
done

printf 'Verified 77 required routes, specification markers, legacy aliases, CLI, Diff, and MCP search terms, navigation order, ERD alt text, and legacy-link isolation.\n'
