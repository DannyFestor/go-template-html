#!/bin/sh
# git writes one "<local ref> <local sha> <remote ref> <remote sha>" line per pushed ref to stdin.
set -eu

protected_ref=$1

while read -r _local_ref _local_sha remote_ref _remote_sha; do
  if [ "$remote_ref" = "$protected_ref" ]; then
    echo "refuse-push-to-ref: pushes to ${protected_ref#refs/heads/} are refused, push a branch and open a pull request." >&2
    exit 1
  fi
done
