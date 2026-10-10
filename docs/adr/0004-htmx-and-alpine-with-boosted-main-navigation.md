# htmx and Alpine with boosted `<main>` navigation

Interactivity uses htmx 4 for every server interaction and Alpine for state that lives only in the browser, because both consume server-rendered HTML and keep state on the server the way Livewire does. Both are bundled from npm by the frontend toolchain ([ADR-0013](0013-vite-plus-frontend-toolchain-with-a-build-watch-dev-loop.md)). Every flow works as a plain form or link first; htmx enhances it through the same URL and Action, and a handler renders a fragment only for `HX-Request-Type: partial`. Navigation is boosted within each layout and swaps only that layout's `<main>`, so the sidebar's DOM and Alpine state persist; when a boosted request lands on the other layout (a link, a logout or a redirect such as an expired session), the server answers with `HX-Retarget: body`, `HX-Reselect: body` and `HX-Reswap: outerSync`, telling it apart by each layout's `<main>` having its own id in `HX-Target`. The page ships a strict Content-Security-Policy (`script-src 'self'`, no `unsafe-eval`), so Alpine is the `@alpinejs/csp` build and htmx features that evaluate JS (`hx-on`, `js:` values) are not used.

## Considered Options

- **React, Vue or Svelte (SPA or islands)**: moves state to the client, and every page needs a client-side rendering model next to templ.
- **Datastar**: closest to Livewire's model, but it has no `hx-boost` equivalent, so progressive enhancement is wired form by form.
- **Native `<dialog>`, `popover` and invoker commands instead of Alpine**: zero JS, but the starter would need a second way of doing client interactions once a project adds real client state.
- **Full page loads for navigation**: simplest, but every click flashes and drops sidebar state, unless it is persisted per element in `sessionStorage`.
- **Boosting the whole `<body>`**: no layout-crossing problem, but the sidebar is re-rendered and loses its state on every click.
- **A client-side layout guard that re-fetches on crossings**: one place in JS, but every crossing costs a discarded request and a full reload.
- **Standard Alpine build with `unsafe-eval`**: unrestricted expressions, but HTML injected past templ's escaping could then run arbitrary JS through Alpine attributes.

## Consequences

- Alpine expressions are limited to simple property access, assignment and operators; anything else goes in an `Alpine.data(...)` component in a `.js` file.
- The active navigation link is derived by Alpine from the current path, not rendered by the server, because the sidebar is never re-rendered during boosted navigation.
- Out-of-band swaps appear only in fragments, never in a layout: a layout region marked `hx-swap-oob` is dropped when a crossing swaps the whole body.
- Both layouts must share one `<head>`, since boosted swaps never replace it.
- The theme is a cookie the server renders as the `dark` class on `<html>`, so no inline script is needed to avoid a flash of the wrong theme.
