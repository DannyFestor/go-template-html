#!/bin/sh
set -eu

template_module='github.com/DannyFestor/go-template-html'
template_slug='go-template-html'
template_db_name='go_template_html'
template_display_name='Go Template'
slug_pattern='^[a-z][a-z0-9-]*$'

say() {
  echo "rename: $*"
}

fail() {
  say "$*" >&2
  exit 1
}

usage() {
  fail "usage: $0 sweep <module> [<display name>] | $0 verify"
}

# This script and its tests keep the template's tokens, so a renamed clone still recognises the template and refuses.
grep_tracked_files_for_tokens() {
  git -c core.quotePath=false grep -I -F "$@" \
    -e "$template_module" -e "$template_slug" -e "$template_db_name" -e "$template_display_name" \
    -- ':(top)' ':(top,exclude)scripts/rename.sh' ':(top,exclude)tests/hooks/rename_test.go'
}

current_module() {
  awk '$1 == "module" { print $2; exit }' go.mod
}

require_unrenamed_template() {
  module_in_go_mod=$(current_module)
  [ "$module_in_go_mod" = "$template_module" ] ||
    fail "already renamed: go.mod's module is $module_in_go_mod, not the template's."
}

require_valid_slug() {
  printf '%s\n' "$1" | LC_ALL=C grep -Eq "$slug_pattern" ||
    fail "MODULE must be a module path whose last segment matches $slug_pattern, such as MODULE=example.com/acme/shop."
}

title_case() {
  printf '%s\n' "$1" | tr '-' ' ' |
    awk '{ for (i = 1; i <= NF; i++) $i = toupper(substr($i, 1, 1)) substr($i, 2) } 1'
}

escape_pattern() {
  printf '%s\n' "$1" | sed 's/[]\$*.^[|]/\\&/g'
}

escape_replacement() {
  printf '%s\n' "$1" | sed 's/[\&|]/\\&/g'
}

sed_substitution() {
  printf 's|%s|%s|g\n' "$(escape_pattern "$1")" "$(escape_replacement "$2")"
}

# The module path holds the slug, so it is replaced first.
sed_program() {
  sed_substitution "$template_module" "$module"
  sed_substitution "$template_slug" "$slug"
  sed_substitution "$template_db_name" "$db_name"
  sed_substitution "$template_display_name" "$display_name"
}

# BSD and GNU sed disagree on how -i takes its backup suffix, so each file goes through a scratch file. cat keeps
# the file's mode.
rewrite_file() {
  sed -f "$program" "$1" >"$scratch"
  cat "$scratch" >"$1"
}

rewrite_tracked_files() {
  grep_tracked_files_for_tokens --name-only | while IFS= read -r file; do
    rewrite_file "$file"
  done
}

derive_names() {
  module=$1
  slug=${module##*/}
  db_name=$(printf '%s\n' "$slug" | tr '-' '_')
  display_name=${2:-$(title_case "$slug")}
}

create_scratch_files() {
  program=$(mktemp)
  scratch=$(mktemp)
  trap 'rm -f "$program" "$scratch"' EXIT
}

sweep() {
  require_unrenamed_template
  derive_names "$@"
  require_valid_slug "$slug"
  create_scratch_files
  sed_program >"$program"
  rewrite_tracked_files
  say "renamed to module $module, slug $slug, database $db_name, display name $display_name."
}

print_next_steps() {
  say "no template token is left. Next steps:"
  echo "  make setup"
  echo "  make github-setup"
  echo "  commit the result"
}

verify() {
  if grep_tracked_files_for_tokens --line-number >&2; then
    fail "the template tokens above are left, replace them by hand."
  fi
  print_next_steps
}

main() {
  case ${1:-} in
    sweep)
      [ $# -eq 2 ] || [ $# -eq 3 ] || usage
      shift
      sweep "$@"
      ;;
    verify)
      [ $# -eq 1 ] || usage
      verify
      ;;
    *) usage ;;
  esac
}

main "$@"
