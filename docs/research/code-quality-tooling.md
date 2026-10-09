# Code-quality and architecture-enforcement tooling for Go

Research for issue #6 (map #1). Facts and trade-offs only; decisions belong to a later grilling ticket. Researched 2026-10-09.

Source key: [GL] golangci-lint docs (golangci-lint.run), [GO] go.dev, [R] revive docs, [T] templ.guide, [LH] lefthook README, [PC] pre-commit.com, [VN] varnamelen README, [AL] go-arch-lint README, [PT] prettier-plugin-tailwindcss README, [VC] pkg.go.dev govulncheck. Where a fact came from the GitHub releases API it is marked [GH]. Items marked "unverified" could not be confirmed on a primary page.

## Current versions

| Tool | Version | Date | Source |
|---|---|---|---|
| Go | 1.27.2 (1.27.0 released 2026-08-19) | 2026-10-08 | [GO] release history |
| golangci-lint | v2.14.0 | 2026-09-24 | [GH] |
| go-arch-lint | v1.19.0 | 2026-09-07 | [GH] |
| gofumpt | v0.12.0 | 2026-09-07 | [GH] |
| templ | v0.3.1070 | 2026-10-04 | [GH] |
| lefthook | v2.2.1 | 2026-10-08 | [GH], [LH] |

## 1. golangci-lint v2

### Config format [GL config file]
- File names searched: `.golangci.yml`, `.golangci.yaml`, `.golangci.toml`, `.golangci.json` (cwd, then parents, then home).
- `version: "2"` is required and the only allowed value.
- `linters.default`: `standard` (default), `all`, `none`, `fast`. `linters.enable` / `disable` adjust it. Per-linter options live only in the file under `linters.settings`.
- `linters.exclusions`: `generated` (`strict` | `lax` | `disable`), `warn-unused`, `presets`, `rules` (match on `path`, `path-except`, `linters`, `text`, `source`), `paths`, `paths-except`.
- `formatters`: `enable`, `settings`, `exclusions`. Formatters are a separate top-level section in v2, not linters.
- `issues`: `max-issues-per-linter`, `max-same-issues`, `new`/`new-from-*` (report only changed code), `fix`.
- `run`: `timeout`, `tests`, `build-tags`, `go`, `concurrency`, and others.
- Exclusion presets: `comments`, `common-false-positives`, `legacy`, `std-error-handling` [GL false-positives].
- `//nolint:linter // reason` syntax: no spaces around `:`; a trailing explanation is allowed [GL false-positives]. `nolintlint` can require a specific linter, an explanation, and no unused directives [GL linters/configuration].
- Generated files: the `generated` exclusion mode exists [GL config file]. The pages read do not say which header markers are recognised; whether templ's `_templ.go` header is matched in `strict`/`lax` mode is unverified and should be tried in a spike.

### Installation: conflict with the "tools via `go.mod`" constraint [GL install/local]
- Recommended: binary install (script, Homebrew, Docker, mise, etc.).
- `go install` is listed under "Install from Sources" with a warning that it is "not guaranteed to work".
- The page states: "We don't recommend using `go tool`". If used anyway, it suggests isolating golangci-lint in a dedicated module file or module, and never updating its dependencies manually.
- So the standing constraint (tools pinned via `tool` directives) is officially discouraged for golangci-lint specifically. Options: (a) separate tools module (or `go tool -modfile=...`) holding only golangci-lint, (b) pin the binary version in CI/mise/Docker and in a Makefile/task variable, (c) accept the `tool` directive in the main `go.mod` against the upstream recommendation. Trade-off: (a) keeps pinning inside Go tooling without polluting the app's dependency graph; (b) matches upstream but adds a non-Go version source.
- Tool directive mechanics [GO managing-dependencies]: `go get -tool <pkg>` adds a `tool` directive plus `require`; run with `go tool <name>`; `go install tool` installs all; tool requirements take part in MVS and respect `replace`/`exclude`. Due to module pruning, a dependency's own tool requirements usually do not become yours.

### Standard default set
`linters.default: standard` is the default [GL config file]. The pages read do not list its members (`golangci-lint linters` prints them); commonly errcheck, govet, ineffassign, staticcheck, unused (unverified here).

