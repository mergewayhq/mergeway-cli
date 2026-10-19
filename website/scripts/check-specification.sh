#!/bin/sh

set -eu

repo_root=$(CDPATH= cd -- "$(dirname -- "$0")/../.." && pwd)
fixture_source=$repo_root/website/testdata/specification-v1
cli=${MERGEWAY_CLI:-$repo_root/.cache/website-check/mergeway-cli}

if [ ! -x "$cli" ]; then
  mkdir -p "$(dirname -- "$cli")"
  go build -o "$cli" "$repo_root/cmd/mergeway-cli"
fi

fixture=$(mktemp -d)
cleanup() {
  rm -rf "$fixture"
}
trap cleanup EXIT HUP INT TERM

cp -R "$fixture_source/." "$fixture/"

validation_output=$($cli --root "$fixture" validate)
printf '%s\n' "$validation_output" | grep -q 'status: validation succeeded'

$cli --root "$fixture" --format json config export --type Customer > "$fixture/customer-export.json"
grep -q '"x-reference-type": "User"' "$fixture/customer-export.json"
grep -q '"https://json-schema.org/draft/2020-12/schema"' "$fixture/customer-export.json"

for example in \
  mergeway.yaml \
  entities/core.yaml \
  entities/files.yaml \
  entities/interoperability.yaml \
  schemas/customer.json \
  data/users/alice.yaml \
  data/events.yaml \
  data/catalog.json \
  data/knowledge/orders.md
do
  if [ ! -s "$fixture_source/$example" ]; then
    echo "missing specification example: $example" >&2
    exit 1
  fi
done

printf 'Verified normative specification examples and JSON Schema interoperability against mergeway.version 1.\n'
