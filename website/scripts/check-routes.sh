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
/docs/how-to/ docs/how-to/index.html
/docs/how-to/install/ docs/how-to/install/index.html
/docs/reference/ docs/reference/index.html
/docs/explanation/ docs/explanation/index.html
/docs/explanation/metadata-as-code/ docs/explanation/metadata-as-code/index.html
/docs/explanation/repository-model/ docs/explanation/repository-model/index.html
/docs/explanation/validation-model/ docs/explanation/validation-model/index.html
/docs/project/ docs/project/index.html'

echo "$routes" | while read -r route file; do
  [ -n "$route" ] || continue
  if [ ! -s "$public_dir/$file" ]; then
    echo "missing rendered route $route ($public_dir/$file)" >&2
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

printf 'Verified 19 required routes, legacy aliases, component search terms, navigation order, and legacy-link isolation.\n'
