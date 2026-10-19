#!/bin/sh

set -eu

repo_root=$(CDPATH= cd -- "$(dirname -- "$0")/../.." && pwd)
reference=$repo_root/website/content/docs/reference/diff/_index.md
expected_json=$repo_root/website/testdata/diff/modified-user.json
fixture=$repo_root/examples/full
tmp_dir=$(mktemp -d)
trap 'rm -rf "$tmp_dir"' EXIT HUP INT TERM
diff_bin=$tmp_dir/mergeway-diff
cli_bin=$tmp_dir/mergeway-cli

go build -o "$diff_bin" "$repo_root/cmd/mergeway-diff"
go build -o "$cli_bin" "$repo_root/cmd/mergeway-cli"

help=$($diff_bin --help)
for flag in --config --format --help --root; do
  printf '%s\n' "$help" | grep -Fq -- "$flag" || {
    echo "mergeway-diff help is missing $flag" >&2
    exit 1
  }
  grep -Fq -- "$flag" "$reference" || {
    echo "mergeway-diff reference is missing $flag" >&2
    exit 1
  }
done

init_repo() {
  root=$1
  cp -R "$fixture" "$root"
  git -C "$root" init -q
  git -C "$root" config user.name 'Mergeway Documentation'
  git -C "$root" config user.email 'docs@example.com'
  git -C "$root" add .
  git -C "$root" commit -q -m 'Initial fixture'
}

# The checked JSON sample is generated from the same v0.10.0-equivalent source
# and examples/full data used by the published reference.
json_root=$tmp_dir/json-repository
init_repo "$json_root"
sed 's/name: Alice Example/name: Alice Reviewed/' \
  "$json_root/data/users/alice.yaml" > "$json_root/data/users/alice.yaml.new"
mv "$json_root/data/users/alice.yaml.new" "$json_root/data/users/alice.yaml"
$diff_bin --root "$json_root" --format json > "$tmp_dir/actual.json"
cmp "$expected_json" "$tmp_dir/actual.json"

# Exercise the exact zero-, one-, and two-revision snapshot boundaries.
snapshot_root=$tmp_dir/snapshot-repository
init_repo "$snapshot_root"
base=$(git -C "$snapshot_root" rev-parse HEAD)

sed 's/Bob Example/Bob Staged/' \
  "$snapshot_root/data/users/bob.json" > "$snapshot_root/data/users/bob.json.new"
mv "$snapshot_root/data/users/bob.json.new" "$snapshot_root/data/users/bob.json"
git -C "$snapshot_root" add data/users/bob.json

sed 's/Product Launch/Product Launch Reviewed/' \
  "$snapshot_root/data/tags/product.yaml" > "$snapshot_root/data/tags/product.yaml.new"
mv "$snapshot_root/data/tags/product.yaml.new" "$snapshot_root/data/tags/product.yaml"

mkdir -p "$snapshot_root/data/posts"
printf '%s\n' \
  'id: post-002' \
  'title: Launch retrospective' \
  'status: DRAFT' \
  'author: user-alice' \
  'tags:' \
  '  - tag-update' \
  'body: Follow-up notes from the launch.' \
  > "$snapshot_root/data/posts/retrospective.yaml"

$diff_bin --root "$snapshot_root" > "$tmp_dir/zero.txt"
grep -Fq 'MODIFIED Tag[tag-product]' "$tmp_dir/zero.txt"
grep -Fq 'ADDED Post[post-002]' "$tmp_dir/zero.txt"
if grep -Fq 'User[user-bob]' "$tmp_dir/zero.txt"; then
  echo 'zero-argument diff unexpectedly included a staged-only edit' >&2
  exit 1
fi

$diff_bin --root "$snapshot_root" HEAD > "$tmp_dir/one.txt"
grep -Fq 'MODIFIED User[user-bob]' "$tmp_dir/one.txt"
grep -Fq 'MODIFIED Tag[tag-product]' "$tmp_dir/one.txt"
grep -Fq 'ADDED Post[post-002]' "$tmp_dir/one.txt"

