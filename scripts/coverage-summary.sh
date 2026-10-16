#!/bin/sh
set -eu

scripts=$(dirname "$0")

awk -v target="$2" -f "$scripts/coverage-profile.awk" -f "$scripts/coverage-summary.awk" "$1"