## 2. Mapping coding principles to enforcement

Defaults below are from [GL linters/configuration] unless noted.

### Small types
| Mechanism | Detail |
|---|---|
| revive `max-public-structs` | Max public structs per file (default 5). Limits types per file, not fields per type. |
| revive `file-length-limit` | `max`, `skip-comments`, `skip-blank-lines`. Proxy for small units. |
| `interfacebloat` | Limits methods per interface. |
| `funcorder` | Orders constructors, exported then unexported methods (consistency, not size). |
| `recvcheck` | Receiver-type consistency. |
| revive `nested-structs` | Flags inline struct definitions. |
| `gocritic hugeParam` | Param size threshold (default 80 bytes). |
| `embeddedstructfieldcheck` | Embedded fields first. |

Gap: no linter found in the pages read caps fields per struct or methods per concrete type. "Small class" is only approximated by file length, types-per-file and interface size.

### Single-responsibility, small functions
| Mechanism | Settings (defaults) |
|---|---|
| `funlen` | `lines` 60, `statements` 40, `ignore-comments` true; negative disables. |
| `cyclop` | `max-complexity` 10; `package-average` (off at 0.0). |
| `gocyclo` | `min-complexity` 30 (docs recommend 10-20). |
| `gocognit` | `min-complexity` 30 (docs recommend 10-20). |
| `nestif` | `min-complexity` 5. |
| revive `function-length` | Two args statements/lines, defaults 50 / 75 [R]. |
| revive `cognitive-complexity`, `cyclomatic` | defaults 7 and 10 [R]. |
| revive `argument-limit`, `function-result-limit` | defaults 8 and 3 [R]. |
| revive `max-control-nesting` | default 5 [R]. |
| revive `flag-parameter` | Flags boolean params that steer control flow [R]. |
| gocritic `nestingReduce`, `tooManyResultsChecker`, `unnamedResult` | Settings listed in [GL]; `bodyWidth` 5, `maxResults` 5. |
| `nakedret`, `nonamedreturns` | Return clarity. |
| `unparam` | Unused parameters. |

Overlap: cyclop, gocyclo and revive `cyclomatic` all measure cyclomatic complexity; gocognit and revive `cognitive-complexity` both measure cognitive. Enabling several duplicates findings. Single responsibility itself is not mechanically checkable; length and complexity are proxies.

### Design patterns where they fit
Not mechanically enforceable. Partial guard rails only: `ireturn` ("Accept Interfaces, Return Concrete Types"), `iface` (interface pollution), `inamedparam`, `containedctx`, `gochecknoglobals`, `gochecknoinits` (discourage global-state patterns). Whether a pattern fits is a review/design judgement.

### No duplication
- `dupl`: token-based, `threshold` default 150 (docs example 100). Detects fragments, not semantic duplication.
- `goconst`: repeated strings. `mnd` / revive `add-constant`: magic numbers/strings. `dupword`: duplicated words in comments.
- Caveat: generated code (templ output) can trigger `dupl`; relies on generated-file exclusion (see section 1).

### Comments explain only why
| Mechanism | Detail |
|---|---|
| `godot` | Comment style (period, capital); `scope` default `declarations`. Style, not content. |
| `godoclint` | godoc rules; default `basic` (`pkg-doc`, `single-pkg-doc`, `start-with-name`, `deprecated`), plus `require-doc` and `max-len` options. `require-doc` ignores unexported by default; enabling it for unexported identifiers would conflict with "comment only why". |
| revive `exported`, `package-comments` | Push toward doc comments on exported symbols. The `comments` exclusion preset suppresses these (and staticcheck ST1000/1020/1021/1022). |
| revive `comments-density` | Minimum density only (inverse of the goal). |
| revive `comment-spacings` | `//text` formatting. |
| `godox` | Flags TODO/FIXME-style keywords. |
| `nolintlint` | Can require an explanation on every `nolint` (a "why" comment). |
| gocritic `commentedOutCode` | Setting `minLength` 15 exists per [GL]; the linters page does not list it as a separate check, and whether it is in the default checker set is unverified. |

