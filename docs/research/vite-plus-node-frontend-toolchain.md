# Vite+ and a Node frontend toolchain for Go SSR

Research for [#25](https://github.com/DannyFestor/go-template-html/issues/25), checked against primary sources on 2026-10-10. It asks what a Node-based frontend toolchain would cost and give the templ + htmx 4 + `@alpinejs/csp` app that [#13](https://github.com/DannyFestor/go-template-html/issues/13), [#14](https://github.com/DannyFestor/go-template-html/issues/14) and [#21](https://github.com/DannyFestor/go-template-html/issues/21) settled without Node.

Claims marked **(unverified)** were not confirmed against a primary source or by running code.

## Summary

- Vite+ 1.0 went stable on 2026-09-28, and 1.1.0 (2026-10-07) is the latest release. It is MIT-licensed and free, with no paid tier. It bundles Vite 8, Rolldown, Vitest 5, Oxlint, Oxfmt, tsdown and Vite Task behind one `vp` CLI, and it also manages Node.js and the package manager.
- Vite's backend integration fits a Go server: `build.manifest` writes a JSON map from each source entry to hashed files, and Go reads it. No official Go or templ integration exists. The best-known Go library (`olivere/vite`) has been inactive since June 2025, and by default it emits an inline React preamble in dev mode, which our CSP blocks.
- The production build is CSP-compatible: `<script type="module" src>`, `<link rel="stylesheet">` and `<link rel="modulepreload">` tags, and none of them inline. The ADR-0004 / #14 claim that ES modules need an inline script holds only for import maps, which a bundler makes unnecessary.
- Dev mode works with `script-src 'self'` only if the Vite server is same-origin. Otherwise the dev CSP has to allow the Vite origin and its websocket. Vite's client uses no `eval`. JS-imported CSS and the error overlay inject `<style>` elements, so they would need `style-src` allowances. CSS linked with `<link>` hot-swaps without a reload. An Alpine component edit triggers a full page reload unless the module accepts HMR.
- `@tailwindcss/vite` scans `.templ` files: it uses Vite's root as the base and scans every non-ignored text file there, then registers the scanned files as watch dependencies.
- Oxlint (1.x) and Oxfmt (0.x, beta) handle `static/js/*.js`. Both ship as standalone binaries that need no Node, so they could be adopted without Vite. Neither supports `.templ`, which `templ fmt` keeps covering.
- Cost: Node and npm dependencies with a lockfile, a Node build stage in the Dockerfile, a `package.json` next to `go.mod`, a manifest-reading resolver in `internal/asset`, and the `go:embed` dot/underscore rule. Gain: bundling and minifying JS (fewer `<script>` tags), npm-managed versions instead of the vendored-JS manifest, CSS HMR, and Vitest for component tests.

## 1. Vite+ status, licence, pricing, contents, install and pinning

### Status and releases

- `v1.0.0` "Vite+ 1.0 is stable", published 2026-09-28. It promotes `v1.0.0-rc.1` with no code changes. ([release](https://github.com/voidzero-dev/vite-plus/releases/tag/v1.0.0))
- `v1.1.0`, published 2026-10-07, is marked Latest. ([releases](https://github.com/voidzero-dev/vite-plus/releases))
- The 1.0 announcement says "Vite+ reached 1.0 but is far from feature-complete", and lists remote caching, `vp release` and `vp docs` as planned. ([announcement](https://voidzero.dev/posts/announcing-vite-plus-1-0))
- Starting with `v1.0.0-rc.0`, the CLI requires Node.js `^22.18.0 || ^24.11.0 || >=26.0.0`. ([v1.0.0 notes](https://github.com/voidzero-dev/vite-plus/releases/tag/v1.0.0))

### Licence and pricing

- MIT. The repository licence is MIT ([repo](https://github.com/voidzero-dev/vite-plus)), and the site says "Free and open source under the MIT license" ([viteplus.dev](https://viteplus.dev/)).
- VoidZero joined Cloudflare on 2026-06-04: "Vite, Vitest, Rolldown, Oxc, and Vite+ will remain open-source and MIT-licensed." The same post says "We experimented with a mixed licensing model for Vite+, but it didn't feel right." The company's revenue model is now a service (Void), not Vite+ licensing. ([VoidZero post](https://voidzero.dev/posts/voidzero-cloudflare), [Cloudflare press release](https://www.cloudflare.com/press/press-releases/2026/cloudflare-acquires-voidzero-to-build-the-future-of-the-ai-native-web/))
- No paid tier appears on viteplus.dev or in the 1.0 announcement.
- Telemetry: the docs don't mention any **(unverified either way)**.

### What it bundles

Bundled versions per release, from the release notes:

| Tool | v1.0.0 | v1.1.0 |
| --- | --- | --- |
| vite | 8.3.1 | 8.3.3 |
| rolldown | 1.2.11 | 1.2.12 |
| vitest | 5.0.1 | 5.0.3 |
| oxlint | 1.85.0 | 1.87.0 |
| oxfmt | 0.70.0 | 0.72.0 |
| tsdown | 0.23.0 | not listed in the 1.1.0 notes |
| oxlint-tsgolint | 7.0.2003 | not listed in the 1.1.0 notes |

Commands: `vp dev`/`build`/`preview` (Vite and Rolldown), `vp check`/`lint`/`fmt` (Oxc), `vp test` (Vitest), `vp pack` (tsdown), `vp run` (Vite Task with caching), `vp staged`, package-manager commands, `vp env` (Node.js and package-manager versions), and `vp create`/`migrate`. ([local CLI guide](https://github.com/voidzero-dev/vite-plus/blob/main/docs/guide/local-cli.md))

The v1.1.0 notes say the newer lint and format versions "can flag code that passed before". Bumping Vite+ therefore bumps the linter and formatter with it.

### Install and pinning

There are two parts:

- **Global `vp` CLI**: `curl -fsSL https://vite.plus | bash`, pinned with `VP_VERSION=1.2.3` ([global CLI guide](https://github.com/voidzero-dev/vite-plus/blob/main/docs/guide/global-cli.md)). Each release publishes per-platform `vp-*.tar.gz` assets (macOS arm64/x64, Linux arm64/x64 glibc and musl, Windows) plus a `vp-checksums.txt`, so the curl-and-verify pattern from #21 (as used for `bin/tailwindcss`) could fetch it. Since `rc.0`, "standalone installs and `vp upgrade` verify release provenance". The global CLI downloads Node.js itself, resolving the version from `.node-version`, `devEngines.runtime` or `engines.node`.
- **Project-local `vite-plus` npm package**: a dev dependency recorded in `package.json` and the lockfile. Manual installs also need package-manager overrides that alias `vite` to `@voidzero-dev/vite-plus-core@<version>` and pin `vitest` to the bundled version. Otherwise, in the guide's words, dependency bots "can update these packages independently and leave incompatible versions installed together" ([local CLI guide](https://github.com/voidzero-dev/vite-plus/blob/main/docs/guide/local-cli.md)). The global CLI delegates `vp dev`/`build`/`test` to the project-local version when one is installed.
- **Docker image**: `ghcr.io/voidzero-dev/vite-plus:<major>.<minor>.<patch>` for `linux/amd64` and `linux/arm64`. It is meant for builds and CI, "not intended as a production runtime image", and it runs as a non-root `vp` user with passwordless `sudo` ([Docker guide](https://github.com/voidzero-dev/vite-plus/blob/main/docs/guide/docker.md)).

### Plain Vite instead of Vite+

The same tools can be installed separately as `vite`, `@tailwindcss/vite`, `vitest`, `oxlint` and `oxfmt` npm packages, without the `vp` layer, the overrides, or Node and package-manager management. Vite+ adds one version for the whole toolchain plus `vp check`, `vp run` caching and `vp staged`. The template already covers several of these with Make, lefthook and Go tooling (ADR-0010).

## 2. Vite backend integration for a Go server

### What Vite documents

From the [backend integration guide](https://vite.dev/guide/backend-integration):

- Config: set `build.manifest: true` (written to `<outDir>/.vite/manifest.json`), set the entry through `build.rolldownOptions.input` in place of `index.html`, and set `server.cors.origin` to the backend origin. With a non-HTML entry, add `import 'vite/modulepreload-polyfill'` at the top of the entry unless the polyfill is disabled.
- Dev tags in the server template: `<script type="module" src="http://localhost:5173/@vite/client">` and `<script type="module" src="http://localhost:5173/main.js">`. Static assets need either a proxy to Vite or `server.origin`.
- Manifest: `Record<name, ManifestChunk>`, where each chunk has `file`, `src`, `css`, `assets`, `isEntry`, `name`, `isDynamicEntry`, `imports` and `dynamicImports`. Entry keys are source paths relative to the project root.
- Production tags, in order: a stylesheet link for each `css` entry, the CSS of each recursively imported chunk, then `<script type="module" src>` for the entry, and optionally `<link rel="modulepreload">` for each imported chunk.
- `build.manifest` may also be a string path relative to `build.outDir` ([build options](https://vite.dev/config/build-options)).

Latest Vite is `v8.3.4` (2026-10-08) ([releases](https://github.com/vitejs/vite/releases)).

### Fit with `internal/asset`

The #14 design has `asset.URL(...)` behind a resolver interface, with startup hashing in production and plain paths in dev. With Vite:

- The production resolver would look up the source path in the manifest and return `file`, plus the CSS and preload lists for entries. Vite computes the hashes, so the Go startup hashing would no longer apply to Vite outputs. Vendored files that stay outside Vite could keep it.
- **`go:embed` excludes names beginning with `.` or `_`** unless the pattern uses `all:` ([embed docs](https://pkg.go.dev/embed)). The default `.vite/manifest.json` would be silently skipped by `//go:embed static`. Either set `build.manifest` to a plain path such as `manifest.json`, or embed with `all:static`. The `all:` prefix also covers any emitted chunk whose name starts with `_` (whether Rolldown emits such names for this app's graph is **unverified**).
- #14's stale-hash fallback (an old hashed URL gets the current file with `no-cache`) relies on the Go side knowing the hash format. With Vite-named files, it would need a rule that maps an unknown hashed name to the current file; that design is open.
- Hashed files go under `build.assetsDir` (default `assets`), which is a different layout from `static/css` and `static/js`.

### Existing Go and templ integrations

- No official templ integration. A search of the [templ repository](https://github.com/a-h/templ) docs finds no mention of Vite. templ is at `v0.3.1070` (2026-10-04).
- [`olivere/vite`](https://github.com/olivere/vite) (MIT, 217 stars, last push 2025-06-16, one tag `v0.1.0`) has a `Fragment` helper and an `http.Handler`, and reads `.vite/manifest.json` in production. In dev, when no `ViteTemplate` is set, **it emits the React refresh preamble inline by default** (`fragment.go`: "If the Vite template value is less than 1 … the default React preamble is applied"). It would need explicit configuration to avoid the inline script, and it targets `html/template`, not templ.
- [`torenware/vite-go`](https://github.com/torenware/vite-go) targets Vite 2 and 3, and was last pushed in 2023-08.
- The manifest format is small and documented, and the template already has `internal/asset`, so a Go reader there is likely less code than taking on a dependency (judgement, not measured).

### Dev server versus the templ `--proxy` loop

- templ's proxy injects `<script src="/_templ/reload/script.js">`, an external same-origin script that carries a nonce when the page CSP has one, and reloads on `.templ` and `.go` changes ([`proxy.go`](https://github.com/a-h/templ/blob/main/cmd/templ/generatecmd/proxy/proxy.go)).
- Vite's client (`/@vite/client`) opens its own websocket. From [`client.ts`](https://github.com/vitejs/vite/blob/main/packages/vite/src/client/client.ts):
  - CSS referenced through `<link>` gets a `css-update`: the client clones the link with a `?t=` query and removes the old one after load. **No page reload.**
  - CSS imported from JS is injected as a `<style data-vite-dev-id>` element.
  - A `full-reload` payload calls `location.reload()`.
- Server side ([`hmr.ts`](https://github.com/vitejs/vite/blob/main/packages/vite/src/node/server/hmr.ts)): a change with no accepting boundary (a "dead end") sends `full-reload`. Alpine component files that register `Alpine.data` and never call `import.meta.hot.accept` therefore reload the full page on edit. Re-registering a component in place after `Alpine.start()` is **unverified**.
- That makes two reload channels: templ's proxy for `.templ` and `.go`, and Vite for CSS and JS. Tailwind in Vite also rebuilds the CSS when a `.templ` file changes (section 3), so a `.templ` edit gives a templ reload plus a Vite CSS swap (observed behaviour **unverified**).
- Alternative without HMR: `vite build --watch` writes to disk and the #14 air plus `templ generate --notify-proxy` loop stays. The page then never loads anything from the Vite server, so no dev-only CSP change is needed, but CSS also does a full reload.

## 3. Tailwind v4 Vite plugin and `.templ`

- Tailwind's automatic detection scans "every file in your project" except `.gitignore`d files, `node_modules`, binary files, CSS files and package-manager lock files. `@source`, `@source not` and `source(none)` adjust it ([docs](https://tailwindcss.com/docs/detecting-classes-in-source-files)).
- In `@tailwindcss/vite`, the base is Vite's `config.root`. With no `source(...)`, the scanner uses `{ base, pattern: '**/*' }`, and every scanned file and `@source` glob is registered with `addWatchFile`, so edits trigger CSS regeneration ([`packages/@tailwindcss-vite/src/index.ts`](https://github.com/tailwindlabs/tailwindcss/blob/main/packages/@tailwindcss-vite/src/index.ts)).
- So `.templ` files under the Vite root are scanned, the same as with the standalone CLI chosen in #14. If `vite.config` doesn't live at the repo root, use `@import "tailwindcss" source("../..")` or `@source` to reach the views.
- `@tailwindcss/vite` 4.3.3 has the peer range `vite ^5.2.0 || ^6 || ^7 || ^8` (npm registry). Tailwind's latest release is `v4.3.3` (2026-07-16).

## 4. Oxlint, Oxfmt and `.templ`

### Oxlint

- Lints `.js`, `.mjs`, `.cjs`, `.ts`, `.mts`, `.cts`, `.jsx` and `.tsx`, plus the `<script>` blocks of `.vue`, `.svelte` and `.astro` files ([linter guide](https://oxc.rs/docs/guide/usage/linter.html)). `static/js/*.js` classic scripts are plain `.js` and in scope.
- Enables the `correctness` category by default. `env` (for example `browser`) and `globals` (`readonly`/`writable`/`off`) declare globals ([config](https://oxc.rs/docs/guide/usage/linter/config.html)), so a global `Alpine` can be declared readonly.
- Version 1.x; latest `oxlint_v1.87.0` (2026-10-05) ([releases](https://github.com/oxc-project/oxc/releases)).

### Oxfmt

- Formats JS, TS, JSON, CSS/SCSS/Less, GraphQL, TOML, YAML and Markdown natively in Rust. HTML, Vue, Svelte, MDX, Handlebars and Angular go through a bundled Prettier, which "require[s] Node.js and [is] only available with the `oxfmt` npm package" ([language support](https://oxc.rs/docs/guide/usage/formatter/language-support.html)).
- Aims to match Prettier's output; it has built-in Tailwind class sorting and import sorting ([formatter guide](https://oxc.rs/docs/guide/usage/formatter.html)).
- Status: beta since 2026-02-24 ([Oxfmt Beta](https://oxc.rs/blog/2026-02-24-oxfmt-beta)). Latest `oxfmt_v0.72.0` (2026-10-05). No stable 1.0 announcement was found.

### Standalone binaries, no Node

Every oxlint and oxfmt release on GitHub ships per-platform binaries: macOS arm64/x64, Linux arm64/x64 glibc and musl, Windows, and more ([oxlint 1.87.0](https://github.com/oxc-project/oxc/releases/tag/oxlint_v1.87.0), [oxfmt 0.72.0](https://github.com/oxc-project/oxc/releases/tag/oxfmt_v0.72.0)). Linting and formatting the vendored-JS setup from #14 therefore needs no Node or Vite: the binaries fit the #21 `bin/tailwindcss` curl pattern. Difference: these releases publish no checksum file (only the archives are listed), so the pin would carry sha256 values computed when the version is bumped.

### `.templ`

- Oxlint and Oxfmt don't list `.templ`.
- `templ fmt` formats `.templ` files, and `templ fmt -fail .` fails CI on unformatted files. "If `prettierd`, `prettier` or `npx` is found in your `PATH`, `templ fmt` will use prettier to format `script` and `style` elements" ([templ CLI docs](https://github.com/a-h/templ/blob/main/docs/docs/09-developer-tools/01-cli.md)). Adding Node to the dev machine can therefore change `templ fmt` output for those elements, though the template avoids inline scripts anyway.
- No Prettier plugin for templ was found on GitHub (searched `prettier-plugin-templ`).

## 5. Strict CSP (`script-src 'self'`, no inline, no `unsafe-eval`)

### Production build

- The emitted tags are `<link rel="stylesheet">`, `<script type="module" src>` and optionally `<link rel="modulepreload">`, all external. `script-src 'self'` covers them when served from the app origin.
- The modulepreload polyfill is "auto injected into the proxy module of each `index.html` entry". With a non-HTML entry it is an `import` inside the entry, so it ships in a bundled file, not inline ([build options](https://vite.dev/config/build-options)). `build.modulePreload: { polyfill: false }` drops it.
- Import maps are the inline case: the experimental `build.chunkImportMap` emits an `importmap.json` to inject as `<script type="importmap">` ([backend integration](https://vite.dev/guide/backend-integration)). Leave it off.
- Vite inlines assets under 4 KiB as `data:` URIs by default. Vite's CSP section says to allow `data:` in `img-src`/`font-src` or set `build.assetsInlineLimit: 0`, and warns never to allow `data:` in `script-src` ([features](https://vite.dev/guide/features)).
- `html.cspNonce` adds nonces to `<script>`, `<style>` and `<link>` tags in Vite-processed HTML and emits a `<meta property="csp-nonce">`. That applies to Vite-rendered HTML; with templ rendering the page it applies only if the app renders that meta tag itself (behaviour with a Go-rendered page **unverified**).

### Dev mode

- **Different origin**: Vite serves from `http://localhost:5173`, while the page comes from the Go server or the templ proxy. `script-src 'self'` blocks the dev scripts unless the dev CSP adds the Vite origin, and the HMR websocket needs `connect-src` for it **(the exact directive set is unverified)**. Proxying Vite through the Go dev server keeps things same-origin but needs `server.hmr` settings (**unverified**).
- **No eval**: `client.ts` and `overlay.ts` contain no `eval` or `new Function`.
- **Injected `<style>`**: CSS imported from JS and the error overlay create `<style>` elements, nonced from the `csp-nonce` meta tag when one is present. A strict `style-src` would block them without that nonce. Linking the Tailwind CSS with `<link>` avoids the JS-injected style path for app CSS.
- The React preamble is the only inline script the integration guide mentions, and it applies only with `@vitejs/plugin-react`. Our stack doesn't use that plugin; `olivere/vite` emits the preamble by default (section 2).

### Alpine and htmx as modules

- `@alpinejs/csp` 3.17.4 and `@alpinejs/persist` 3.17.4 publish `dist/module.esm.js`, and `htmx.org@4.0.0` (dist-tag `next`; `latest` is still 2.0.11) publishes `dist/htmx.esm.js` (npm registry). A Vite entry could import them and call `Alpine.start()`. The CSP build's expression limits from ADR-0004 stay the same, because they come from Alpine and not from the bundler.
- Whether the htmx 4 ESM build also sets `window.htmx`, and how the load order from #14 (htmx, components, persist, alpine-csp) carries over into a single bundle, is **unverified**.

## 6. Effect on the Dockerfile and Make targets

### Dockerfile

Today (#21) the build stage runs on the Debian `golang` image, runs `make build` (including `bin/tailwindcss`), and copies `bin/web`, `bin/worker` and `bin/migrate` into the final image.

With Vite, the built assets must exist before `go build`, because `go:embed` reads them at compile time:

1. Add a Node build stage, either `ghcr.io/voidzero-dev/vite-plus:<exact>` or an official `node` image. It copies `package.json`, the lockfile and `.node-version`, installs with a frozen lockfile, copies the frontend sources plus the `.templ` files Tailwind scans, then runs `vp build` / `vite build`. The Vite+ image runs as user `vp`, so the guide uses `COPY --chown=vp:vp`.
2. The Go build stage copies the built `internal/asset/static` output from that stage, then builds as before. The `make build` target can no longer build the assets itself inside the Go stage unless Node is installed there too.
3. The runtime image is unchanged: it holds the Go binaries only, with no Node. The Vite+ guide's "static SPA" pattern (Node only at build time) is the closest match.

Pinning grows by the Node version (`.node-version`), `package.json` plus the lockfile, and the Vite+ image tag or digest. Tailwind moves from `tools/tailwind.version` plus `tools/tailwind.sha256` to an npm dependency.

### Make targets (ADR-0010: one command per leaf target)

The likely changes, assuming ADR-0010 stays as written:

- `make dev`: `vite` / `vp dev` (or `vite build --watch`) replaces the Tailwind `--watch` watcher. Whether air keeps watching `static/` depends on the HMR-versus-build-watch choice above.
- `make build`: depends on `node_modules` (install from the lockfile) and the Vite build, in place of `bin/tailwindcss`.
- `make lint` / `make fmt`: add oxlint and oxfmt, either through `vp lint`/`vp fmt` or as standalone binaries.
- `make test`: Vitest for Alpine component tests is new. The build tags from ADR-0011 cover Go only.
- The vendored-JS update task and the `static/vendor/` manifest from #14/#21 go away, replaced by npm versions in the lockfile.
- `docs/upgrading.md` gains sections for Node, Vite+ (or each npm tool), and the Vite+ `vite`/`vitest` override alignment.
- A fresh clone needs Node, or the global `vp` that provisions it, before `make dev` and `make build` work. That reverses #21's rejection of mise for "an extra install before the repo works".

## Cost and gain

**Gains**
- JS is bundled and minified into one or a few hashed files, in place of a `<script defer>` per component.
- npm and a lockfile manage the versions of htmx, Alpine, Tailwind and the linters, replacing the vendored-JS download-and-verify manifest.
- CSS hot-swaps without a page reload when linked with `<link>`.
- Vitest is available for Alpine component logic, and oxlint/oxfmt for JS (these two also exist without Node).
- One toolchain version through Vite+.

**Costs**
- Node, a package manager and `node_modules` on every dev machine, in CI and in the Docker build stage. Vite+ adds its own CLI layer and dependency overrides.
- A Go manifest reader in `internal/asset`, the `go:embed` `.`/`_` rule, and a redesign of the stale-hash fallback.
- Two reload channels in dev, and a dev-only CSP relaxation unless Vite is proxied same-origin.
- No templ-aware Go integration exists; the most visible one is inactive and emits an inline preamble by default.
- Oxfmt is still 0.x, and Vite+ bumps can change lint and format results.
- ADR-0004 states "no Node or JS build step", so adopting this needs a superseding ADR, and #14 and #21 need revising.

## Sources

- Vite+: [site](https://viteplus.dev/), [repo](https://github.com/voidzero-dev/vite-plus), [v1.0.0](https://github.com/voidzero-dev/vite-plus/releases/tag/v1.0.0), [v1.1.0](https://github.com/voidzero-dev/vite-plus/releases/tag/v1.1.0), [1.0 announcement](https://voidzero.dev/posts/announcing-vite-plus-1-0), [VoidZero joins Cloudflare](https://voidzero.dev/posts/voidzero-cloudflare), [local CLI](https://github.com/voidzero-dev/vite-plus/blob/main/docs/guide/local-cli.md), [global CLI](https://github.com/voidzero-dev/vite-plus/blob/main/docs/guide/global-cli.md), [Docker](https://github.com/voidzero-dev/vite-plus/blob/main/docs/guide/docker.md)
- Vite: [backend integration](https://vite.dev/guide/backend-integration), [features (CSP)](https://vite.dev/guide/features), [build options](https://vite.dev/config/build-options), [HMR API](https://vite.dev/guide/api-hmr), [`client.ts`](https://github.com/vitejs/vite/blob/main/packages/vite/src/client/client.ts), [`hmr.ts`](https://github.com/vitejs/vite/blob/main/packages/vite/src/node/server/hmr.ts)
- Tailwind: [detecting classes](https://tailwindcss.com/docs/detecting-classes-in-source-files), [`@tailwindcss/vite` source](https://github.com/tailwindlabs/tailwindcss/blob/main/packages/@tailwindcss-vite/src/index.ts)
- Oxc: [linter](https://oxc.rs/docs/guide/usage/linter.html), [linter config](https://oxc.rs/docs/guide/usage/linter/config.html), [formatter](https://oxc.rs/docs/guide/usage/formatter.html), [formatter language support](https://oxc.rs/docs/guide/usage/formatter/language-support.html), [Oxfmt Beta](https://oxc.rs/blog/2026-02-24-oxfmt-beta), [releases](https://github.com/oxc-project/oxc/releases)
- templ: [CLI docs](https://github.com/a-h/templ/blob/main/docs/docs/09-developer-tools/01-cli.md), [proxy source](https://github.com/a-h/templ/blob/main/cmd/templ/generatecmd/proxy/proxy.go)
- Go: [`embed`](https://pkg.go.dev/embed)
- Go integrations: [`olivere/vite`](https://github.com/olivere/vite), [`torenware/vite-go`](https://github.com/torenware/vite-go)
- npm registry metadata: `@alpinejs/csp`, `@alpinejs/persist`, `htmx.org`, `@tailwindcss/vite`
