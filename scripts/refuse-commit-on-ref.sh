#!/bin/sh
# A detached HEAD, as during a rebase, has no symbolic ref and passes.
set -eu

protected_ref=$1

if [ "$(git symbolic-ref --quiet HEAD || true)" = "$protected_ref" ]; then
  echo "refuse-commit-on-ref: commits on ${protected_ref#refs/heads/} are refused, branch first." >&2
  exit 1
fi