Cannot be enforced: whether a comment says why or restates what. Tooling checks presence, format and commented-out code; reviewers (or an LLM review step) must judge content. Linters that mandate comments (`exported`, `godoclint require-doc`, `comments-density`) work against the principle; `godot`, `comment-spacings`, `godox`, `nolintlint` are hygiene only.

### Well-named identifiers
| Mechanism | Detail |
|---|---|
| revive `var-naming`, `receiver-naming` (`max-length`), `import-alias-naming` (default regex `^[a-z][a-z0-9]{0,}$`), `confusing-naming`, `package-directory-mismatch`, `redundant-import-alias` | [R] |
| `varnamelen` | Name length vs scope: `max-distance` 5, `min-name-length` 3, `ignore-names` etc. [VN]. |
| `predeclared` | Shadowing of predeclared identifiers. |
| `errname` | Sentinel errors `Err*`, error types `*Error`. |
| `importas` | Consistent import aliases (`alias`, `no-unaliased`, `no-extra-aliases`). |
| `inamedparam` | Named interface params. |
| `staticcheck` ST checks | Naming/style checks exist, but the linters page did not enumerate them (unverified). |
| `forbidigo` | Forbids identifiers matching configured patterns (for example vague names). |

Cannot be enforced: semantic quality of a name (intention-revealing). Lint only checks length, casing, consistency.

### Separation of concerns / modularity
| Mechanism | Detail |
|---|---|
| `internal/` | Compiler-enforced Go rule: only code in the parent tree may import it. No tool needed (not re-fetched this session). |
| `depguard` | Per-rule `files` globs, `allow`/`deny` of import paths, `list-mode` (`original`/`strict`/`lax`); with no rules the default is stdlib-only for all files. Can express "domain must not import `net/http` or the db driver". |
| `gomodguard_v2` | Allow/block lists for direct module dependencies (`match-type`, version constraints); `gomodguard` is deprecated in its favour. Fits the "stdlib-first; dependencies must be justified" constraint as an allow-list. |
| `gomoddirectives` | Restricts `replace`/`retract`/`exclude` in go.mod. |
| go-arch-lint | See section 3. |
| revive `package-directory-mismatch`, `max-public-structs`, `file-length-limit` | Structure hygiene. |
| `gochecknoglobals`, `gochecknoinits`, `containedctx` | Reduce hidden coupling. |

Cannot be enforced: whether the chosen boundaries are right; only that declared boundaries hold.

## 3. Architecture enforcement

### depguard (inside golangci-lint) [GL]
Rules keyed by name; each has `files` (globs, `$all`, negation with `!`), `allow`, `deny` (`pkg` plus `desc`), and `list-mode`. Special `$gostd`. Layer rules are hand-written per-directory deny lists; there is no component concept.

### go-arch-lint (fe3dback) [AL]
- Standalone CLI, v1.19.0 (2026-09-07 [GH]). Install from source needs Go 1.25+. MIT; about 592 stars, 185 commits, 17 open issues at read time. Released within the last month [GH].
- Config `.go-arch-lint.yml` (`--arch-file` to override): `version` (example uses 3), `workdir`, `components` (`in:` path patterns, `*` one level, `**` many), `commonComponents`, `deps` with `mayDependOn`. A vendors section exists upstream but the page read did not describe it.
- Commands: `check` (flags `--project-path`, `--arch-file`, `--max-warnings` default 512, `--json`/`--output-type`; exit 1 on warnings), `graph` (dependency graph).
- Compared with depguard: a component-level allow-graph in one file, flags undeclared dependencies (the README example is a handler receiving a repository in `main.go`), and renders a graph. Cost: a second tool and config outside golangci-lint, needing its own CI step, hook entry and pinning (as a Go module it should work with a `tool` directive; unverified for this repo).

### `internal/` boundaries
Compiler-enforced, zero config, but coarse (module-external vs internal). Cannot express layer rules inside `internal/`.

Options: (1) `internal/` only; (2) `internal/` + depguard (one tool, per-layer deny lists); (3) `internal/` + go-arch-lint (explicit component graph, extra tool).

