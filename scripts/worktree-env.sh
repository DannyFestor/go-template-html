#!/bin/sh
set -eu

default_db_name=$1

git_dir=$(git rev-parse --absolute-git-dir)
common_dir=$(git rev-parse --path-format=absolute --git-common-dir)
worktree=$(git rev-parse --show-toplevel)
main_env=$(dirname "$common_dir")/.env

if [ "$git_dir" = "$common_dir" ]; then
  echo "worktree-env: $worktree is the main checkout, run this in a linked worktree." >&2
  exit 1
fi

if [ -e "$worktree/.env" ]; then
  echo "worktree-env: $worktree/.env already exists." >&2
  exit 1
fi

if [ ! -f "$main_env" ]; then
  echo "worktree-env: $main_env is missing, create it from .env.example first." >&2
  exit 1
fi

db_name=$(sed -n 's/^DB_NAME=//p' "$main_env" | tail -n 1)
suffix=$(printf '%s' "$worktree" | git hash-object --stdin | cut -c 1-6)

{
  cat "$main_env"
  if [ -n "$(tail -c 1 "$main_env")" ]; then
    echo
  fi
  printf 'TEST_DB_NAME=%s_wt_%s\n' "${db_name:-$default_db_name}" "$suffix"
} >"$worktree/.env"

echo "worktree-env: wrote $worktree/.env from $main_env"
