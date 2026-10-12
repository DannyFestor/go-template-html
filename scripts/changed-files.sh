#!/bin/sh
# pre-push jobs filter this list by their own globs and are skipped when nothing matches.
set -eu

base_ref=origin/main

# Without a remote main, as on the first push to an empty remote, there is no merge-base, so every job runs.
if ! git rev-parse --verify --quiet "$base_ref" >/dev/null; then
  git ls-files
  exit
fi

git diff --name-only "$base_ref...HEAD"
