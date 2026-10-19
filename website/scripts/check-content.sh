#!/bin/sh

set -eu

public_dir=${1:-website/public}
ledger=website/migration-ledger.md

if [ ! -s "$ledger" ]; then
  echo "missing migration ledger: $ledger" >&2
  exit 1
fi

source_count=0
for source in $(find docs/src -type f -print; find docs/arch -type f -print; printf '%s\n' docs/STYLEGUIDE.md); do
  source_count=$((source_count + 1))
  if ! grep -Fq "| \`$source\` |" "$ledger"; then
    echo "migration ledger does not account for $source" >&2
    exit 1
  fi
done

ledger_count=$(grep -c '^| `docs/' "$ledger")
if [ "$ledger_count" -ne "$source_count" ]; then
  echo "migration ledger has $ledger_count source rows; expected $source_count" >&2
  exit 1
fi

if grep '^| `docs/' "$ledger" | grep -Ev '\| (Move|Split|Merge|Archive) \|'; then
  echo "migration ledger contains an invalid disposition" >&2
  exit 1
fi

tmp_dir=$(mktemp -d)
trap 'rm -rf "$tmp_dir"' EXIT HUP INT TERM

go build -o "$tmp_dir/mergeway-cli" ./cmd/mergeway-cli

for fixture in json-schema jsonpath inheritance reference-union path-identifier path-segments okf; do
  "$tmp_dir/mergeway-cli" --root "examples/$fixture" validate >"$tmp_dir/$fixture.out"
done

"$tmp_dir/mergeway-cli" --root examples/jsonpath list --type User >"$tmp_dir/jsonpath-list.out"
grep -q 'User-001' "$tmp_dir/jsonpath-list.out"
grep -q 'User-002' "$tmp_dir/jsonpath-list.out"

"$tmp_dir/mergeway-cli" --root examples/inheritance list --type Animal >"$tmp_dir/inheritance-list.out"
grep -q 'dog-1' "$tmp_dir/inheritance-list.out"

"$tmp_dir/mergeway-cli" --root examples/reference-union get --type Activity activity-2 >"$tmp_dir/union-get.out"
grep -q 'team-core' "$tmp_dir/union-get.out"

"$tmp_dir/mergeway-cli" --root examples/path-identifier list --type Note >"$tmp_dir/path-id-list.out"
grep -q 'data/notes/alpha.yaml' "$tmp_dir/path-id-list.out"

"$tmp_dir/mergeway-cli" --root examples/path-segments get --type Page guide-install >"$tmp_dir/path-get.out"
grep -q 'section: guides' "$tmp_dir/path-get.out"
grep -q 'filename: install.yaml' "$tmp_dir/path-get.out"

"$tmp_dir/mergeway-cli" --root examples/okf --format json get --type Knowledge customer-orders.md >"$tmp_dir/okf-get.out"
grep -q '"body"' "$tmp_dir/okf-get.out"

duplicate_hashes=$(find website/content/docs -type f -name '*.md' -exec sha256sum {} \; | sort | cut -d' ' -f1 | uniq -d)
if [ -n "$duplicate_hashes" ]; then
  echo "duplicate canonical Markdown content detected: $duplicate_hashes" >&2
  exit 1
fi

page_count=0
for source in $(find website/content/docs -type f -name '*.md' -print); do
  relative=${source#website/content/}
  case $relative in
    */_index.md) route=${relative%/_index.md}/ ;;
    *) route=${relative%.md}/ ;;
  esac
  rendered="$public_dir/$route/index.html"
  if [ ! -s "$rendered" ]; then
    echo "canonical page has no rendered route: $source ($rendered)" >&2
    exit 1
  fi
  page_count=$((page_count + 1))
done

if find website/content/docs -type f -name '*.md' ! -path '*/components/vscode.md' \
  -exec grep -n -i -E 'placeholder page|later-phase scope|will be added in a later|content will be added|coming soon' {} +; then
  echo "delivered documentation contains placeholder language" >&2
  exit 1
fi

printf 'Migration ledger accounts for %s legacy sources.\n' "$source_count"
printf 'Validated 7 released advanced fixtures and representative results.\n'
printf 'Orphan report: 0 unexplained pages across %s rendered canonical pages.\n' "$page_count"
printf 'Duplicate-content report: 0 byte-identical canonical Markdown pages.\n'
