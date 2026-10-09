# Client-side interactivity options for Go SSR

Research for issue #8 (map #1). Facts and trade-offs only; the decision belongs to a later grilling ticket. Gathered 2026-10-09.

Standing constraints: Go 1.27 SSR app, stdlib-first, Tailwind CSS from day one, reference experience is Laravel Livewire. templ vs `html/template` is undecided; Node acceptability is undecided.

## Summary table

| Option | Livewire-likeness | JS build step | `html/template` | templ | Tailwind | Maturity (2026-10-09) |
|---|---|---|---|---|---|---|
| No JS (forms + full reloads) | Low | None | Native | Native | Native | n/a |
| HTMX + Alpine.js | Medium-high | None (script tags) | Native | Native | Native | htmx 4.0.0 (2026-08-28), ~49.5k stars; Alpine 3.17.4, ~31.9k stars |
| Datastar | High (server-driven, SSE) | None (script tag) | Works (any HTML) | Works (any HTML) | Native | v1.0.4 (2026-09-21), ~5.1k stars, official Go SDK |
| Unpoly | Medium | None (CDN) / npm optional | Native | Native | Native | 3.14.3, ~2.8k stars, MIT |
| Turbo (+ Stimulus) | Medium | None (script tag) / bundler for npm | Native | Native | Native | 8.0.23, ~7.4k stars; Stimulus ~13.1k stars |
| React/Vue/Svelte SPA | Low (client-driven, JSON API) | Node required | Only as a shell | Only as a shell | Via Node build | Most mature, least aligned with SSR |
| Islands (Astro/Vite per component) | Low-medium | Node required | Shell + mount points | Shell + mount points | Via Node build | Mature tooling, manual Go integration |

Likeness ratings are the researcher's judgement, not a sourced claim.

## No JavaScript

Plain HTML forms, POST/redirect/GET, full-page reloads. No build step; works with both templating options and Tailwind unchanged. Gaps against Livewire: no live validation, no partial updates, and only the interactivity native elements (`<dialog>`, `<details>`) and CSS give. This follows from the platform rather than a cited source.

## HTMX + Alpine.js

