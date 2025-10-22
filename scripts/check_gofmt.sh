#!/usr/bin/env bash
set -euo pipefail

mapfile -t files < <(find . -type f -name '*.go' -not -path './.git/*' -not -path './.cache/*')
if [ ${#files[@]} -eq 0 ]; then
  exit 0
fi

CHANGED=$(gofmt -l "${files[@]}")
if [ -n "$CHANGED" ]; then
  echo "gofmt found unformatted files:" >&2
  echo "$CHANGED" >&2
  exit 1
fi
