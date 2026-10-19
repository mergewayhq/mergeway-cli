#!/bin/sh

set -eu

public_dir=${1:-website/public}
inventory=website/legacy-routes.tsv
base_url=http://127.0.0.1:14140
server_log=$(mktemp)
server_pid=

cleanup() {
  if [ -n "$server_pid" ]; then
    kill "$server_pid" 2>/dev/null || true
    wait "$server_pid" 2>/dev/null || true
  fi
  rm -f "$server_log"
}
trap cleanup EXIT HUP INT TERM

[ -s "$inventory" ] || { echo "missing legacy route inventory: $inventory" >&2; exit 1; }
[ "$(head -n 1 "$inventory")" = "legacy_path	live_status	disposition	destination	implementation" ] || {
  echo "invalid legacy route inventory header" >&2
  exit 1
}

published=$(awk -F '\t' 'NR > 1 && $2 == "published" { count++ } END { print count + 0 }' "$inventory")
[ "$published" -eq 36 ] || { echo "expected 36 published live HTML routes, found $published" >&2; exit 1; }

while IFS='	' read -r legacy status disposition destination implementation; do
  [ "$legacy" != "legacy_path" ] || continue
  relative=${legacy#/}
  case "$disposition" in
    retained|system)
      [ -s "$public_dir/$relative" ] || { echo "missing retained route $legacy" >&2; exit 1; }
      ;;
    redirect)
      [ "$implementation" = "hugo-alias" ] || { echo "unsupported redirect implementation for $legacy" >&2; exit 1; }
      [ -s "$public_dir/$relative" ] || { echo "missing alias route $legacy" >&2; exit 1; }
      grep -Fq "url=https://mergewayhq.github.io$destination" "$public_dir/$relative" || {
        echo "alias $legacy does not target $destination" >&2
        exit 1
      }
      ;;
    retired)
      [ "$destination" = "-" ] || { echo "retired route $legacy has a destination" >&2; exit 1; }
      [ ! -e "$public_dir/$relative" ] || { echo "retired route remains in artifact: $legacy" >&2; exit 1; }
      ;;
    *) echo "unknown disposition $disposition for $legacy" >&2; exit 1 ;;
  esac
done < "$inventory"

node website/scripts/check-html.mjs "$public_dir"

search_index=$(find "$public_dir" -maxdepth 1 -type f -name 'en.search-data*.json' -print -quit)
[ -n "$search_index" ] || { echo "missing search data" >&2; exit 1; }
for term in \
  'Specification' mergeway-cli mergeway-diff mergeway-mcp mergeway-lsp 'VS Code' \
  'config export' gen-erd mergeway.version identifier json_schema entity_list object_get \
  textDocumentSync workspaceSymbolProvider mergeway.lsp.path mergeway.lsp.trace.server \
  'unknown entity' 'already exists' 'configuration valid'
do
  grep -Fqi "$term" "$search_index" || { echo "search data lacks release-candidate term: $term" >&2; exit 1; }
done

css_file=$(find "$public_dir" -maxdepth 1 -type f -name 'book.min*.css' -print -quit)
grep -Fq 'focus-visible' "$css_file" || { echo "artifact lacks visible-focus styling" >&2; exit 1; }
grep -Fq 'prefers-reduced-motion' "$css_file" || { echo "artifact lacks reduced-motion styling" >&2; exit 1; }

if find "$public_dir" -type f \( \
  -name '*.map' -o -name '.hugo_build.lock' -o -path '*/scratch/*' -o \
  -path '*/testdata/*' -o -path '*/resources/_gen/*' \
\) -print | grep .; then
  echo "release candidate contains forbidden development output" >&2
  exit 1
fi
go run website/scripts/static-server.go -dir "$public_dir" >"$server_log" 2>&1 &
server_pid=$!
attempt=0
until curl --fail --silent --output /dev/null "$base_url/"; do
  attempt=$((attempt + 1))
  if ! kill -0 "$server_pid" 2>/dev/null || [ "$attempt" -ge 20 ]; then
    cat "$server_log" >&2
    echo "release-candidate server did not become ready" >&2
    exit 1
  fi
  sleep 1
done

mapped=0
while IFS='	' read -r legacy status disposition destination implementation; do
  [ "$legacy" != "legacy_path" ] || continue
  [ "$disposition" != "retired" ] || continue
  curl --fail --silent --output /dev/null "$base_url$legacy"
  if [ "$disposition" = "redirect" ]; then
    curl --fail --location --silent --output /dev/null "$base_url$destination"
    mapped=$((mapped + 1))
  fi
done < "$inventory"

printf 'Legacy route inventory: %s published routes; %s redirects have reachable source files and destinations; 2 mdBook-only views intentionally retired.\n' "$published" "$mapped"
printf 'Production artifact hygiene, production canonicals, search coverage, focus/reduced-motion CSS, and static route crawling passed.\n'
