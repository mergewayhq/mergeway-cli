#!/bin/sh

set -eu

public_dir=${1:-website/public}
hugo_command=${HUGO:-hugo}
hugo_cache=${HUGO_CACHEDIR:-.cache/hugo}
linkcheck_command=${LINKCHECK:-lychee}
host=127.0.0.1
port=14139
base_url=http://$host:$port
input_file=$(mktemp)
server_log=$(mktemp)
server_pid=

cleanup() {
  if [ -n "$server_pid" ]; then
    kill "$server_pid" 2>/dev/null || true
    wait "$server_pid" 2>/dev/null || true
  fi
  rm -f "$input_file" "$server_log"
}
trap cleanup EXIT HUP INT TERM

(cd website && HUGO_CACHEDIR="$hugo_cache" "$hugo_command" server \
  --environment production \
  --baseURL "$base_url/" \
  --buildDrafts=false \
  --buildFuture=false \
  --disableFastRender \
  --panicOnWarning \
  --bind "$host" \
  --port "$port") >"$server_log" 2>&1 &
server_pid=$!

attempt=0
until curl --fail --silent --output /dev/null "$base_url/"; do
  attempt=$((attempt + 1))
  if ! kill -0 "$server_pid" 2>/dev/null || [ "$attempt" -ge 20 ]; then
    cat "$server_log" >&2
    echo "website preview server did not become ready" >&2
    exit 1
  fi
  sleep 1
done

find "$public_dir" -type f -name '*.html' -print | sort | while read -r file; do
  relative=${file#"$public_dir"}
  case "$relative" in
    */index.html) relative=${relative%index.html} ;;
  esac
  printf '%s%s\n' "$base_url" "$relative"
done >"$input_file"

"$linkcheck_command" --config website/lychee.toml --files-from "$input_file"
