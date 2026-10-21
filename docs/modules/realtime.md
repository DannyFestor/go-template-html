# Realtime module

The Realtime module pushes server-rendered fragments to open pages over websockets on `/ws` ([ADR-0009](../adr/0009-broadcast-only-websockets-with-channels-fixed-per-socket.md)). It ships wired and running. Its pitfalls are in [toolchain.md](../toolchain.md#realtime-module).

## Where it lives

The module is an area in each layer it touches ([ADR-0002](../adr/0002-layered-architecture-with-actions.md)), plus one wiring file:

| Path | Holds |
| --- | --- |
| `internal/bootstrap/realtime.go` | The wiring: the channel registry, the composers that turn a broadcast into a channel and a fragment, the Valkey listener and the shutdown |
| `internal/handler/realtime` | The `/ws` handler, the channel registry and the in-process hub |
| `internal/adapter/websocket` | The `coder/websocket` adapter |
| `internal/adapter/broadcast` | `Broadcaster[B]`, which renders a fragment and publishes it |
| `internal/adapter/valkey/realtime_channels.go` | Valkey pub/sub on `<VALKEY_KEY_PREFIX>realtime:` |
| `internal/middleware/session_reader.go` | `SessionReader`, which re-reads a socket's session without restarting its idle lifetime |
| `internal/view/realtime` | The connect element, the banner and the toast, and the channel names |
| `internal/domain/broadcast.go` | `Announcement` and `UserNotice` |

`internal/adapter/discard` is not part of the module. Its `Broadcaster[B]` does nothing, and core code gets it in place of a real broadcaster once the module is gone.

## Adding a channel or a broadcast

1. Register the channel in `realtimeChannels()` in `internal/bootstrap/realtime.go`. `Public(name)` lets anyone subscribe. `Private(pattern, authorize)` needs a signed-in, verified user for whom `authorize` returns true. A `{param}` segment in the pattern matches one dot-separated segment and reaches `authorize` by name.
2. Add the fragment to `internal/view/realtime`. It must carry `hx-swap-oob`, because the connect element swaps nothing itself.
3. Add a domain type for the broadcast and a composer that maps it to a `broadcast.Message`, then build a `broadcast.NewBroadcaster(compose, channels, reporter)` in `newRealtimeModule`.
4. An Action that broadcasts declares its own interface, `Broadcast(ctx context.Context, broadcast B)`, and calls it after its transaction commits. Bootstrap passes it the broadcaster.
5. Put the channel in a connect element: the layout's `realtime.Socket(...)` for every page, or a `realtime.Socket(...)` inside `<main>` for one page, which closes when the page is left.

## Removal checklist

Work through it top to bottom, then run the checks at the end.

1. Delete the module's own files:

   ```sh
   git rm -r internal/handler/realtime internal/adapter/websocket internal/adapter/broadcast internal/view/realtime \
     internal/bootstrap/realtime.go internal/adapter/valkey/realtime_channels.go \
     internal/adapter/valkey/realtime_channels_test.go internal/domain/broadcast.go \
     internal/middleware/session_reader.go internal/middleware/session_reader_test.go \
     tests/feature/realtime_test.go tests/feature/socket_test.go tests/e2e/realtime_test.go docs/modules/realtime.md
   ```

2. In `internal/bootstrap/web.go`, delete the `Announcements`, `UserNotices` and `realtime` fields of `WebApp` and their `nil` lines in `NewWebApp`, the `app.realtime.closeSocketsOnShutdown(server)` line in `serveWebApp`, the `a.realtime.close()` block in `Close`, the `a.startRealtime(...)` line and the `realtime:` field in `route`, and the now unused `ctx` parameter of `route` and its argument in `NewWebApp`.
3. In `internal/bootstrap/web_routes.go`, delete the `realtime` field of `webRoutes` and the `route.Realtime` line in `newWebMux`. In `internal/bootstrap/web_auth.go`, delete the `authenticator` field of `authRoutes` and its line in `newAuthRoutes`.
4. Wherever bootstrap passed `Announcements` or `UserNotices` into an Action, pass `discard.Broadcaster[domain.Announcement]{}` or `discard.Broadcaster[domain.UserNotice]{}`, and move the type the Action broadcasts back into `internal/domain`. Delete the Action's broadcast call instead if nothing should replace it.
5. In `internal/view/layout/public.templ`, delete `@realtime.Socket(...)` and `@realtime.BannerSlot()`. In `internal/view/layout/app.templ`, delete `@realtime.BannerSlot()`, `@realtime.ToastStack()` and `@realtime.Socket(...)`. Delete the `internal/view/realtime` import from both, then run `make generate-templ`.
6. Delete the `htmx.org/dist/ext/hx-ws.js` import from `internal/asset/src/app.ts`, and `Realtime` from `internal/route/route.go`.
7. In `internal/adapter/valkey/session_store.go`, delete `Peek`. Delete the three `Peek` tests in `internal/adapter/valkey/session_store_test.go` and the `Peek` method of `memoryStore` in `internal/middleware/session_test.go`.
8. In `.go-arch-lint.yml`, delete the `websocket` vendor, the `handler-realtime` and `adapter-websocket` components, their `deps` entries, both from bootstrap's `mayDependOn`, and `websocket` from feature-test's `canUse`. In `.golangci.yml`, delete `github.com/coder/websocket` from `gomodguard_v2` and its `importas` alias.
9. Run `go mod tidy`, which drops `github.com/coder/websocket`.
10. Delete the `## Realtime module` section of `docs/toolchain.md`, the `hx-ws` bullet of its htmx and Alpine section and `realtime:` from its Valkey section, the `## coder/websocket` section of `docs/upgrading.md` and the `hx-ws` mention in its htmx section, and the `### Realtime` section of `GLOSSARY.md`. Leave ADR-0009 as the record of the decision.
11. Run `make generate drift-check check vite-build`, then `make services test-integration test-feature test-e2e`. All of them pass with the module gone.
