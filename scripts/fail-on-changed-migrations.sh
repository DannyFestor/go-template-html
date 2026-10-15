#!/bin/sh
set -eu

base_ref=$1
dir=$2

if ! git rev-parse --verify --quiet "$base_ref" >/dev/null; then
  echo "fail-on-changed-migrations: $base_ref is missing, fetch it first." >&2
  exit 1
fi

changed=$(git diff --name-only --diff-filter=MDR "$base_ref...HEAD" -- "$dir/*.sql")

if [ -n "$changed" ]; then
  echo "fail-on-changed-migrations: migrations already on $base_ref were changed, add a new migration instead:" >&2
  echo "$changed" >&2
  exit 1
fi
