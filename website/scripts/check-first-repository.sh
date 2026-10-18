#!/bin/sh

set -eu

repository_root=$(CDPATH= cd -- "$(dirname "$0")/../.." && pwd)
temporary_root=$(mktemp -d "${TMPDIR:-/tmp}/mergeway-first-repository.XXXXXX")
trap 'rm -rf "$temporary_root"' EXIT HUP INT TERM

cli=${MERGEWAY_CLI:-$temporary_root/mergeway-cli}
if [ -z "${MERGEWAY_CLI:-}" ]; then
  go build -o "$cli" "$repository_root/cmd/mergeway-cli"
fi

cp -R "$repository_root/examples/full" "$temporary_root/repository"

entities=$("$cli" --root "$temporary_root/repository" entity list)
[ "$entities" = "$(printf 'Comment\nPost\nTag\nUser')" ]

users=$("$cli" --root "$temporary_root/repository" list --type User)
printf '%s\n' "$users" | grep -qx 'user-alice'
printf '%s\n' "$users" | grep -qx 'user-bob'

post=$("$cli" --root "$temporary_root/repository" get --type Post post-001)
printf '%s\n' "$post" | grep -q '^author: user-alice$'
printf '%s\n' "$post" | grep -q '^    - tag-product$'

cat > "$temporary_root/new-post.yaml" <<'YAML'
id: post-002
title: Documentation Tour
status: DRAFT
author: user-bob
tags:
  - tag-update
body: |
  Learning Mergeway one record at a time.
YAML

created=$("$cli" --root "$temporary_root/repository" create --type Post --file "$temporary_root/new-post.yaml")
[ "$created" = 'Post post-002 created' ]
grep -q '^author: user-bob$' "$temporary_root/repository/data/posts/post-002.yaml"

validation=$("$cli" --root "$temporary_root/repository" validate)
[ "$validation" = 'status: validation succeeded' ]

"$cli" --root "$temporary_root/repository" --format json export --output "$temporary_root/repository.json"
grep -q '"id": "post-002"' "$temporary_root/repository.json"
grep -q '"author": "user-bob"' "$temporary_root/repository.json"
grep -q '"tag-update"' "$temporary_root/repository.json"

printf 'Verified first-repository tutorial against a temporary copy of examples/full.\n'
