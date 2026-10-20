#!/bin/sh

set -eu

public_dir=${1:-website/public}
inventory=website/legacy-routes.tsv

while IFS='	' read -r legacy _status _disposition destination implementation; do
  [ "$legacy" != "legacy_path" ] || continue
  [ "$implementation" = "generated-static-alias" ] || continue

  relative=${legacy#/}
  case "$relative" in
    */) relative=${relative}index.html ;;
  esac
  output=$public_dir/$relative
  mkdir -p "$(dirname "$output")"
  target=https://mergewayhq.github.io$destination
  printf '%s\n' "<!doctype html><html lang=en><head><title>$target</title><link rel=canonical href=$target><meta charset=utf-8><meta http-equiv=refresh content=\"0; url=$target\"></head></html>" >"$output"
done < "$inventory"
