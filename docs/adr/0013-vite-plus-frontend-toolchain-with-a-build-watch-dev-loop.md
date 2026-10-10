# Vite+ frontend toolchain with a build-watch dev loop

Frontend assets are built, linted, formatted, type-checked and tested by Vite+, and Node is a hard dependency of the template. The alternative was Tailwind's standalone CLI plus hand-vendored JS, which needs no Node. We chose Vite+ anyway: it pins Node and every frontend package through one lockfile, bundles htmx and Alpine from npm rather than committing minified copies, and brings strict TypeScript, oxlint, oxfmt and Vitest, so the TypeScript gets the same linter-enforced discipline as the Go code. `bin/vp` is downloaded and checksum-verified by Make, like every other pinned binary, and it provisions the Node version in `.node-version`. A fresh clone still needs only Go, Docker and git. One `app.ts` entry imports htmx, `@alpinejs/csp`, persist and every component, registers the components and starts Alpine. The build emits one module script and one stylesheet (Tailwind through `@tailwindcss/vite`), plus a manifest that `internal/asset` reads to resolve hashed URLs. In development `vp build --watch` writes the same output to disk, and the existing air and templ proxy loop reloads the page. There is no Vite dev server, so dev runs the production CSP and the same manifest code path.

## Considered Options

- **No Node** (Tailwind standalone CLI, vendored JS, standalone oxlint and oxfmt binaries): fewest moving parts, but versions are pinned by hand-written checksum manifests, there is no TypeScript or JS unit testing, and components can't `import` each other.
- **Node only for linting and testing**: Node becomes a prerequisite anyway, but the asset pipeline still pins and vendors by hand.
- **Vite dev server with HMR**: CSS swaps without a reload, but it needs a dev-only CSP that allows the Vite origin, its websocket and injected `<style>` elements (or a same-origin proxy). It also adds a second reload channel next to templ's and a dev-only resolver. Alpine component edits reload the full page either way.
- **Plain Vite with separate oxlint, oxfmt and Vitest packages**: one less layer, but nothing keeps their versions in step. Vite+ releases them together.
- **npm**: matches the Laravel starter kit. pnpm was picked instead, because it blocks dependency lifecycle scripts by default and `minimumReleaseAge` holds back freshly published versions.
- **The official `vite-plus` Docker image as a build stage**: it runs with passwordless `sudo`, and a separate stage would pin Node a second time.

## Consequences

- `package.json`, `pnpm-lock.yaml`, `.node-version`, `tsconfig.json` and `vite.config.ts` sit at the repo root. The Vite root is the repo root, so Tailwind's automatic detection reaches every `.templ` file.
- Sources live in `internal/asset/src/`. The output goes to `internal/asset/dist/`, which is gitignored apart from a committed `.gitkeep`, and it is embedded with `//go:embed all:dist` so a fresh clone compiles. The manifest is written to `dist/manifest.json` because `go:embed` skips Vite's default `.vite/` directory. Production fails at startup if the manifest is missing. Dev re-reads it on every lookup, because each rebuild changes the hashes.
- Vite computes the hashes. An unknown hash for a known `[name]` still serves the current file with `no-cache`, so HTML from before a deploy keeps its CSS and JS.
- The strict CSP holds: the modulepreload polyfill and import maps are off, and `assetsInlineLimit` is 0, so no `data:` URIs are emitted.
- Installs always use the frozen lockfile. `vite-plus`, its `vite` and `vitest` overrides and `bin/vp` are bumped together by `make vp-upgrade`. Node is bumped by hand in `.node-version`.
- `templ fmt` uses prettier for `<script>` and `<style>` elements whenever Node tooling is on `PATH`. The CSP rules those elements out of templates, so the output does not change.
