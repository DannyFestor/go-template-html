#!/bin/sh
# For releases that publish their own checksum file, which is then kept verbatim as the pinned-binary checksums.
set -eu

fail() {
  echo "pin-latest-release: $*" >&2
  exit 1
}

# GitHub redirects /releases/latest to /releases/tag/<tag>, so no API token or JSON parsing is needed.
latest_tag() {
  latest_url=$(curl --fail --silent --show-error --location --head --output /dev/null \
    --write-out '%{url_effective}' "https://github.com/$1/releases/latest")
  echo "${latest_url##*/}"
}

main() {
  [ $# -eq 4 ] || fail "usage: $0 <owner/repo> <checksum-asset> <version-file> <checksum-file>"
  repo=$1
  checksum_asset=$2
  version_file=$3
  checksum_file=$4

  tag=$(latest_tag "$repo")
  curl --fail --silent --show-error --location --output "$checksum_file" \
    "https://github.com/$repo/releases/download/$tag/$checksum_asset"
  echo "$tag" >"$version_file"
}

main "$@"