- **What it is.** htmx swaps server-returned HTML fragments into the page via HTML attributes. In htmx 4 all requests use native `fetch()` (XMLHttpRequest removed), SSE ships as a core extension `hx-sse`, attribute inheritance is explicit (`:inherited`), and 4xx/5xx responses now swap by default (only 204/304 do not). Source: [htmx 4 migration notes](https://four.htmx.org/htmx-4).
- **Version and activity.** GitHub release `v4.0.0` is non-prerelease, published 2026-08-28; repo pushed 2026-10-02, about 49.5k stars. Source: [bigskysoftware/htmx releases](https://github.com/bigskysoftware/htmx/releases). The page fetched does not state a support policy or end of life for 2.x; docs exist for v1, v2 and v4.
- **Alpine.js.** Handles client-only state (dropdowns, modals, toggles). Latest release v3.17.4 (2026-09-21), about 31.9k stars. Install is a `<script defer>` tag; the npm route needs a bundler. Sources: [Alpine installation](https://alpinejs.dev/essentials/installation), [releases](https://github.com/alpinejs/alpine/releases).
- **Livewire-likeness.** Request/response server rendering of fragments, so the server owns state, but there is no component abstraction and no automatic state hydration; the developer wires endpoints and fragment templates. Livewire components pair a PHP class with a Blade template and bind public properties (`wire:model`, `wire:submit`), see [Livewire docs](https://livewire.laravel.com/docs/understanding-livewire).
- **Build step.** None: a `<script>` tag or vendored file. Source: [templ htmx guide](https://templ.guide/server-side-rendering/htmx) (download `htmx.min.js`, add script tag).
- **templ.** Documented by templ (guide above, including `templ.JSFuncCall` for passing server data to `hx-on:`). Fragment endpoints map naturally to templ components.
- **`html/template`.** Works; fragments are `{{define}}` blocks executed by name via `ExecuteTemplate`. Not a documented htmx feature; follows from both being plain HTML.
- **Tailwind.** Compatible; classes live in HTML either way. Tailwind scans templates for class names ([Tailwind CLI docs](https://tailwindcss.com/docs/installation/tailwind-cli)); `.templ` or `.html` files must be in its scan paths.
- **Caveat.** htmx 4 has breaking changes versus 2.x, so tutorials and third-party examples written for 1.x/2.x may not apply.

## Datastar

- **What it is.** Hypermedia framework combining backend-driven DOM patching with small client signals. A `text/html` response is morphed into the DOM by element ID; a `text/event-stream` response carries `datastar-patch-elements` events over a long-lived connection. Signals use a `$` prefix (`data-bind`, `data-text`, `data-computed`, `data-signals`); the backend can patch signals via JSON Merge Patch (RFC 7396) or `datastar-patch-signals` SSE events. Sources: [getting started](https://data-star.dev/guide/getting_started), [reactive signals](https://data-star.dev/guide/reactive_signals).
- **Version and activity.** v1.0.4 (2026-09-21), about 5.1k stars, MIT. Source: [starfederation/datastar](https://github.com/starfederation/datastar).
- **Go SDK.** Official `github.com/starfederation/datastar-go`, MIT, requires Go 1.24+, last pushed 2026-06-02, about 208 stars. Features: read client signals into a struct, open an SSE writer (`datastar.NewSSE(w, r)`), patch/remove elements, patch signals, execute JS, redirect. Source: [datastar-go README](https://github.com/starfederation/datastar-go).
- **Livewire-likeness.** Closest of the hypermedia group on one axis: the server drives UI state and can push updates over SSE, with client signals similar in spirit to Livewire's bound properties (researcher's comparison).
- **Build step.** None: `<script type="module">` from a CDN or self-hosted; "doesn't require any npm packages or other dependencies". Source: getting started page above.
- **templ / `html/template`.** The Datastar docs fetched do not mention templ; the SDK takes HTML strings, so either engine works. templ's guide lists a Datastar page in its navigation; its content was not retrieved.
- **Tailwind.** No coupling; classes in HTML.
- **Licensing note.** Core is open source. Datastar Pro (extra attributes, actions, bundler, Inspector, Stellar CSS alpha) is a paid one-time license ($349 solo, $1,299 team up to 25 employees) and cannot be placed in public repos or open-source projects, which matters for a public template repo. Source: [Datastar Pro](https://data-star.dev/reference/datastar_pro).
- **Trade-off.** SSE holds a long-lived connection per open page; relevant to the Redis/websocket plans in the map. Not measured here.

## Unpoly

- Progressive-enhancement library: links and forms update page fragments, with layers (modals, drawers). Works with "any server that can render HTML"; optional server bindings read/write HTTP headers (no Go binding found in the docs fetched). Install via CDN `<link>` and `<script defer>`, or `npm install unpoly`; "doesn't require a build tool", no dependencies. Source: [Unpoly install](https://unpoly.com/install).
- Latest GitHub tag v3.14.3, about 2.8k stars, MIT, pushed 2026-10-08. Source: [unpoly/unpoly](https://github.com/unpoly/unpoly).
- Livewire-likeness: medium; strong on layered UI patterns and form validation flows, no signals or server push (researcher's judgement). Pairs with both templating options; Tailwind neutral.
- Go-ecosystem adoption: not measured beyond star counts.

## Turbo (Hotwire)

- Components: Drive (intercepts links/forms, swaps `<body>`), Frames (`<turbo-frame>` scoped navigation, lazy loading), Streams (`<turbo-stream>` actions such as append/replace/remove, delivered over WebSocket, SSE or form responses), Native (iOS/Android). Source: [Turbo handbook](https://turbo.hotwired.dev/handbook/introduction).
- Install: script tag (no bundler) or npm (bundler such as esbuild). Source: [installing Turbo](https://turbo.hotwired.dev/handbook/installing).
- v8.0.23 (2026-01-29), repo pushed 2026-09-29, about 7.4k stars. Stimulus (for client-only behavior) about 13.1k stars. Sources: [hotwired/turbo](https://github.com/hotwired/turbo), [hotwired/stimulus](https://github.com/hotwired/stimulus).
- The handbook says Turbo works without a backend framework, but the only first-party integration is Rails (turbo-rails); no Go integration was found in primary sources.
- Livewire-likeness: medium; Streams give server-pushed partial updates. Client-only interactions need Stimulus, a second library with a controller-per-behavior model.
- Pairs with both templating options; Tailwind neutral.

## React / Vue / Svelte

### Full SPA

- Go serves a JSON (or similar) API plus a static bundle; routing, state and rendering move to the client. Requires Node for the build (Vite et al.). Go templating shrinks to a shell page. Auth, CSRF, validation and flash flows must cross the API boundary; this is the opposite of Livewire's server-driven model.
- Current versions: React v19.3.0 (2026-09-09), Vue v3.5.43 (2026-09-17), Svelte 5.57.2 (2026-10-06), SvelteKit 3.0.1, Vite v8.3.4. Sources: GitHub releases for [react](https://github.com/facebook/react/releases), [vue](https://github.com/vuejs/core/releases), [svelte](https://github.com/sveltejs/svelte/releases), [kit](https://github.com/sveltejs/kit/releases), [vite](https://github.com/vitejs/vite/releases).
- Tailwind v4.3.3 (2026-07-16) per [tailwindcss releases](https://github.com/tailwindlabs/tailwindcss/releases); first-class in Node toolchains.
- Inertia.js bridges server routing and SPA components. Its docs name one official server adapter (Laravel) and rely on community adapters otherwise; no Go adapter is mentioned ([Inertia server-side setup](https://inertiajs.com/server-side-setup)). Community Go adapter `romsar/gonertia`: about 266 stars, last pushed 2026-04-11 (GitHub API).

### Islands on server-rendered pages

- Go renders the pages; selected components hydrate on the client. Astro documents `client:load`, `client:idle`, `client:visible` and server islands; the page fetched does not say how to use Astro alongside a separate non-Node backend ([Astro islands](https://docs.astro.build/en/concepts/islands/)).
- A lighter route is Vite with a traditional backend: in production the backend template emits tags from `.vite/manifest.json`; in development it loads `@vite/client` from the dev server. Needs `build.manifest: true`, CORS and asset proxy settings, and a Go-side manifest reader (not provided by Vite) ([Vite backend integration](https://vite.dev/guide/backend-integration)).
- Cost: Node toolchain, manifest handling in Go, two dev servers. Pairs with both templating engines as mount points (for example `<div id=... data-props=...>`); Tailwind via the same Node build.

## Build tooling summary

- No Node required: no-JS, HTMX + Alpine, Datastar, Unpoly (CDN/vendored), Turbo (script tag).
- Tailwind can run without Node via standalone CLI executables ([Tailwind CLI docs](https://tailwindcss.com/docs/installation/tailwind-cli)), so "Tailwind from day one" does not by itself force Node.
- Node required: SPA and islands (React/Vue/Svelte). Optional for Alpine, Unpoly and Turbo, needed only if installed through npm.
- All script-tag options can be vendored into Go `embed` assets; that is a design option, not a documented requirement.

## Templating pairing notes

- All hypermedia options consume plain HTML, so none depends on templ vs `html/template`. The difference is ergonomics: templ components are typed functions returning fragments; `html/template` uses named `{{define}}` blocks. templ needs a code-generation step (a Go tool, not Node); the pages fetched do not describe the command.
- Alpine (`@click`, `:attr`) and Datastar (`data-on:click`, `data-signals`) attributes embed JS expressions; how each engine's contextual escaping treats them was not verified and should be prototyped if either is shortlisted.

## Open questions for the grilling ticket

1. Is a JS build step acceptable at all, or is "no Node" a hard requirement (the Tailwind standalone CLI makes that feasible)?
2. Does server push (the optional websocket module) come from the same library as page interactivity (Datastar SSE, Turbo Streams, htmx `hx-sse`) or stay separate?
3. How much Livewire-style stateful components versus stateless request/response fragments is wanted?
4. Is a public template repo compatible with Datastar Pro licensing, or is core-only enough?
5. Tolerance for htmx 4 being new (2026-08-28) versus long-established versions.

## Not verified

- Library bundle sizes (not stated on the pages fetched).
- Go-community adoption beyond GitHub stars; no survey source was consulted.
- templ's Datastar page contents.
- Whether htmx.org recommends 4.x over 2.x as the default.
