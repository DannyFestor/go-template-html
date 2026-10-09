# Websocket module cost in a Go SSR app

Research for issue #7 (map #1). Researched 2026-10-09. Facts only; the decision belongs to a later grilling ticket.

Vocabulary follows `GLOSSARY.md`: **Optional module**, **Public channel**, **Private channel**.

Conventions in this file:

- **Fact** statements cite a primary source (library docs or source, RFC, spec, first-party docs).
- **Estimate** and **Inference** mark my own derivations. They are not sourced facts. The repo has no application code yet, so every footprint number is an estimate.
- Versions and dates are as of 2026-10-09, from the GitHub API and pkg.go.dev.

## Summary

- Websocket is not the only way to get realtime into an SSR page. Server-Sent Events (SSE) need no library (stdlib `http.Flusher` or `http.ResponseController`), carry cookies, and are what Datastar and htmx 4's `hx-sse` consume. Websocket is only required for browser-to-server messages over the same socket.
- Library choice for websocket: `coder/websocket` (zero dependencies, `context`-first, v1.8.15 released 2026-06-15) or `gorilla/websocket` (v1.5.3, last release 2024-06-14, last push 2025-03-19, one dependency). `golang.org/x/net/websocket` itself points users to those two.
- Private channel authorization with the existing session is straightforward: the browser sends cookies on the handshake (RFC 6455 section 4.1, 10.5), so the same session middleware can authenticate the upgrade request. The browser `WebSocket` constructor has no headers parameter, so cookies (or a URL token) are the only credential channels. Origin must be checked explicitly: Go's `CrossOriginProtection` exempts GET, which is what a handshake is.
- Fan-out across instances via Redis pub/sub works as expected but is at-most-once; a subscriber that is disconnected loses messages. go-redis `PubSub` reconnects and resubscribes automatically. This matches ADR-0001, which already names realtime fan-out as a Redis use.
- Frontend: htmx 4.0.0 (released 2026-08-28) ships `hx-ws` and `hx-sse` as first-party extensions that swap HTML sent by the server. Datastar speaks SSE only. Alpine or a JS framework would use the browser APIs directly or a client library. Templating (templ vs `html/template`) barely matters for this module: both render to an `io.Writer`.
- Estimated footprint: roughly 400 to 600 lines of Go (without tests) in one package, one new dependency for websocket (zero for SSE-only), one script tag or attribute in the layout, no new Docker service.
- Removal is clean only if domain code never calls the module directly. Anything else that publishes events must go through a small interface with a no-op default, or removing the module means editing feature code.

## 1. Transport options

### 1.1 Library comparison

| | `coder/websocket` | `gorilla/websocket` | `golang.org/x/net/websocket` | SSE (stdlib) |
|---|---|---|---|---|
| Latest release | v1.8.15, 2026-06-15 [C1][C2] | v1.5.3, 2024-06-14 [G1][G2] | part of `x/net` | n/a |
| Repo state | active; maintained by Coder since 2024 (nhooyr authored 2019-2024) [C1] | not archived; last push 2025-03-19 [G2] | package docs recommend the two libraries instead [X1] | n/a |
| License | ISC [C1] | BSD-2-Clause [G1] | not read | n/a |
| Module Go version | 1.23 [C3] | 1.20 [G3] | n/a | n/a |
| Dependencies | zero [C1] | `golang.org/x/net` v0.26.0 [G3] | n/a | none |
| Concurrency | concurrent writes supported [C1] | one concurrent reader and one concurrent writer; `Close` and `WriteControl` may be called concurrently [G4] | n/a | n/a |
| `context.Context` | first-class [C1] | not in API [G4] | no | request context |
| Origin check | rejects cross-origin by default; `OriginPatterns` or `InsecureSkipVerify` to relax [C2][C4] | default `CheckOrigin` rejects when `Origin` host differs from `Host`; a package-level `Upgrade` function does not check at all [G4] | default only checks that `Origin` is a valid URL [X1] | same-origin by cookie policy |
| Compression | RFC 7692 permessage-deflate, off by default [C1][C2] | permessage-deflate in no-context-takeover mode only [C1] | none | HTTP-level |
| HTTP/2 | not supported; roadmap item #4 [C5] | not documented | not documented | native |
| Handshake mechanism | `http.Hijacker`, errors if the `ResponseWriter` does not implement it [C4] | hijack | hijack | `Flusher` |

