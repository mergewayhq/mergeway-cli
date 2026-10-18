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
/docs/how-to/ docs/how-to/index.html
/docs/reference/ docs/reference/index.html
/docs/explanation/ docs/explanation/index.html
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

printf 'Verified 14 required routes, navigation order, and legacy-link isolation.\n'
