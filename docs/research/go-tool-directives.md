# Pinning dev tools with go.mod `tool` directives

Research for issue [#3](https://github.com/DannyFestor/go-template-html/issues/3). Parent map: [#1](https://github.com/DannyFestor/go-template-html/issues/1).

Facts only. The choice is for the grilling ticket.

Local checks were run with Go 1.27.2 (`darwin/arm64`) in a scratch directory outside the repo.

## Summary

- `tool` directives (Go 1.24+) pin a tool's version in `go.mod` and run it with `go tool <name>`. Tool requirements become normal module requirements.
- Adding one tool to the app's `go.mod` pulls its dependencies into it as `// indirect` requires. The local templ test added 15 indirect requires to a fresh module.
- The `-modfile` flag keeps those requirements in a separate file. The app's `go.mod` stayed unchanged in the local test.
- golangci-lint's own docs say tool-directive installs are not guaranteed to work, and they recommend binary installs. Their `go tool` section says they do not recommend `go tool`, and if you use it, to isolate it in a dedicated module or `-modfile`.
- The Tailwind standalone CLI is a non-Go binary, so `tool` directives cannot pin it. Options are mise, a download script, or Taskfile.
- mise can pin golangci-lint, templ, and the Tailwind binary, but it does not replace `tool` directives for Go tools. Its lockfile records checksums only for some backends.

## 1. How `tool` directives work

Sources: [Go modules reference, `tool` directive](https://go.dev/ref/mod#go-mod-file-tool), [Managing dependencies, "Tool dependencies"](https://go.dev/doc/modules/managing-dependencies), [Go 1.24 release notes, Tools](https://go.dev/doc/go1.24).

- Syntax is `tool <package-path>`, one per line, or grouped in `tool ( ... )`.
- `go get -tool <pkg>@<version>` adds the `tool` line and the `require` lines. `go get -tool @none` removes it.
- Run with `go tool <name>`, where `<name>` is the last path element (major-version suffix excluded). Use the full package path when names collide or clash with a Go-distribution tool.
- `go tool` with no args lists tools. `go get tool` upgrades all tools at once. `go install tool` installs all of them to `GOBIN`.
- Tool requirements take part in MVS and respect `require`, `replace`, and `exclude`. Module pruning means a dependency's tool requirements usually do not leak into your build list.
- A hand-written `tool` line needs a matching `require`. `go mod tidy` adds missing ones.
- Pre-1.24 workaround: a blank import in a build-ignored file (the old `tools.go` pattern).

Local check (`go get -tool github.com/a-h/templ/cmd/templ@latest` in a fresh module):

```
tool github.com/a-h/templ/cmd/templ
require ( ... 15 lines, all // indirect, including golang.org/x/tools, golang.org/x/mod, fsnotify, ... )
```

`go tool templ version` ran and reported `v0.3.1070`. The `go.sum` grew to 30 lines.

## 2. Candidate tools

| Tool | Official install route | `tool` directive route | Source |
|---|---|---|---|
| golangci-lint v2 | Binary install script (recommended), Homebrew, Docker, mise (unofficial) | Documented as "not guaranteed to work"; `go tool` use "not recommended" | [Install, local](https://golangci-lint.run/docs/welcome/install/local/) |
| templ | `go install ...@latest`, GitHub binaries, Nix, Docker | Documented: `go get -tool github.com/a-h/templ/cmd/templ@latest`, then `go tool templ` | [templ install](https://templ.guide/quick-start/installation) |
| govulncheck | `go install golang.org/x/vuln/cmd/govulncheck@latest` | Not documented by go.dev; `go tool` route works in principle (not tested here) | [go.dev govulncheck tutorial](https://go.dev/doc/tutorial/govulncheck) |
| Tailwind standalone CLI | GitHub release binaries (not Go) | Not applicable | [Tailwind CLI docs](https://tailwindcss.com/docs/installation/tailwind-cli), [release v4.3.3](https://github.com/tailwindlabs/tailwindcss/releases/latest) |

Not checked in this research: sqlc, migration CLIs, live-reload tools (air and similar). Confirm their install docs before deciding.

golangci-lint official text (from [Install, local](https://golangci-lint.run/docs/welcome/install/local/)):

- "We recommend using binary installation."
- "Using `go install`/`go get`, 'tools pattern', and `tool` command/directives installations aren't guaranteed to work."
- The tools-pattern and `tool` routes can change other tools' or the project's dependencies.
- "We don't recommend using `go tool`." If you still want it, isolate golangci-lint in a dedicated module or module file.
- Its `-modfile` example: `go get -tool -modfile=golangci-lint.mod github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.14.0`.

## 3. Known problems

**Dependency bleed into the app's `go.mod`.** Each `tool` adds its dependency graph as indirect requires. Local result: templ alone added 15 indirect requires. golangci-lint's docs say the same in general terms (tool dependencies can modify the project's dependencies).

**Separate tool modules via `-modfile`.** Local check, on a fresh module with a copy of `go.mod` and `go.sum` seeded into `tools/`:

```
go get -tool -modfile=tools/tools.mod github.com/a-h/templ/cmd/templ@latest
go tool -modfile=tools/tools.mod templ version    # v0.3.1070
```

- The app's `go.mod` was unchanged (`diff` clean).
- `tools/tools.mod` got the `tool` block and the indirect requires.
- Caveats: `-modfile` must be passed on every `go tool` call, so a Taskfile or mise task has to wrap it. The `-modfile` target must exist first; `go get -tool -modfile` on a missing file fails with `open ...: no such file or directory`. Seed it from `go.mod` and `go.sum`.
- golangci-lint documents this route (Method 1: dedicated module file; Method 2: dedicated module linked through a Go workspace).

**Other notes.**
- Tools must be run from inside the module (or a workspace containing it).
- Name clashes need the full package path.

## 4. Non-Go binaries (Tailwind standalone CLI)

Sources: [Tailwind CLI docs](https://tailwindcss.com/docs/installation/tailwind-cli) (standalone-binary link only, no asset names), [GitHub release v4.3.3 via API](https://api.github.com/repos/tailwindlabs/tailwindcss/releases/tags/v4.3.3).

- Latest release at research time: `v4.3.3`, published 2026-07-16.
- Assets: `tailwindcss-{linux-arm64, linux-arm64-musl, linux-x64, linux-x64-musl, macos-arm64, macos-x64}` and `tailwindcss-windows-x64.exe`, plus `sha256sums.txt`.
- The release publishes `sha256sums.txt`, so a download script can verify checksums.
- Pinning means choosing the release tag in the download URL. The Tailwind docs gave no pinning guidance.

Options:

1. **mise.** Declare the tool in `mise.toml`. The [mise docs](https://mise.jdx.dev/dev-tools/) describe the `github:` backend for release-asset tools, with `matching` and `rename_exe` options. Not verified locally: whether the Tailwind assets resolve cleanly through it.
2. **Download script.** Fetch the release asset for the platform, check it against `sha256sums.txt`, and place it in a project bin directory. Fully explicit; the script must handle platform detection and musl.
3. **Taskfile.** A task runner that calls either of the above. Not researched in depth here.

## 5. Can mise alone replace `tool` directives?

Sources: [mise dev tools](https://mise.jdx.dev/dev-tools/), [mise lockfile](https://mise.jdx.dev/dev-tools/mise-lock.html), [golangci-lint install, local](https://golangci-lint.run/docs/welcome/install/local/).

- mise pins tool versions in `mise.toml` `[tools]` and resolves them through backends. The `go:` backend installs Go binaries (e.g. `"go:github.com/mikefarah/yq/v4"`).
- golangci-lint's install page lists mise as a route (`mise use -g golangci-lint@v2.14.0`), flagged as not officially maintained.
- Lockfile: `mise lock` records the resolved version for every tool. It records a URL and checksum per platform for aqua, github, gitlab, forgejo, http, and s3. The `go` backend records the version only. `mise install --locked` installs from the lock.

Trade-offs, without a recommendation:

- mise handles all tools in one place, including non-Go ones (Tailwind), with no `go.mod` pollution.
- mise does not put a Go tool's version in `go.mod`, so the app's Go code cannot rely on it being present through `go tool`.
- For Go tools, the `go` backend's lock is version-only. The `tool` directive records the version in `go.mod` and `go.sum`, which the Go toolchain checksums.
- mise is a second dependency on every machine that builds the project.

## Open items for the grilling ticket

- Pick between: `tool` directives in the app `go.mod`; `tool` directives via `-modfile`; mise for everything; mise for non-Go tools plus Go tools pinned some other way.
- Confirm sqlc, migration CLI, and live-reload tools against their install docs.
- Decide whether golangci-lint is pinned by mise or by binary install, given its docs.
