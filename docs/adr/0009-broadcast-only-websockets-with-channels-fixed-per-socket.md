# Broadcast-only websockets with channels fixed per socket

The realtime module pushes server-rendered fragments to open pages over websockets (`coder/websocket` behind an adapter, served by `app serve` on `/ws`), and the browser never sends a message on the socket. Each socket's channels are fixed by its connect URL and authorized once at upgrade; a layout's socket sits outside `<main>` and survives boosted navigation, while a page-specific channel gets its own connect element inside `<main>` and closes when the page is left. Keeping the socket outbound-only means every write still goes through htmx, a handler and an Action, with CSRF protection and rate limits, and fixing channels per socket means subscription lifetime is DOM element lifetime, with no custom JS and no unsubscribe tracking. Instances fan out over Valkey pub/sub, at-most-once; events are UI nudges, never the record.

## Considered Options

- **Server-Sent Events**: stdlib-only and enough for one-way broadcasts, rejected by preference for real websockets.
- **One socket per tab with subscribe/unsubscribe messages** (Laravel Echo): any page can join channels on the fly, but `hx-ws` sends only form-shaped JSON, so it needs custom JS around the raw socket and server-side unsubscribe tracking across boosted navigations.
- **Actions over the socket** (Livewire-style): a second way into Actions that bypasses the HTTP layer's CSRF protection, rate limits and middleware.
- **`gorilla/websocket`**: no release since 2024 and one reader and one writer per connection; `x/net/websocket`'s own docs point to `coder/websocket` and gorilla.
- **A separate realtime process** (Laravel Reverb): needed in PHP's request-per-process model, not in Go, where each connection already runs in its own goroutine.

## Consequences

- Any inbound message closes the socket with 1008; a feature that needs browser-to-server realtime input would have to revisit this decision.
- Missed broadcasts during a reconnect are lost; anything that must survive goes through the database or the job queue.
- Close codes drive htmx's reconnect: 1001 on shutdown reconnects, 1008 for an invalid session, an inbound message or a slow client does not.
- `ResponseWriter` wrappers must keep `Hijacker`, and production proxies must forward `Upgrade` headers.
- An open socket outlives its session, so the module re-validates the session and credential fingerprint on every 30-second ping.
