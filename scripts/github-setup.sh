#!/bin/sh
# shellcheck disable=SC2016 # The single-quoted $names are jq variables.
set -eu

repo_path='repos/{owner}/{repo}'
rulesets_path="$repo_path/rulesets"

repo_without_admin_filter='if .permissions.admin then empty else .full_name end'
ruleset_id_filter='.[] | select(.name == $want.name) | .id'
ruleset_fields_filter='{name, target, enforcement, bypass_actors, conditions, rules}'
ruleset_report_filter='. as $after | [$before | keys[] | select($before[.] != $after[.])] as $changed
  | if $changed == [] then "ruleset \(.name) already up to date"
    else "updated ruleset \(.name): \($changed | join(", "))" end'
merge_settings_changes_filter='. as $repo | $want | to_entries[] | select($repo[.key] != .value)
  | "  \(.key): \($repo[.key]) -> \(.value)"'

say() {
  echo "github-setup: $*"
}

fail() {
  say "$*" >&2
  exit 1
}

require_gh() {
  command -v gh >/dev/null 2>&1 || fail "gh is not installed, see https://cli.github.com."
}

require_authenticated() {
  gh auth status --active >/dev/null 2>&1 || fail "gh is not authenticated, run gh auth login first."
}

require_admin() {
  repo_without_admin=$(gh api "$repo_path" --jq "$repo_without_admin_filter")
  [ -z "$repo_without_admin" ] ||
    fail "you lack admin rights on $repo_without_admin, which the ruleset and merge settings need."
}

# gh api only runs jq on a response, so a local JSON file reaches a filter as a literal bound to a variable.
bind_json() {
  printf '(%s) as $%s | ' "$1" "$2"
}

bind_json_file() {
  bind_json "$(cat "$1")" "$2"
}

ruleset_id() {
  gh api --paginate "$rulesets_path?includes_parents=false" \
    --jq "$(bind_json_file "$ruleset_file" want)$ruleset_id_filter"
}

create_ruleset() {
  name=$(gh api --method POST "$rulesets_path" --input "$ruleset_file" --jq .name)
  say "created ruleset $name"
}

# The API adds defaults to what it stores, so only its own views from before and after the update compare equal.
update_ruleset() {
  ruleset_path="$rulesets_path/$1"
  before=$(gh api "$ruleset_path" --jq "$ruleset_fields_filter")
  say "$(gh api --method PUT "$ruleset_path" --input "$ruleset_file" \
    --jq "$(bind_json "$before" before)$ruleset_report_filter")"
}

sync_ruleset() {
  id=$(ruleset_id)
  if [ -z "$id" ]; then
    create_ruleset
  else
    update_ruleset "$id"
  fi
}

merge_settings_changes() {
  gh api "$repo_path" --jq "$(bind_json_file "$merge_settings_file" want)$merge_settings_changes_filter"
}

sync_merge_settings() {
  changes=$(merge_settings_changes)
  if [ -z "$changes" ]; then
    say "merge settings already up to date"
    return
  fi
  gh api --method PATCH "$repo_path" --input "$merge_settings_file" --silent
  say "updated merge settings:"
  echo "$changes"
}

main() {
  [ $# -eq 2 ] || fail "usage: $0 <ruleset-file> <merge-settings-file>"
  ruleset_file=$1
  merge_settings_file=$2
  require_gh
  require_authenticated
  require_admin
  sync_ruleset
  sync_merge_settings
}

main "$@"
