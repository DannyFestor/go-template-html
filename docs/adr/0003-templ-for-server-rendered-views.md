# templ for server-rendered views

Views are written in templ, not `html/template`, despite the stdlib-first constraint: templ turns pages, layouts and components into typed Go functions, so a wrong field or missing argument fails at compile time instead of rendering blank at runtime, and layouts compose as plain components instead of per-page template sets. Views live in `internal/view` (`layout`, `component`, one package per area) and receive view models built by handlers, never `domain` or `action` types; the data every page shares comes from one `ShellBuilder` in `internal/http` and is passed explicitly, not read from `ctx`.

## Considered Options

- **`html/template`**: no dependency and no codegen, but data is untyped, layouts need a template set per page, and dev reload has to be built by hand.
- **templ behind an adapter**: the views are templ, so swapping engines means rewriting them whatever sits in front. Only `internal/view` imports templ; `internal/http` renders through its own `Render(ctx, io.Writer) error` interface.
- **Shared page data from `ctx`** (Blade's `auth()`, Inertia's shared props): less plumbing, but hides each component's real inputs and couples views to middleware-owned context keys.
- **Gitignoring generated code**: one source of truth, but `go build` silently compiles a stale `*_templ.go` after a pull, and every build, test, lint and image path must run `go generate` first.

## Consequences

- All generated code (templ, sqlc, mocks) is committed, marked `linguist-generated` in `.gitattributes`, and CI fails when `go generate ./...` produces a diff.
- Contributors need the templ editor plugin for highlighting and completion in `.templ` files.
