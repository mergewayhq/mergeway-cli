#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")"/.. && pwd)"
BIN_DIR="$ROOT_DIR/e2e_test/bin"
MW_BIN="$BIN_DIR/mw"
STATE_DIR="$ROOT_DIR/e2e_test/state"
INPUT_DIR="$ROOT_DIR/e2e_test/state/input"

mkdir -p "$BIN_DIR"

echo "Building mw CLI..."
go build -o "$MW_BIN" "$ROOT_DIR"

if [ -d "$STATE_DIR" ]; then
  echo "Removing existing state..."
  rm -rf "$STATE_DIR"
fi

mkdir -p "$STATE_DIR"
cd "$STATE_DIR"

mkdir -p "$INPUT_DIR"

echo "Initializing repository..."
"$MW_BIN" --root "$STATE_DIR" init

cat > "$STATE_DIR/mergeway.yaml" <<'CONFIG'
version: 1
entities:
  User:
    identifier: id
    include:
      - data/users/*.yaml
    fields:
      id:
        type: string
        required: true
      name:
        type: string
        required: true
      email:
        type: string
CONFIG

echo "Listing types..."
"$MW_BIN" --root "$STATE_DIR" type list

echo "Showing type User..."
"$MW_BIN" --root "$STATE_DIR" --format json type show User

echo "Creating dummy object..."
cat > "$INPUT_DIR/user.yaml" <<'PAYLOAD'
id: user-001
name: Example User
email: user@example.com
PAYLOAD
"$MW_BIN" --root "$STATE_DIR" create --type User --file "$INPUT_DIR/user.yaml"

echo "Getting dummy object..."
"$MW_BIN" --root "$STATE_DIR" --format json get --type User user-001

echo "Updating dummy object..."
cat > "$INPUT_DIR/user_update.yaml" <<'UPDATE'
email: updated@example.com
UPDATE
"$MW_BIN" --root "$STATE_DIR" update --type User --id user-001 --merge --file "$INPUT_DIR/user_update.yaml"

echo "Getting updated object..."
echo
"$MW_BIN" --root "$STATE_DIR" --format yaml get --type User user-001
echo

echo "Deleting dummy object..."
"$MW_BIN" --root "$STATE_DIR" --yes delete --type User user-001

echo "Listing users to confirm deletion..."

echo
"$MW_BIN" --root "$STATE_DIR" list --type User
echo


echo "Validating example dataset..."
"$MW_BIN" --root "$ROOT_DIR/examples/full" validate

echo "Listing example users..."
"$MW_BIN" --root "$ROOT_DIR/examples/full" list --type User

echo "Showing example post..."
"$MW_BIN" --root "$ROOT_DIR/examples/full" --format json get --type Post post-001

echo "Done."
