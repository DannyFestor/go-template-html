#!/bin/sh
set -eu

dir=$1
name=${2:-}

case $name in
  '' | *[!a-z0-9_]*)
    echo "create-migration: NAME must be snake_case, such as NAME=create_users." >&2
    exit 1
    ;;
esac

highest=0
for file in "$dir"/[0-9]*_*.sql; do
  [ -e "$file" ] || continue
  base=${file##*/}
  number=${base%%_*}
  while [ "$number" != "${number#0}" ]; do
    number=${number#0}
  done
  if [ "${number:-0}" -gt "$highest" ]; then
    highest=$number
  fi
done

path=$(printf '%s/%05d_%s.sql' "$dir" $((highest + 1)) "$name")
printf -- '-- +goose Up\n' >"$path"
echo "$path"
