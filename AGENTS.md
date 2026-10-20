## Navigation

- Commands: `make help`. Read a target's recipe only to change it.
- `docs/toolchain.md`: per-tool pitfalls, agent-hook rules, CI. List its sections with `grep -n '^## ' docs/toolchain.md`, then read the whole section for every tool, config and test kind you touch before editing.
- `docs/upgrading.md`: bumping a pinned tool.
- Linked worktree: run `make worktree-setup` first.

## Agent skills

### Issue tracker

Issues are tracked in GitHub Issues for this repo, using the `gh` CLI. See `docs/agents/issue-tracker.md`.

### Triage labels

Uses the five default triage labels: `needs-triage`, `needs-info`, `ready-for-agent`, `ready-for-human`, `wontfix`. See `docs/agents/triage-labels.md`.

### Domain docs

Single-context: one `GLOSSARY.md` and `docs/adr/` at the repo root. See `docs/agents/domain.md`.
