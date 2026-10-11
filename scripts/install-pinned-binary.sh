#!/bin/sh
# The checksum file uses `sha256sum` format, one line per release asset.
set -eu

fail() {
  echo "install-pinned-binary: $*" >&2
  exit 1
}

sha256_of() {
  # macOS ships shasum, most Linux images ship sha256sum.
  if command -v sha256sum >/dev/null 2>&1; then
    sha256sum "$1" | cut -d ' ' -f 1
  else
    shasum -a 256 "$1" | cut -d ' ' -f 1
  fi
}

pinned_sha256_of() {
  awk -v asset="$2" '$2 == asset { print $1 }' "$1"
}

verify_checksum() {
  asset_path=$1
  checksum_file=$2
  asset_name=$(basename "$asset_path")
  expected=$(pinned_sha256_of "$checksum_file" "$asset_name")
  [ -n "$expected" ] || fail "no checksum for $asset_name in $checksum_file (unsupported platform?)"
  actual=$(sha256_of "$asset_path")
  [ "$actual" = "$expected" ] || fail "checksum mismatch for $asset_name: expected $expected, got $actual"
}

extract_binary() {
  asset_path=$1
  binary_in_archive=$2
  if [ -z "$binary_in_archive" ]; then
    echo "$asset_path"
    return
  fi
  archive_dir=$(dirname "$asset_path")
  tar -xf "$asset_path" -C "$archive_dir"
  echo "$archive_dir/$binary_in_archive"
}

main() {
  [ $# -ge 3 ] || fail "usage: $0 <destination> <asset-url> <checksum-file> [<binary-path-in-archive>]"
  destination=$1
  asset_url=$2
  checksum_file=$3
  binary_in_archive=${4:-}

  workdir=$(mktemp -d)
  trap 'rm -rf "$workdir"' EXIT
  asset_path="$workdir/${asset_url##*/}"

  curl --fail --silent --show-error --location --output "$asset_path" "$asset_url"
  verify_checksum "$asset_path" "$checksum_file"
  binary=$(extract_binary "$asset_path" "$binary_in_archive")
  mkdir -p "$(dirname "$destination")"
  install -m 0755 "$binary" "$destination"
}

main "$@"