git -C "$snapshot_root" add .
git -C "$snapshot_root" commit -q -m 'Semantic changes'
right=$(git -C "$snapshot_root" rev-parse HEAD)
sed 's/title: Launch Day/title: Dirty local title/' \
  "$snapshot_root/data/posts/launch.yaml" > "$snapshot_root/data/posts/launch.yaml.new"
mv "$snapshot_root/data/posts/launch.yaml.new" "$snapshot_root/data/posts/launch.yaml"
$diff_bin --root "$snapshot_root" "$base" "$right" > "$tmp_dir/two.txt"
grep -Fq 'MODIFIED User[user-bob]' "$tmp_dir/two.txt"
grep -Fq 'MODIFIED Tag[tag-product]' "$tmp_dir/two.txt"
grep -Fq 'ADDED Post[post-002]' "$tmp_dir/two.txt"
if grep -Fq 'Dirty local title' "$tmp_dir/two.txt"; then
  echo 'two-revision diff unexpectedly included a working-tree edit' >&2
  exit 1
fi

# Execute the published tutorial in a separate disposable repository.
tutorial_root=$tmp_dir/tutorial-repository
init_repo "$tutorial_root"
sed 's/title: Launch Day/title: Launch Day reviewed/' \
  "$tutorial_root/data/posts/launch.yaml" > "$tutorial_root/data/posts/launch.yaml.new"
mv "$tutorial_root/data/posts/launch.yaml.new" "$tutorial_root/data/posts/launch.yaml"
printf '%s\n' \
  'id: post-002' \
  'title: Launch retrospective' \
  'status: DRAFT' \
  'author: user-alice' \
  'tags:' \
  '  - tag-update' \
  'body: Follow-up notes from the launch.' \
  > "$tutorial_root/data/posts/retrospective.yaml"
rm "$tutorial_root/data/comments/launch-comment.json"
$cli_bin --root "$tutorial_root" validate | grep -q 'validation succeeded'
$diff_bin --root "$tutorial_root" HEAD > "$tmp_dir/tutorial.txt"
grep -Fq 'REMOVED Comment[comment-001]' "$tmp_dir/tutorial.txt"
grep -Fq 'MODIFIED Post[post-001]' "$tmp_dir/tutorial.txt"
grep -Fq 'title: "Launch Day" -> "Launch Day reviewed"' "$tmp_dir/tutorial.txt"
grep -Fq 'ADDED Post[post-002]' "$tmp_dir/tutorial.txt"

# Configuration-only edits do not become entries.
config_root=$tmp_dir/config-repository
init_repo "$config_root"
sed 's/Directory of users who can author and moderate posts/Directory of repository users/' \
  "$config_root/entities/User.yaml" > "$config_root/entities/User.yaml.new"
mv "$config_root/entities/User.yaml.new" "$config_root/entities/User.yaml"
test "$($diff_bin --root "$config_root")" = 'No changes.'

# Invalid repository data is an error and never a successful/non-empty diff.
invalid_root=$tmp_dir/invalid-repository
init_repo "$invalid_root"
printf '%s\n' 'id: user-bob' 'name: [' > "$invalid_root/data/users/bob.json"
if $diff_bin --root "$invalid_root" > "$tmp_dir/invalid.out" 2> "$tmp_dir/invalid.err"; then
  echo 'mergeway-diff unexpectedly accepted invalid repository data' >&2
  exit 1
fi
test ! -s "$tmp_dir/invalid.out"
grep -Fq 'data error' "$tmp_dir/invalid.err"
grep -Fq 'data/users/bob.json' "$tmp_dir/invalid.err"

printf 'Verified zero-, one-, and two-revision snapshot modes, including staged, unstaged, untracked, committed, and configuration-only boundaries.\n'
printf 'Verified added, modified, and removed records; JSON schema v1 sample; invalid-data failure; and the complete temporary-repository tutorial.\n'
printf '%s\n' 'Human example:'
cat "$tmp_dir/tutorial.txt"
printf '%s\n' 'JSON example:'
cat "$tmp_dir/actual.json"
