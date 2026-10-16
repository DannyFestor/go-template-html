#!/bin/sh
set -eu

scripts=$(dirname "$0")

awk -f "$scripts/coverage-profile.awk" -f "$scripts/drop-generated-coverage.awk" "$1"