Notes:

- The `coder/websocket` README states "Compiles to Wasm", ping/pong API, `CloseRead` helper for write-only connections, and `wsjson` helpers [C1][C2]. The `gorilla/websocket` README lists prepared writes and configurable buffer sizes as its advantages relative to `coder/websocket` [C1].
- Both libraries pass the Autobahn test suite per their READMEs [C1][G1].
- `coder/websocket` docs warn that using `r.Context()` after `Accept` returns can misbehave because of hijacking [C2][C4]. Handlers need their own context.
- Not evaluated: higher-level servers such as Centrifugo/centrifuge, which bring their own protocol, client SDK and server process. They conflict with the stdlib-first constraint by construction, but I did not read their docs.

### 1.2 SSE as the alternative

- SSE is a standard: `EventSource`, `text/event-stream`, UTF-8, automatic reconnection after a delay, server-set delay via `retry`, `Last-Event-ID` request header on reconnect when an `id` was seen, HTTP 204 to stop reconnection [S1].
- `EventSource` accepts `{ withCredentials: true }` to send credentials cross-origin [S1]. Same-origin requests carry cookies by default (Inference; the spec text I read only describes the cross-origin switch).
- The spec warns that clients that honour HTTP's per-server connection limit can have trouble when a site opens several pages that each hold an `EventSource` to one domain, and suggests unique domains, a per-page toggle, or a shared worker [S1].
- The spec recommends a comment line (`:`) "every 15 seconds or so" because legacy proxies may drop idle connections [S1].
- Stdlib support: `http.ResponseController` exposes `Flush`, `SetWriteDeadline`, `SetReadDeadline`, `Hijack`, `EnableFullDuplex` [H1]. `http.Server.WriteTimeout` is "the maximum duration before timing out writes of the response" and "does not let Handlers make decisions on a per-request basis" [H2]. Inference: a global `WriteTimeout` would cut long-lived SSE or websocket-adjacent responses unless the handler extends its own deadline with `ResponseController.SetWriteDeadline` (SSE) or the connection is hijacked (websocket, where the hijacked connection leaves the server's per-request timeouts behind; I did not find this stated in the docs I read, so verify before relying on it).
- One-way only: client-to-server messages go through ordinary HTTP requests. htmx's SSE docs say the same: "SSE is a uni-directional service" [HX3].
- No third-party code is needed for the server side. The Datastar Go SDK (`datastar-go`, MIT, Go 1.24+) is optional sugar that writes the SSE framing [D4].

### 1.3 Websocket properties relevant to a handshake

- RFC 6455 does not prescribe how servers authenticate clients during the handshake; they may use "any mechanism a generic HTTP server could", with cookies named as an example [R1 section 10.5, 4.1].
- Servers "SHOULD verify the Origin field is an origin they expect" and answer 403 otherwise. The check defends against untrusted JavaScript making a browser open the socket; it does not stop non-browser clients [R1 section 10.2, 4.2.2].
- A handshake lacking `Origin` "SHOULD NOT be interpreted as coming from a browser client" [R1 section 4.2.1]. Consistent with that, `coder/websocket` allows requests with no `Origin` header [C4].
- Keepalive: a receiver MUST answer Ping with Pong; unsolicited Pong may act as a one-way heartbeat [R1 section 5.5.2, 5.5.3]. `coder/websocket`'s `Ping` must run concurrently with a reader [C2].
- The browser constructor is `new WebSocket(url, protocols)`; it has no options or headers parameter [M1]. Inference: the credentials available to the handshake are cookies, the URL, and `Sec-WebSocket-Protocol` values.

## 2. Authorizing private channels with the existing session

### 2.1 Laravel reference model

- Public channel: `Echo.channel('orders.1')`. Private channel: `Echo.private('orders.1')` [L1].
- For a private channel, Echo makes an HTTP request to the app, by default `/broadcasting/auth`, with the channel name; the app decides with a callback registered per channel pattern in `routes/channels.php`, for example `Broadcast::channel('orders.{orderId}', fn (User $user, int $orderId) => ...)`. The callback receives the authenticated user plus wildcard parameters and returns true or false [L1].
- Under the Pusher protocol the server answers the auth request with an HMAC-SHA256 signature of `<socket_id>:<channel_name>` using the app secret, returned as `<key>:<signature>` [P1]. Whether Reverb uses exactly this wire protocol was not verified; the Reverb docs I read do not state it.
- Reverb is a separate long-running process on its own port, proxied at `/app` and `/apps` [L2]. Origins are restricted by `allowed_origins` in `config/reverb.php` [L2].

### 2.2 What the Go equivalent can use

Facts:

- The handshake request is an ordinary HTTP request that carries cookies [R1 section 4.1, 10.5]; the htmx 2 `ws` extension docs say an omitted scheme falls back to the page's scheme, host and port "so browsers send cookies" [HX5]; the htmx 4 `hx-ws` source resolves relative URLs against `window.location` [HX2].
- ADR-0001 stores sessions in Redis, so a handshake handler placed behind the session middleware can read the user without a new auth mechanism.
- `coder/websocket.Accept` takes `(w, r, *AcceptOptions)` and returns `*Conn`; `Subprotocols` negotiates the sub-protocol; `OriginPatterns` authorizes cross-origin hosts [C2].
- `net/http.CrossOriginProtection` rejects non-safe cross-origin browser requests using `Sec-Fetch-Site` or `Origin` vs `Host`, and "The GET, HEAD, and OPTIONS methods are safe methods and are always allowed" [H3]. A websocket handshake is a GET [R1 section 4.1], so it passes this protection. Origin therefore must be checked by the websocket library (default on in both libraries listed above, except gorilla's deprecated package-level `Upgrade`) [C4][G4].
- Subprotocol-as-token is possible because `Sec-WebSocket-Protocol` values are readable server-side [M1]. Not needed when cookies work; relevant only for cross-origin socket hosts (a separate ws subdomain).

Where authorization can happen (all three are possible with either library; none is documented by Laravel as the only way):

1. At upgrade time, per URL: `/ws/orders/42` runs the authorizer before `Accept`. One socket per channel. Failed authorization is a plain HTTP 401/403 before the upgrade.
2. At subscribe time, per message: one socket per page; the client sends `{"subscribe": "private-orders.42"}` and the server runs the authorizer for that channel name against the session user captured at upgrade. Closest to Laravel's flow; requires a small message protocol (websocket) or a separate HTTP call.
3. SSE: one `EventSource`/fetch stream per channel (or a comma list in the query string, as the htmx SSE docs note parameters can ride in the URL [HX3]); authorize before writing headers.

Inference (not a documented fact): a live connection outlives the session unless the module re-checks or closes it. A logout, password change or session expiry in Redis does not by itself close an open socket. Options include a periodic session check, or publishing a "session revoked" message on the Redis bus that the hub uses to close that user's connections. This is a design point for the grilling ticket.

### 2.3 Public channels

No authorization step; origin check still applies (RFC 6455 section 10.2 [R1]). Same transport code with an authorizer that always allows, or no authorizer registered for the channel pattern.

## 3. Fan-out across instances with Redis pub/sub

Facts from Redis docs [RD1]:

- Delivery is at-most-once: "If the subscriber is unable to handle the message (for example, due to an error or a network disconnect) the message is forever lost." Redis Streams are the documented option for stronger guarantees.
- Pub/sub is unrelated to the keyspace and database numbers; "prefix the channels with the name of the environment" for scoping.
- `PSUBSCRIBE` supports glob patterns; a client subscribed to both a channel and a matching pattern receives the message twice.
- Under RESP2 a subscribed connection may only issue subscribe-family commands, `PING`, `QUIT`, `RESET`; under RESP3 it can issue any command.
- Sharded pub/sub (`SSUBSCRIBE`, `SPUBLISH`) exists from Redis 7.0 and matters for Redis Cluster only.

Facts from go-redis v9 source (`pubsub.go`, current release v9.23.0, 2026-10-05; module requires Go 1.26.0) [RG1][RG2]:

- `PubSub` "automatically reconnects to Redis Server and resubscribes to the channels in case of network errors". Receiving is not safe for concurrent use.
- `Channel()` returns a Go channel; "If the Go channel is blocked full for 1 minute the message is dropped"; defaults: buffer 100, health-check ping every 3 s, send timeout 60 s, ping timeout 5 s, reconnect timeout 10 s (all configurable with `WithChannel...` options).
- `ChannelWithSubscriptions` can be used to detect reconnections (messages published during the gap are still lost, per [RD1]).

Resulting shape (Inference, standard hub pattern):

- Each app instance holds one Redis subscription connection and an in-process hub mapping channel name to local connections.
- Publishing from any instance goes to Redis `PUBLISH`; every instance (including the publisher) receives it and writes to its own local subscribers. This is the same arrangement Laravel documents for Reverb scaling: Redis pub/sub relays messages between Reverb servers, behind a load balancer [L2].
- A slow client must not block the hub: per-connection buffered send queue with drop-or-disconnect policy. The libraries do not decide this for you.
- Redis is already a day-one dependency per ADR-0001 ("Redis also serves later needs such as rate limiting and realtime fan-out"), so no new infrastructure.
- Postgres `LISTEN`/`NOTIFY` as a Redis-free alternative was not researched.

## 4. Frontend subscription and the interactivity layer

The server sends one of: rendered HTML fragments, or JSON events. The first fits an SSR app and removes client rendering code.

### 4.1 Templating engine (templ vs `html/template`)

- templ: `Component` is `Render(ctx context.Context, w io.Writer) error` [T1].
- `html/template`: `Template.ExecuteTemplate(wr io.Writer, name string, data any) error`; templates "may be executed safely in parallel, although if parallel executions share a Writer the output may be interleaved" [T2].
- Inference: either one can render a fragment into a `bytes.Buffer` and hand the bytes to the hub, so the broadcast path is the same for both. The hub should receive pre-rendered bytes (or a render function) rather than depend on a template engine. Per-user fragments must be rendered per subscriber, not once; whether a channel may carry user-specific HTML is a design choice.

### 4.2 By interactivity layer

| Layer | Websocket | SSE | Notes |
|---|---|---|---|
| htmx 4.0.0 (2026-08-28) | `hx-ws` extension: `hx-ws:connect="/chat"`, incoming messages swapped using htmx swap rules, `hx-swap-oob` and `<hx-partial>` for multi-target updates, `hx-ws:send` sends form data as JSON with an htmx `headers` object, reconnect with jitter [HX2] | `hx-sse` extension: stream via `hx-get`/`hx-post`, or persistent `hx-sse:connect="/events"`, `hx-sse:close`, `Last-Event-ID` sent on reconnect, `reconnectDelay` 500 ms, `reconnectMaxDelay` 60 s, `reconnectMaxAttempts` unlimited by default in source [HX1][HX4] | Both are separate scripts under `dist/ext/` in the v4.0.0 tag, loaded after core. Source for hx-sse shows it uses `fetch` with a stream parser (not `EventSource`), which allows non-GET methods [HX4]. |
| htmx 2.x | `ws` extension (`htmx-ext-ws`, docs show 2.0.4): `ws-connect`, `ws-send`; messages swapped by element `id` using OOB logic; reconnect with full-jitter exponential backoff [HX5] | `sse` extension (`htmx-ext-sse`, docs show 2.2.4): `sse-connect`, `sse-swap`, named events only (no catch-all), own exponential backoff on top of the browser's [HX3] | Extension attribute names differ between 2.x and 4.x (`ws-connect` vs `hx-ws:connect`), so a module's templates are tied to the htmx major version. |
| Datastar (core v1.0.4, 2026-09-21; Go SDK `datastar-go` MIT, Go 1.24+ [D4]) | Not supported; its guide describes fetch and SSE only [D1] | Native: `data-init="@get('/endpoint')"` opens a stream; events `datastar-patch-elements` and `datastar-patch-signals` [D1][D2][D3]. `@get` retry options: `retry` default `auto`, `retryInterval` 1000 ms, `retryScaler` 2, `retryMaxWait` 30000 ms, `retryMaxCount` 10; `openWhenHidden` default false for GET, so the stream closes while the tab is hidden and reopens when visible [D5] | Websocket module would not apply; an SSE-only module is the match. `credentials` and `Last-Event-ID` handling are not documented on the pages I read. |
| Alpine.js (alongside htmx) | Plain browser `WebSocket` inside an `x-data`/`x-init` block; Alpine has no websocket feature of its own in the docs I read; `x-init` runs code at initialization [A1] | Plain `EventSource` the same way | htmx 4 ships an Alpine compatibility extension (`hx-alpine-compat`) [HX1]; its relevance to websocket content swaps was not read. Cleanup on element removal: not documented on the page I read. |
| JS framework (React, Vue, Svelte, ...) | Browser `WebSocket`, or a library; Laravel's reference uses Echo with Reverb, Pusher or Ably [L3] | `EventSource` | Client rendering of JSON events; the SSR template then needs a JSON event contract, not HTML fragments. |

Browser behavior that matters for all rows: `new WebSocket(url, protocols)` takes no headers [M1]; SSE is bound by the per-server connection limit warning in [S1].

### 4.3 Laravel Echo with a Go server

- Echo's supported broadcasters are Reverb, Pusher and Ably [L3]. For Pusher, the client flow is: connect, get a socket ID, POST `socket_id` and `channel_name` to an auth endpoint, receive `key:signature` [P1]; Echo's Reverb support is described as sharing Echo's client setup in [L1], but I did not verify Reverb's wire protocol.
- Inference: reusing Echo against a Go server would require implementing whatever wire protocol Echo's Reverb or Pusher broadcaster speaks (at least the auth signature above) in Go. That is more code than a bespoke thin protocol, and I found no primary source suggesting a stdlib-sized Go implementation exists. Not researched further.

## 5. Laravel reference experience

Source: Laravel 12.x docs [L1][L2] (the docs site marks 12.x as old; 13.x exists, so re-check if parity with latest matters).

- Opt in: `php artisan install:broadcasting` (broadcasting is "not enabled in new Laravel applications"). It prompts for a broadcaster, creates `config/broadcasting.php` and `routes/channels.php`; with `--reverb` it installs Reverb's Composer and NPM packages and runs `reverb:install`; Echo scaffolding and configuration are injected into the app [L1][L2].
- Config: `BROADCAST_CONNECTION` environment variable selects the driver; Reverb credentials are `REVERB_APP_ID`, `REVERB_APP_KEY`, `REVERB_APP_SECRET` [L1][L2].
- Server: `php artisan reverb:start` (default `0.0.0.0:8080`) as a separate long-running process; behind a reverse proxy it is routed from 443; scaling horizontally is `REVERB_SCALING_ENABLED=true` plus a central Redis [L2].
- Events: a class implementing `ShouldBroadcast` with a `broadcastOn()` method returning `Channel`, `PrivateChannel` or `PresenceChannel`; dispatching the event broadcasts it. Broadcasting goes through a queue worker unless the event uses `ShouldBroadcastNow` [L1].
- Authorization: `routes/channels.php` callbacks as in section 2.1. `php artisan channel:list` lists them [L1].
- Client: `Echo.channel(name).listen('Event', cb)` and `Echo.private(name).listen(...)` [L1].
- Removal in Laravel: not documented as a command in the pages read.

What the "same experience" would map to in a Go template (Inference, for the grilling ticket): one switch to enable (config flag plus mount line), one file of channel authorization callbacks, one `Broadcast(channel, event)` call from feature code, one client attribute or snippet per page. Go has no `artisan`; the map's note says reuse is by cloning with no scaffolding CLI, so "install" would be either shipped-on-by-default with a feature flag, or ship-present and deleted by the developer if unwanted.

## 6. Code footprint, touch points and removal

All figures are my estimates for a module of this scope; no code exists in the repo yet.

### 6.1 Module internals (one package, for example `internal/realtime`)

| Part | Responsibility | Estimated Go lines |
|---|---|---|
| Hub | channel name to local subscribers; join, leave, fan-out to local conns; per-conn bounded send queue | 80 to 130 |
| Redis bridge | one `PubSub`, publish function, forward incoming messages to the hub, prefix channels with environment | 50 to 80 |
| Channel authorizer registry | pattern to callback, public vs private, wildcard parameters (Laravel's `Broadcast::channel` equivalent) | 50 to 90 |
| Transport adapter(s) | websocket handler (accept, origin options, read loop, ping, close) and/or SSE handler (headers, flush, keepalive comment) | 60 to 120 each |
| Publisher interface and no-op | what feature code calls; no-op when the module is removed or disabled | 15 to 30 |
| Example routes | one public-channel page and one private-channel page with their fragments | 60 to 120 |
| Config struct and wiring | feature flag, allowed origins, Redis channel prefix | 30 to 50 |

Total: roughly 350 to 650 lines excluding tests; tests (hub, authorizer, a Redis integration test using the Docker service) would roughly double it. Client-side JS written by the project: zero for htmx or Datastar (attributes only), a small snippet for Alpine or framework use.

Dependencies added: websocket path: one (`coder/websocket`, zero transitive dependencies [C1], or `gorilla/websocket` with `x/net` [G3]); SSE-only: none, apart from go-redis, which the Redis session store may already require (the session library is decided in another ticket). go-redis v9.23.0 declares Go 1.26.0 in its `go.mod` [RG1], compatible with Go 1.27.

### 6.2 Touch points outside the module

| Area | Touch | Notes |
|---|---|---|
| Router | mount `GET /ws` (and/or `/events`) handler(s) and the example pages | one mount function, so one line to remove |
| Middleware | handshake runs behind the existing session and authentication middleware (private channels); public channel mounted outside it | must also respect any request-logging or compression middleware that wraps `ResponseWriter` without implementing `Hijacker` or `Flusher`: `coder/websocket.Accept` errors if the writer does not implement `http.Hijacker` [C4]. `http.ResponseController` can unwrap writers that expose `Unwrap` (Inference from its API [H1]; verify against the controller docs before relying on it). |
| HTTP server config | `WriteTimeout` applies per response [H2] and affects SSE; per-handler deadline control via `ResponseController` [H1] | a global setting owned by the app, not the module |
| CSRF / origin | `CrossOriginProtection` (if adopted) does not cover GET [H3]; origin allow-list must live in the module's `Accept` options | config: allowed origin patterns |
| Layout templates | load the client extension script (`hx-ws`/`hx-sse` for htmx 4; `ws`/`sse` for htmx 2; none for Datastar, which is SSE in core) and put the connect attribute on the page or a shared partial | one script tag in the base layout if loaded globally; or only on pages that use it |
| Config / env | feature flag for the module, Redis channel prefix, allowed origins | the map distinguishes feature flags (developer) from security settings (user) per `GLOSSARY.md`; this is a feature-flag-level switch |
| Docker (dev) | none new; Postgres and Redis are already there (ADR-0001) | unlike Reverb, which is a separate process on its own port [L2] |
| Production reverse proxy | must forward `Upgrade` and `Connection` headers for websocket; Laravel's Nginx example sets `proxy_http_version 1.1`, `Upgrade` and `Connection "Upgrade"` [L2]; SSE needs response buffering off (Inference: standard proxy behavior, not read from a primary source here) | outside the template (host-specific deploy configs are out of scope per the map) |
| OS limits | each connection is an open file; Laravel's docs call out `ulimit -n` and proxy connection limits [L2] | only matters at scale |
| Graceful shutdown | long-lived connections delay `http.Server.Shutdown` unless the module closes them on a shutdown signal | Inference; part of the module, hooked from `main` |
| CI | Redis integration tests need the Redis service | CI job already needs Redis for sessions |

### 6.3 What removing the module requires

With the module as a self-contained package behind an interface, removal is:

1. Delete the package directory and its tests.
2. Delete the mount call in the router and the module's config fields and env vars.
3. Delete the script tag or attribute from the layout, plus the example pages and their nav links.
4. `go mod tidy` (drops `coder/websocket` or `gorilla/websocket` if nothing else uses them; go-redis stays if sessions use it).
5. If feature code called the publisher interface, either keep the no-op implementation or delete those call sites.

Items 1 to 4 are mechanical. Item 5 is the risk: the glossary defines an Optional module as removable "without touching unrelated code". That holds only if the publisher interface sits in a shared location with a no-op default, or no shipped feature code publishes events (the in-scope routes in the map, such as auth and settings, have no obvious need to broadcast). Whether any template feature should use realtime (for example, logging a user out on password change) is undecided.

## 7. Facts that vary with open decisions

- **Templ vs `html/template`**: no impact on the transport or hub (section 4.1). Impact only on how fragments are rendered and where the render helper lives.
- **htmx 2 vs 4 vs Datastar vs JS framework**: determines which transport is natural (websocket for htmx and frameworks; SSE only for Datastar), the client attribute names, and whether fragments or JSON go over the wire (section 4.2). htmx 4's extension names and syntax differ from htmx 2.
- **Session store library**: determines whether go-redis is already a dependency, and how the handshake reads the session.
- **Auth model for subscribe** (upgrade-time vs subscribe-time, section 2.2): affects message protocol size, not the library.

## 8. Not covered or not verified

- Postgres `LISTEN`/`NOTIFY` as a fan-out alternative.
- Centrifugo/centrifuge and other full realtime servers.
- Presence channels (who is online); Laravel has them [L1]; the glossary has only public and private.
- Exact Pusher wire protocol details beyond the signature scheme.
- Whether `net/http` timeouts apply to hijacked connections; verify before choosing server timeouts.
- HTTP/2 websockets (RFC 8441): `coder/websocket` lists HTTP/2 as unimplemented [C5]; effect on deployments that terminate HTTP/2 at a proxy was not researched.
- Alpine's cleanup of WebSocket/EventSource objects when elements are removed.
- Performance and memory per connection for either library (no benchmark read; only README claims).

## Sources

Library docs and repositories:

- [C1] coder/websocket README. https://github.com/coder/websocket
- [C2] coder/websocket package docs, v1.8.15. https://pkg.go.dev/github.com/coder/websocket
- [C3] coder/websocket `go.mod` (`go 1.23`). https://github.com/coder/websocket/blob/master/go.mod
- [C4] coder/websocket `accept.go` (Hijacker requirement, `authenticateOrigin`, empty `Origin` allowed). https://github.com/coder/websocket/blob/master/accept.go
- [C5] coder/websocket README roadmap, "HTTP/2 #4". https://github.com/coder/websocket#roadmap
- [G1] gorilla/websocket README. https://github.com/gorilla/websocket
- [G2] gorilla/websocket releases and repo metadata via GitHub API (latest v1.5.3, 2024-06-14; not archived; last push 2025-03-19). https://github.com/gorilla/websocket/releases
- [G3] gorilla/websocket `go.mod` (`go 1.20`, requires `golang.org/x/net`). https://github.com/gorilla/websocket/blob/main/go.mod
- [G4] gorilla/websocket package docs, v1.5.3 (concurrency, origin, buffers). https://pkg.go.dev/github.com/gorilla/websocket
- [X1] `golang.org/x/net/websocket` package docs. https://pkg.go.dev/golang.org/x/net/websocket

Standards:

- [R1] RFC 6455, The WebSocket Protocol (sections 1.3, 4.1, 4.2.1, 4.2.2, 5.5.2, 5.5.3, 10.2, 10.5). https://datatracker.ietf.org/doc/html/rfc6455
- [S1] WHATWG HTML Living Standard, Server-sent events (section 9.2). https://html.spec.whatwg.org/multipage/server-sent-events.html
- [M1] MDN, `WebSocket()` constructor. https://developer.mozilla.org/en-US/docs/Web/API/WebSocket/WebSocket (reference documentation, not a standard; used only for the constructor signature)

Go standard library:

- [H1] `net/http.ResponseController`. https://pkg.go.dev/net/http#ResponseController
- [H2] `net/http` `server.go`, `Server.WriteTimeout` comment. https://github.com/golang/go/blob/master/src/net/http/server.go
- [H3] `net/http.CrossOriginProtection` docs and `csrf.go`. https://pkg.go.dev/net/http#CrossOriginProtection , https://github.com/golang/go/blob/master/src/net/http/csrf.go
- [T2] `html/template` `Execute` and `ExecuteTemplate`. https://pkg.go.dev/html/template#Template.ExecuteTemplate

Redis:

- [RD1] Redis Pub/sub documentation. https://redis.io/docs/latest/develop/pubsub/
- [RG1] go-redis v9 `pubsub.go` and `go.mod`. https://github.com/redis/go-redis/blob/master/pubsub.go
- [RG2] go-redis releases (v9.23.0, 2026-10-05). https://github.com/redis/go-redis/releases

Frontend:

- [HX1] htmx v4.0.0 release and tag tree (`dist/ext/hx-ws.js`, `hx-sse.js`, `hx-alpine-compat.js`). https://github.com/bigskysoftware/htmx/releases/tag/v4.0.0
- [HX2] htmx 4 `hx-ws` extension docs and source. https://github.com/bigskysoftware/htmx/blob/v4.0.0/www/src/content/extensions/03-hx-ws.md , https://github.com/bigskysoftware/htmx/blob/v4.0.0/src/ext/hx-ws.js
- [HX3] htmx 2 SSE extension docs. https://htmx.org/extensions/sse/
- [HX4] htmx 4 `hx-sse` extension docs and source. https://github.com/bigskysoftware/htmx/blob/v4.0.0/www/src/content/extensions/02-hx-sse.md , https://github.com/bigskysoftware/htmx/blob/v4.0.0/src/ext/hx-sse.js
- [HX5] htmx 2 WebSocket extension docs. https://htmx.org/extensions/ws/
- [D1] Datastar guide, getting started. https://data-star.dev/guide/getting_started
- [D2] Datastar SSE events reference. https://data-star.dev/reference/sse_events
- [D3] Datastar attributes reference (`data-init`). https://data-star.dev/reference/attributes
- [D4] Datastar Go SDK. https://github.com/starfederation/datastar-go
- [D5] Datastar actions reference (`@get` options). https://data-star.dev/reference/actions
- [A1] Alpine.js `x-init`. https://alpinejs.dev/directives/init
- [T1] templ `Component` interface. https://pkg.go.dev/github.com/a-h/templ#Component

Laravel (reference experience):

- [L1] Laravel 12.x Broadcasting. https://laravel.com/docs/12.x/broadcasting
- [L2] Laravel 12.x Reverb. https://laravel.com/docs/12.x/reverb
- [L3] Laravel Echo repository. https://github.com/laravel/echo
- [P1] Pusher Channels authentication signatures. https://pusher.com/docs/channels/library_auth_reference/auth-signatures/