## 4. Formatters [GL formatters, formatters/configuration]
golangci-lint v2 formatters (all autofix): `gci`, `gofmt`, `gofumpt`, `goimports`, `golines`, `swaggo`. `golangci-lint fmt` runs them; the editor setup in [GL integrations] uses `golangci-lint fmt --stdin`.
- `gofmt`: `simplify` (the `-s` flag), `rewrite-rules`.
- `gofumpt`: stricter, backward-compatible with gofmt; settings `module-path`, `extra` (`group-params`, `clothe-returns`, `balance-calls`). `extra-rules` is deprecated.
- `goimports`: `local-prefixes`.
- `gci`: `sections` (standard, default, `prefix(...)`, blank, dot, alias, localmodule), `custom-order`. gci and goimports both manage imports; one is enough (goimports `local-prefixes` is simpler, gci gives finer groups).
- `golines`: `max-len` (page-stated default 100); would also make `lll` redundant.
Typical pairing: `gofumpt` (superset of gofmt) + `goimports` or `gci`. Running both `gofmt` and `gofumpt` is redundant.

## 5. govulncheck [GO tutorial, VC]
- Install: `go install golang.org/x/vuln/cmd/govulncheck@latest`. It is a Go module, so a `tool` directive is possible (not shown on these pages).
- `govulncheck ./...` (source mode, default) reports vulnerabilities reachable from your code, with call stacks, and separately informational findings in imported but uncalled packages. Modes: `source`, `binary` (`-mode binary`, no call stacks), `extract`.
- Flags: `-tags`, `-test`, `-db`, `-show traces|verbose`, `-json`, `-format sarif|openvex`.
- Exit codes: 0 when none found; non-zero when found; with `-json`, `-format sarif` or `-format openvex` the exit code is 0 regardless. For CI gating use text mode, or parse the JSON/SARIF yourself.
- It reports against the Go toolchain in use, so stale patch versions show stdlib vulnerabilities. The tutorial page still labels it experimental.
- No first-party GitHub Action is mentioned on the pages read. Network access makes it a fit for CI or `pre-push` rather than `pre-commit`; that is a design choice, not a documented rule.
- gosec (a golangci-lint linter: "inspects source code for security problems") is complementary static analysis, not a dependency scanner.

## 6. Git hooks

### lefthook [LH]
- Single Go binary, no runtime needed. MIT. v2.2.1 (2026-10-08).
- Install options on its README: `go install github.com/evilmartians/lefthook/v2@v2.2.1` (README states this requires Go >= 1.27), `go get -tool github.com/evilmartians/lefthook/v2@v2.2.1` (fits the `tool` directive constraint), npm, gem, pipx, Homebrew and others.
- `lefthook install` writes the git hooks. Config is `lefthook.yml` (also .yaml/.toml/.json/.jsonc; `lefthook-local` for personal overrides) with hooks (`pre-commit`, `pre-push`) containing `jobs` (or `commands`). Job keys seen: `name`, `run`, `glob`, `root`, `fail_text`, `stage_fixed`; hook-level `parallel`, `piped`. The README example uses `{all_files}`; `{staged_files}` and `stage_fixed` semantics were not confirmed (docs sub-pages returned 404).
- Each fresh clone must run `lefthook install` (or a task wrapping it); this could be part of the rename/setup task from the map.

### pre-commit [PC]
- Python-based (`pip install pre-commit` or a `.pyz` zipapp). Config `.pre-commit-config.yaml` with `repos`/`rev`/`hooks`; `pre-commit install` writes `.git/hooks/pre-commit`.
- Go hooks: `language: golang` builds hook repos with `go install ./...` in an isolated GOPATH; `repo: local` with `language: unsupported` (formerly `system`) uses tools the developer installed, with no environment.
- Adds a Python dependency to a Go repo and a second pinning mechanism (`rev`); local hooks do not use the `go.mod` tool directive.

Comparison: lefthook needs no non-Go runtime and can be pinned in go.mod. pre-commit's hook catalogue is large but mostly redundant once golangci-lint covers Go.

