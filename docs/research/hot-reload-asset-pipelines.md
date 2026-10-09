# Hot reload and asset pipelines for Go SSR

Research for issue #2. Facts and trade-offs only; no decisions. Researched 2026-10-09.

Question: how do Go SSR apps get a Laravel-Vite-style dev loop (edit a template, Go file or CSS and the browser refreshes) with Tailwind CSS, and what does each option cost?

Versions seen (GitHub releases): templ v0.3.1070 (2026-10-04), air v1.67.4 (2026-08-01), wgo v0.7.1 (2026-08-20), Tailwind CSS v4.3.3 (2026-07-16), Vite 8.x.

## 1. templ `generate --watch --proxy`

Sources: [templ live reload docs](https://github.com/a-h/templ/blob/main/docs/docs/09-developer-tools/03-live-reload.md), [CLI docs](https://github.com/a-h/templ/blob/main/docs/docs/09-developer-tools/01-cli.md).

- One command does three jobs: `templ generate --watch --proxy="http://localhost:8080" --cmd="go run ."`.
  - `--watch` watches `*.templ` and `*.go` files and regenerates Go code when `*.templ` files change.
  - `--cmd` runs a command when Go code changes (including Go code inside `.templ` files). The docs name `go run .`, `go build`, `air` and `wgo` as options. The CLI page says it runs via the system shell.
  - `--proxy` starts an HTTP proxy in front of the app and injects JavaScript before `</body>`; the browser reloads when the app restarts. Default proxy address is `127.0.0.1:7331`.
- Injection requirements: the HTML must contain a `<body>` tag, the response must be `text/html`, and it must be uncompressed or a supported encoding such as gzip.
- Flags: `--proxybind` (default `127.0.0.1`), `--proxyport` (default 7331), `--proxy-tls-crt` / `--proxy-tls-key` (HTTPS on the proxy; used together; require `--proxy`).
- `--notify-proxy` triggers a reload from outside the watcher, so another tool (for example a CSS rebuild) can request a browser reload. `--proxyport` targets a non-default proxy port for it.
- Text-only fast path: for changes to HTML or text in `.templ` files, templ writes text files to a `tmp` directory that generated code reads when `TEMPL_DEV_MODE` is enabled, so the server is not restarted. Only changes to Go code in `.templ` files restart it. The docs do not say how `TEMPL_DEV_MODE` is enabled.
- `--open-browser` and `--watch-pattern` are not documented on these pages; treat them as unverified.

## 2. air and wgo

### air

Source: [air README](https://github.com/air-verse/air).

- Rebuilds and restarts the app on file change. Configured with `.air.toml` (`air init` writes a default).
- Built-in proxy with browser reload: `[proxy] enabled = true`, `proxy_port` (the port you open) and `app_port` (where the app listens). It injects a script before `</body>` in HTML responses and reloads after each successful rebuild. `app_start_timeout` defaults to 5000 ms.
- Static files must match `include_dir`, `include_ext` or `include_file`, or edits do not trigger a reload. `delay` (debounce) defaults to 1000 ms.
- Install as a Go tool: `go get -tool github.com/air-verse/air@latest`, run via `go tool air`. `go install` and `go get -tool` require Go 1.25 or higher.
- templ's docs show an `.air.toml` with `cmd = "templ generate && go build -o ./tmp/main ."`, air proxy on `proxy_port = 8383`, app on `app_port = 8282`.
- templ's [multi-tool guide](https://github.com/a-h/templ/blob/main/docs/docs/09-developer-tools/04-live-reload-with-other-tools.md) combines Tailwind, esbuild, air and templ's proxy through a Makefile running five watch processes (`make -j5`). `templ generate --notify-proxy` sends the reload event. It advises serving static assets from disk with `http.Dir` (not embedded `http.FS`) and sending `Cache-Control: no-store` in development.

### wgo

Source: [wgo README](https://github.com/bokwoon95/wgo).

- File watcher that reruns commands: `wgo run ./cmd/app`; `-file` / `-xfile` / `-dir` / `-xdir` regex filters; `-debounce` (default 300 ms).
- `::` chains commands (the next runs only if the previous succeeds); `wgo` after `::` starts a parallel watcher, e.g. `wgo run main.go :: wgo -file .scss sass ...`. Chained commands are not run through a shell.
- No browser reload: "live reload" means restarting the process. A browser refresh must come from templ's proxy, custom SSE/websocket code, or another tool.
- templ's example: `wgo -file=.go -file=.templ -xfile=_templ.go templ generate :: go run main.go`. Excluding `_templ.go` avoids reload loops.
- Install: `go install github.com/bokwoon95/wgo@latest`. Last push 2026-08-20; a small project (about 577 stars at fetch time).

### Custom SSE/websocket injection

air's proxy and templ's proxy both rewrite HTML to add a listener script. Doing it in-app means a dev-only endpoint plus a script tag in the layout. No first-party spec exists for this; the template would own that code.

## 3. Tailwind standalone CLI

Sources: [Tailwind CLI install docs](https://tailwindcss.com/docs/installation/tailwind-cli), [standalone CLI announcement](https://tailwindcss.com/blog/standalone-cli), [v4.3.3 release](https://github.com/tailwindlabs/tailwindcss/releases/tag/v4.3.3), [source detection docs](https://tailwindcss.com/docs/detecting-classes-in-source-files).

- Docs: "The CLI is also available as a standalone executable if you want to use it without installing Node.js." The npm route (`npm install tailwindcss @tailwindcss/cli`) needs Node.
- v4.3.3 release assets: `tailwindcss-linux-arm64`, `-linux-arm64-musl`, `-linux-x64`, `-linux-x64-musl`, `-macos-arm64`, `-macos-x64`, `-windows-x64.exe`, plus `sha256sums.txt`.
- Usage matches the npm CLI: `tailwindcss -i input.css -o output.css --watch`. `--minify` for production appears in the announcement post, which predates v4; the v4 install page does not document it, so confirm with `--help`.
- The announcement says first-party plugins (`@tailwindcss/forms`, `@tailwindcss/typography`) are bundled. That post describes the v3-era build; no equivalent statement was found for v4.
- v4 scans the project automatically, skipping `.gitignore`d files, `node_modules`, binaries, CSS and lockfiles. `.templ` and `.html` files are picked up without config; `@source` adds paths. Classes built by string concatenation are not detected; `@source inline(...)` safelists them.
- Pinning: the binary is not a Go module, so a `go.mod` `tool` directive cannot manage it. It is downloaded per platform; `sha256sums.txt` allows verification.
- The CLI does not reload the browser. It writes the CSS file; a reload needs air or templ's proxy (`--notify-proxy`), or a manual refresh with `no-store` caching.

## 4. Vite with a Go manifest helper

Sources: [Vite backend integration](https://vite.dev/guide/backend-integration), [Vite guide](https://vite.dev/guide/).

- Dev: inject `<script type="module" src="http://localhost:5173/@vite/client">` and the entry script into the server-rendered HTML; either proxy asset requests to Vite or set `server.origin`. `server.cors.origin` must allow the browser-facing origin.
- Production: `build.manifest: true` writes `.vite/manifest.json`. Entries map a source path to `file`, `css`, `imports`, `isEntry` and more. The backend renders, in order: stylesheet links for the entry's `css`, the `css` of imported chunks (recursively), the entry `file` as `<script type="module">`, optional `modulepreload` links. Vite documents a pseudo-implementation of the chunk walk; a Go helper would implement it (a template function reading the manifest from disk or `embed`).
- Provides hashed filenames, JS/TS bundling and HMR for JS/CSS. It does not hot-reload server-rendered HTML; template edits still need a reload hook (templ proxy or similar).
- Node cost: Vite requires Node.js 20.19+ or 22.12+. That adds `package.json`, a lockfile, `node_modules`, Node in the dev Docker image and CI, and a second dependency ecosystem. Tailwind is also installable through npm, so the standalone binary becomes optional.
- Whether Node is acceptable is undecided (map #1). Node buys hashed assets, a manifest, JS bundling and HMR. The no-Node path (Tailwind standalone binary, plain CSS/JS served by Go) avoids the toolchain but leaves hashing and cache-busting to Go code the template would own.
- Third-party Go libraries for reading Vite manifests were not evaluated; the manifest format is documented.

## 5. templ vs `html/template` in live reload

Sources: [templ live reload docs](https://github.com/a-h/templ/blob/main/docs/docs/09-developer-tools/03-live-reload.md), [html/template docs](https://pkg.go.dev/html/template), [Go tool dependencies](https://go.dev/doc/modules/managing-dependencies).

| Aspect | templ | `html/template` |
|---|---|---|
| Template change | `templ generate` regenerates `*_templ.go`; text-only edits skip a restart in dev mode; Go-code edits rebuild and restart | Files are read at parse time; no restart needed only if the app builds a fresh template set per request or on change in dev |
| Checking | Type-checked Go at compile time | Errors at parse or execute time |
| Re-parse | Not applicable | `ParseFS` / `ParseGlob` / `ParseFiles` read from an `fs.FS` or disk. They, and `Clone`, return an error once the template has been executed, so dev reload means a fresh template set, not mutating the old one |
| Watcher | Required (`templ generate --watch`, or `go tool templ` plus air/wgo) | Only a Go watcher for `.go` changes and a browser-reload trigger |
| Extra tool | `templ` CLI, pinnable as a `go.mod` `tool` directive (Go 1.24+, run `go tool templ`) | None; stdlib |
| Browser reload | templ proxy built in | air proxy or custom SSE |

Notes:
- `html/template` applies escaping on first execution (per the `Tree` field comment), and `Funcs` must be registered before parsing.
- Embedded templates (`embed.FS`) change only on rebuild. templ's docs advise serving static assets from disk rather than embedding in dev; the same applies to template files.

## 6. Rendering performance: templ vs `html/template`

Sources: [templ benchmarks README](https://github.com/a-h/templ/blob/main/benchmarks/templ/README.md), [render_test.go](https://github.com/a-h/templ/blob/main/benchmarks/templ/render_test.go), [html/template docs](https://pkg.go.dev/html/template).

- templ's own benchmark (2023-08-17, darwin/arm64, one small `Person` template rendered to a `strings.Builder`):

  | Benchmark | ns/op | B/op | allocs/op |
  |---|---|---|---|
  | templ | 369.1 | 536 | 6 |
  | Go template (`html/template`) | 2,475 | 1,400 | 38 |
  | `io.WriteString` baseline | 56.64 | 320 | 1 |

- Caveats: maintainers' benchmark, dated 2023, tiny template. The suite has separate benchmarks (`BenchmarkTemplRender`, `BenchmarkGoTemplateRender`, `BenchmarkIOWriteString`), not one head-to-head run. No independent benchmark was found among primary sources, and nothing was measured here.
- `html/template` is parsed once and "may be executed safely in parallel". templ compiles to Go code that writes to an `io.Writer`.
- Real request cost is usually dominated by I/O (Postgres, Redis); measure on the real app before treating this as a deciding factor.

## 7. Option summary

| Option | Reloads | Needs Node | Extra binaries | Notes |
|---|---|---|---|---|
| templ `generate --watch --proxy --cmd` | templ files, Go files, browser | No | `templ` (go tool) | templ projects only; text-only edits skip restart |
| air + templ `--notify-proxy` | all, plus other builds | No | `air` (go tool), `templ` | Documented multi-process Makefile pattern |
| air alone (own proxy) | Go and watched static files, browser | No | `air` (go tool) | Works with `html/template`; set `include_ext` for templates and CSS |
| wgo | restarts only | No | `wgo` | Needs a separate browser-reload mechanism |
| Tailwind standalone CLI `--watch` | rebuilds the CSS file only | No | `tailwindcss` binary (not a Go tool) | Needs a reload trigger and a download step |
| Vite + Go manifest helper | HMR for JS/CSS, hashed builds | Yes (20.19+ / 22.12+) | none beyond Node | HTML edits still need a reload hook |

## Open points for the decision ticket

- Node acceptable or not (drives Vite vs standalone Tailwind and who owns asset hashing).
- The templating choice (#12, #14 depend on this) changes the watcher chain.
- If templ is chosen, check templ source for how `TEMPL_DEV_MODE` is enabled and whether `--open-browser` / `--watch-pattern` exist.
- Go 1.27 specifics were not researched; the facts above come from docs stating Go 1.24 (tool directives) and Go 1.25 (air install) minimums.
