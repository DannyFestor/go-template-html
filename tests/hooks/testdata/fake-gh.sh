#!/bin/sh
set -eu

state=$FAKE_GH_STATE
repo_path='repos/{owner}/{repo}'
rulesets_path="$repo_path/rulesets"

fail() {
  echo "fake gh: $*" >&2
  exit 1
}

auth_status() {
  [ ! -e "$state/logged-out" ] || fail "You are not logged into any GitHub hosts. To log in, run: gh auth login"
}

list_rulesets() {
  set -- "$state"/rulesets/*.json
  if [ ! -e "$1" ]; then
    echo '[]'
    return
  fi
  jq --slurp '[.[] | {id, name, source_type: "Repository"}]' "$@"
}

next_ruleset_id() {
  find "$state/rulesets" -name '*.json' | wc -l | awk '{ print $1 + 1 }'
}

store_ruleset() {
  id=$1
  input=$2
  jq --argjson id "$id" '. + {id: $id}' "$input" >"$state/rulesets/$id.json"
  cat "$state/rulesets/$id.json"
}

update_repo() {
  jq --slurp '.[0] * .[1]' "$state/repo.json" "$1" >"$state/repo.json.new"
  mv "$state/repo.json.new" "$state/repo.json"
  cat "$state/repo.json"
}

respond() {
  method=$1
  path=$2
  input=$3
  case "$method $path" in
    "GET $repo_path") cat "$state/repo.json" ;;
    "PATCH $repo_path") update_repo "$input" ;;
    "GET $rulesets_path?includes_parents=false") list_rulesets ;;
    "POST $rulesets_path") store_ruleset "$(next_ruleset_id)" "$input" ;;
    "GET $rulesets_path/"*) cat "$state/rulesets/${path##*/}.json" ;;
    "PUT $rulesets_path/"*) store_ruleset "${path##*/}" "$input" ;;
    *) fail "unexpected call: $method $path" ;;
  esac
}

# gh prints a string result raw and any other result as compact JSON.
api() {
  method=GET
  input=
  filter=.
  silent=false
  path=
  while [ $# -gt 0 ]; do
    case $1 in
      --method) method=$2 && shift 2 ;;
      --input) input=$2 && shift 2 ;;
      --jq) filter=$2 && shift 2 ;;
      --silent) silent=true && shift ;;
      --paginate) shift ;;
      *) path=$1 && shift ;;
    esac
  done
  echo "$method $path" >>"$state/calls"
  response=$(respond "$method" "$path" "$input")
  if [ "$silent" = false ]; then
    printf '%s\n' "$response" | jq --raw-output --compact-output "$filter"
  fi
}

main() {
  command -v jq >/dev/null 2>&1 || fail "the tooling tests need jq on PATH"
  mkdir -p "$state/rulesets"
  case "$1 ${2:-}" in
    "auth status") auth_status ;;
    "api "*) shift && api "$@" ;;
    *) fail "unexpected command: $*" ;;
  esac
}

main "$@"