Hooks are bypassable (`--no-verify`); CI is the real gate (CI pipeline is a separate map item).

## 7. templ and Tailwind formatting

### templ [T cli, ide-support, installation]
- `templ fmt .` formats `.templ` files; `templ fmt -fail .` exits 1 if any file needed changes (CI check). `script`/`style` blocks are formatted via prettier/prettierd/npx when on PATH.
- `templ generate` produces Go (`-watch`, `-lazy`, `-path`, `-f`, `-include-version`, `-include-timestamp`). The last two control generated-file header contents, which affects diff noise if generated code is committed.
- Install: `go install github.com/a-h/templ/cmd/templ@latest` or `go get -tool github.com/a-h/templ/cmd/templ@latest` then `go tool templ` (matches the constraint; pinning via go.mod).
- golangci-lint does not analyse `.templ` source, only generated `_templ.go` (which must be excluded or recognised as generated). No templ-specific linter was found on the pages read.
- Whether to commit generated `_templ.go` is a separate decision (affects lint exclusion and a CI `generate` + diff check).

### Tailwind class sorting [PT]
- `prettier-plugin-tailwindcss` sorts classes into Tailwind's recommended order and removes duplicates and extra whitespace. Needs prettier v3+, ESM-only. Config: `tailwindStylesheet` (v4), `tailwindConfig` (v3), `tailwindAttributes`, `tailwindFunctions`.
- The README mentions neither templ nor Go; `.templ` support is undocumented there. Using it would need Node tooling and a way to feed `.templ` to prettier (unverified). Whether automated class sorting for templ is feasible is an open question for a spike; the alternative is no automated sorting.
- Tailwind editor completion in templ: map the `templ` language to `html` in Tailwind's editor settings; the templ page refers to `tailwind.config.js` including `.templ` and `.go` content paths (Tailwind v3 style; v4 CSS-first config not covered there).

## 8. Editor integration
- VS Code (Go extension) [GL integrations]: `go.lintTool: "golangci-lint"` (or `"golangci-lint-v2"` via `Go: Install/Update Tools`), `go.lintFlags: ["--path-mode=abs", "--fast-only"]`, formatting via `go.formatTool: "custom"`, `go.alternateTools: {"customFormatter": "golangci-lint"}`, `go.formatFlags: ["fmt", "--stdin"]`. Warning on that page: without `--fast-only` the editor can freeze.
- GoLand: built-in golangci-lint support from 2025.1; IntelliJ IDEA needs the Go Linter plugin.
- Neovim/Vim/Emacs: `golangci-lint-langserver` is the listed LSP; no Neovim-specific settings on that page.
- templ [T ide-support]: VS Code extension (Marketplace/Open VSX; needs `templ` on PATH; format on save and Tailwind completion configurable), Neovim via `templ lsp` with lspconfig/Mason plus tree-sitter, JetBrains templ plugin, Helix (unstable branch), Emacs `templ-ts-mode`. `templ lsp` starts its own gopls by default (`-gopls-remote` shares one).
- Shipping editor config in the repo (`.vscode/settings.json`, `.editorconfig`) is a choice, not a requirement from any source read.

## 9. Open points for the grilling ticket
1. golangci-lint pinning: separate tools module, externally pinned binary, or `tool` directive against upstream advice.
2. Complexity linter set: one family (for example cyclop + gocognit) vs revive-only; thresholds (funlen defaults 60/40 vs stricter).
3. Whether any comment-mandating linters (`godoclint require-doc`, revive `exported`) are enabled, given "comments only explain why".
4. depguard vs go-arch-lint (or neither) for layer rules.
5. Formatter combination (gofumpt + goimports vs gci).
6. Hook runner (lefthook vs pre-commit vs none) and which checks sit in pre-commit vs pre-push vs CI only.
7. templ: `templ fmt -fail` in CI, generated file exclusion, committing `_templ.go`.
8. Tailwind class sorting for templ: unverified prettier-plugin support; spike or skip.
9. Remaining mechanically unenforceable: single responsibility, fitting design patterns, intention-revealing names, comment content (why vs what), and small types beyond file/interface size proxies.
