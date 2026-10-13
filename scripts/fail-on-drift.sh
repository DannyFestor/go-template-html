#!/bin/sh
# git diff skips untracked files, so a file make generate newly created only shows once it is intent-to-add.
set -eu

git add --intent-to-add --all

if ! git diff --exit-code --stat; then
  echo "fail-on-drift: generated files are out of date, run make generate and commit the result." >&2
  exit 1
fi
