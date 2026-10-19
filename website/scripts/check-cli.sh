#!/bin/sh

set -eu

repo_root=$(CDPATH= cd -- "$(dirname -- "$0")/../.." && pwd)
reference_dir=$repo_root/website/content/docs/reference/cli
fixture=$repo_root/examples/full
tmp_dir=$(mktemp -d)
trap 'rm -rf "$tmp_dir"' EXIT HUP INT TERM
cli=$tmp_dir/mergeway-cli

go build -o "$cli" "$repo_root/cmd/mergeway-cli"

expected_commands='init
validate
version
entity list
entity show
fmt
config lint
config export
gen-erd
list
files
get
create
update
delete
export'

command_slugs='init:init
validate:validate
version:version
entity list:entity-list
entity show:entity-show
fmt:fmt
config lint:config-lint
config export:config-export
gen-erd:gen-erd
list:list
files:files
get:get
create:create
update:update
delete:delete
export:export'

# Generate the public-operation inventory from executable help. Cobra's help and
# completion utilities and the two grouping commands are intentionally excluded.
root_commands=$($cli --help | awk '
  /^Available Commands:/ { in_commands=1; next }
  /^Flags:/ { in_commands=0 }
  in_commands && /^  [a-z]/ { print $1 }
' | grep -Ev '^(completion|help|config|entity)$')
nested_commands=''
for group in config entity; do
  children=$($cli "$group" --help | awk '
    /^Available Commands:/ { in_commands=1; next }
    /^Flags:/ { in_commands=0 }
    in_commands && /^  [a-z]/ { print $1 }
  ' | grep -Ev '^(completion|help)$')
  for child in $children; do
    nested_commands="${nested_commands}${group} ${child}
"
  done
done
generated_commands=$(printf '%s\n%s' "$root_commands" "$nested_commands" | sed '/^$/d')

if [ "$(printf '%s\n' "$generated_commands" | sort)" != "$(printf '%s\n' "$expected_commands" | sort)" ]; then
  echo "CLI help command inventory differs from the documented released inventory" >&2
  diff -u "$tmp_dir/expected" "$tmp_dir/generated" 2>/dev/null || true
  printf 'expected:\n%s\ngenerated:\n%s\n' "$expected_commands" "$generated_commands" >&2
  exit 1
fi

printf '%s\n' "$command_slugs" | while IFS=: read -r command slug; do
  page=$reference_dir/$slug.md
  [ -s "$page" ] || { echo "missing CLI reference page for $command: $page" >&2; exit 1; }
  grep -Fq "mergeway-cli $command" "$page" || {
    echo "reference page does not contain command spelling: $command" >&2
    exit 1
  }

  help=$($cli $command --help)
  printf '%s\n' "$help" | grep -oE -- '--[a-z][a-z-]*' | sort -u | while read -r flag; do
    case "$flag" in
      --config|--fail-fast|--format|--help|--root|--verbose|--yes)
        grep -Fq -- "$flag" "$reference_dir/_index.md" || {
          echo "global help flag $flag is undocumented" >&2
          exit 1
        }
        ;;
      *)
        grep -Fq -- "$flag" "$page" || {
          echo "$command help flag $flag is undocumented in $page" >&2
          exit 1
        }
        ;;
    esac
  done
done

cp -R "$fixture" "$tmp_dir/repository"
root=$tmp_dir/repository

test "$($cli --root "$root" config lint)" = 'configuration valid'
$cli --root "$root" validate | grep -q 'validation succeeded'
$cli --root "$root" entity list | grep -qx 'Post'
$cli --root "$root" entity show Post | grep -q '^name: Post'
$cli --root "$root" list --type Post | grep -qx 'post-001'
$cli --root "$root" get --type Post post-001 | grep -q '^title: Launch Day'
$cli --root "$root" files --type User | grep -q 'data/users/alice.yaml'
$cli --root "$root" --format json config export --type Post | grep -q 'x-reference-type'

printf '%s\n' 'name: Carol Example' 'email: carol@example.com' 'roles: [author]' |
  $cli --root "$root" create --type User --id user-carol | grep -q 'User user-carol created'
printf '%s\n' 'name: Carol Updated' |
  $cli --root "$root" update --type User --id user-carol --merge | grep -q 'User user-carol updated'
$cli --root "$root" get --type User user-carol | grep -q 'Carol Updated'
$cli --root "$root" --yes delete --type User user-carol | grep -q 'User user-carol deleted'

$cli --root "$root" --format json export User Post --output "$tmp_dir/export.json"
grep -q '"post-001"' "$tmp_dir/export.json"
$cli --root "$root" fmt --lint
$cli --root "$root" gen-erd --path "$tmp_dir/erd.svg"
grep -q '<svg' "$tmp_dir/erd.svg"
$cli --root "$root" validate | grep -q 'validation succeeded'

cp -R "$fixture" "$tmp_dir/invalid"
printf '%s\n' 'id: post-001' 'title: Broken post' 'status: PUBLISHED' 'author: missing-user' > "$tmp_dir/invalid/data/posts/launch.yaml"
if $cli --root "$tmp_dir/invalid" validate >/dev/null 2>&1; then
  echo 'validate unexpectedly returned success for examples/invalid' >&2
  exit 1
fi

printf 'Verified 16 released CLI operations against executable help and one reference page each.\n'
printf 'Executed read, mutation, validation, formatting, export, and ERD workflows on temporary fixtures.\n'
