# Research: Go building blocks for sessions, CSRF, rate limiting, hashing and two-factor authentication

Ticket: [#5](https://github.com/DannyFestor/go-template-html/issues/5). Researched 2026-10-09 against Go 1.27.2.

This document lists options and trade-offs. It does not decide anything; the decisions belong to the grilling ticket (#18).

Constraints applied: Go 1.27, stdlib-first (a dependency must replace security-sensitive or substantial code), sessions in Redis ([ADR-0001](../adr/0001-postgres-and-redis-as-day-one-infrastructure.md)), Postgres, Laravel Fortify / Livewire starter kit parity minus passkeys.

## Contents

1. [Reference behaviour: what Laravel Fortify does](#1-reference-behaviour-what-laravel-fortify-does)
2. [Server-side sessions in Redis](#2-server-side-sessions-in-redis)
3. [Session fixation, logout and remember-me](#3-session-fixation-logout-and-remember-me)
4. [CSRF](#4-csrf)
5. [Rate limiting login and two-factor authentication](#5-rate-limiting-login-and-two-factor-authentication)
6. [Password hashing](#6-password-hashing)
7. [Two-factor authentication (TOTP)](#7-two-factor-authentication-totp)
8. [Recovery codes](#8-recovery-codes)
9. [Signed URLs and tokens for email verification and password reset](#9-signed-urls-and-tokens-for-email-verification-and-password-reset)
10. [Stdlib primitives available in Go 1.27](#10-stdlib-primitives-available-in-go-127)
11. [Pitfall checklist](#11-pitfall-checklist)
12. [Open questions for the grilling ticket](#12-open-questions-for-the-grilling-ticket)

---

## 1. Reference behaviour: what Laravel Fortify does

These are the parity targets, read from source.

| Concern | Laravel behaviour | Source |
|---|---|---|
| Login throttle | 5/minute, keyed by `lower(email)|ip` | [livewire-starter-kit `FortifyServiceProvider`](https://github.com/laravel/livewire-starter-kit/blob/main/app/Providers/FortifyServiceProvider.php) |
| Two-factor challenge throttle | 5/minute, keyed by the pending `login.id` in the session | same file |
| TOTP replay | Caches `fortify.2fa_codes.<md5(code)>` with the accepted timestamp and uses `verifyKeyNewer`, so a code (or an older one) can't be used twice in the window | [`TwoFactorAuthenticationProvider::verify`](https://github.com/laravel/fortify/blob/1.x/src/TwoFactorAuthenticationProvider.php) |
| TOTP secret storage | Encrypted with the app key (`two_factor_secret`) | [`EnableTwoFactorAuthentication`](https://github.com/laravel/fortify/blob/1.x/src/Actions/EnableTwoFactorAuthentication.php) |
| Recovery codes | 8 codes, each `Str::random(10).'-'.Str::random(10)`, stored **encrypted** (not hashed) as one JSON column | same file; [`RecoveryCode`](https://github.com/laravel/fortify/blob/1.x/src/RecoveryCode.php) |
| Two-factor setup | `confirm => true` (user must enter a code before it is active), `confirmPassword => true` | [`stubs/fortify.php`](https://github.com/laravel/fortify/blob/1.x/stubs/fortify.php) |
| Password reset token | `hash_hmac('sha256', Str::random(40), key)`, stored **hashed** with the password hasher, expires 60 min, 60 s resend throttle | [`DatabaseTokenRepository`](https://github.com/laravel/framework/blob/13.x/src/Illuminate/Auth/Passwords/DatabaseTokenRepository.php); [`config/auth.php`](https://github.com/laravel/laravel/blob/13.x/config/auth.php) |
| Email verification | `URL::temporarySignedRoute` (HMAC-signed URL with `expires`), 60 min | [`VerifyEmail::verificationUrl`](https://github.com/laravel/framework/blob/13.x/src/Illuminate/Auth/Notifications/VerifyEmail.php) |
| Password confirmation | Valid for 10800 s (3 h) | `config/auth.php` `password_timeout` |
| Remember-me | Separate encrypted cookie `id|remember_token|hash(password hash)`, lifetime 576000 min (400 days); `remember_token` is 60 random chars on the users row, cycled on logout | [`SessionGuard`](https://github.com/laravel/framework/blob/13.x/src/Illuminate/Auth/SessionGuard.php), [`Recaller`](https://github.com/laravel/framework/blob/13.x/src/Illuminate/Auth/Recaller.php) |
| Password hashing | bcrypt, `BCRYPT_ROUNDS=12` | [`.env.example`](https://github.com/laravel/laravel/blob/13.x/.env.example) |

Note on the remember-me cookie: because it embeds a hash of the password hash, changing the password invalidates every remember-me cookie.

---

## 2. Server-side sessions in Redis

### Options

| Library | Version / status (2026-10-09) | Redis store | Notes |
|---|---|---|---|
| [alexedwards/scs](https://github.com/alexedwards/scs) v2 | v2.9.0 (2025-07-08), last push 2025-11, not archived | First-party sub-modules: `goredisstore` (go-redis v9) and `redisstore` (redigo) | Server-side only; opaque token in cookie |
| [gorilla/sessions](https://github.com/gorilla/sessions) | v1.4.0 (2024-08-20), no commits since | Third-party: [boj/redistore](https://github.com/boj/redistore) v2.0.2 (2026-09-29, redigo); [rbcervilla/redisstore](https://github.com/rbcervilla/redisstore) v9.0.0 (2023-04, go-redis v9) | Built around signed/encrypted cookies (`securecookie`); custom backends supported |
| Hand-rolled on [redis/go-redis](https://github.com/redis/go-redis) | go-redis v9.23.0 (2026-10-05), active | n/a | Roughly: token generation, cookie write, load/save middleware, `SET key val EX ttl` |

Redis client: [redis/go-redis](https://github.com/redis/go-redis) v9 is the official client and is very active. [gomodule/redigo](https://github.com/gomodule/redigo) v1.9.3 (2025-10) is maintained but older in style. Rate limiting (section 5) and the realtime module are both likely to use go-redis, so picking one client for everything avoids two Redis dependencies.

### alexedwards/scs facts (read from v2.9.0 source)

- Token: 32 bytes from `crypto/rand`, base64url ([`data.go` `generateToken`](https://github.com/alexedwards/scs/blob/master/data.go)). That is 256 bits, well above OWASP's 64-bit minimum.
- `HashTokenInStore bool`: if true, the Redis key is `sha256(token)` and not the raw token, so a Redis dump can't be replayed as cookies. Off by default.
- Defaults: `Lifetime` 24 h absolute, `IdleTimeout` 0 (none), cookie `HttpOnly=true`, `SameSite=Lax`, `Secure=false`, `Persist=true`, name `session` ([`session.go` `New`](https://github.com/alexedwards/scs/blob/master/session.go)). `Secure` must be set explicitly; there is no `__Host-` prefix by default, but the cookie name is configurable.
- API: `LoadAndSave` middleware (adds `Vary: Cookie`), `Put/Get/Pop/Remove`, `RenewToken`, `Destroy`, `RememberMe`, `Iterate`, `SetDeadline`, `MergeSession`.
- `goredisstore` key prefix `scs:session:`; Redis `EXPIRE` handles expiry. `Iterate`/`AllCtx` uses `SCAN prefix*` over **all** sessions. There is no per-user index.
- Default codec is `encoding/gob`; stored types must be gob-registered.

### gorilla/sessions facts

- Package doc: "provides cookie and filesystem sessions and infrastructure for custom session backends". The Redis backends are third-party.
- No method on `Session` for regenerating the ID. Fixation defence depends on the store (typically delete and recreate the session).
- `Values` is `map[interface{}]interface{}`, not type-safe.

### Trade-offs

- **scs + goredisstore**: least code. Has a fixation API (`RenewToken`), optional hashed keys, and go-redis v9. Weak spots: "log out other devices" needs either `Iterate` (a full keyspace SCAN) or an app-maintained per-user set of tokens; remember-me is only "persistent cookie with a long Lifetime" (see section 3); the store module pins an old go-redis (`v9.0.2` minimum; MVS will pick the app's newer version).
- **gorilla/sessions + redistore**: cookie-centric design, the core hasn't had a release in two years, the Redis store is third-party and uses redigo. No clear advantage for a server-side-only design.
- **Hand-rolled**: the security-sensitive parts are small and stdlib-covered (`crypto/rand.Text`, `http.Cookie`, `crypto/sha256`). The cost is the load/commit lifecycle (write the cookie before headers are flushed, handle idle and absolute deadlines, concurrency on the session map), which is what scs already provides. Under "a dependency must replace substantial code" scs qualifies, but only just. A hand-rolled version could add a per-user index (`SADD user:<id>:sessions`) for "log out other devices" and invalidate-on-password-change.

---

## 3. Session fixation, logout and remember-me

### Requirements (OWASP [Session Management Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/Session_Management_Cheat_Sheet.html))

- Session IDs: at least 64 bits of entropy from a CSPRNG.
- "The session ID must be renewed or regenerated by the web application after any privilege level change"; regeneration "is mandatory to prevent session fixation attacks". Privilege changes include login, logout, password change and role switch.
- Idle timeouts: "2-5 minutes for high-value applications and 15-30 minutes for low risk applications"; absolute timeouts example 4-8 h. Both enforced server-side.
- Cookie: `Secure`, `HttpOnly` (mandatory), `SameSite=Strict` (preferred) or `Lax`; `__Host-` prefix "Recommended for session IDs".
- Logout must invalidate server-side; "deleting a cookie alone does not invalidate a stolen copy of the session ID".
- [Login CSRF](https://cheatsheetseries.owasp.org/cheatsheets/Cross-Site_Request_Forgery_Prevention_Cheat_Sheet.html): a pre-session can't be promoted into the authenticated session; create a new one on login.

### Where renewal is needed in this template's flows

Renew the session token on: login (after password, and again after the two-factor authentication challenge completes), logout, password change, password reset, enabling/disabling two-factor authentication. In scs that is `RenewToken(ctx)`, called once per request before headers are written.

The two-factor authentication challenge is a half-authenticated state. Fortify stores `login.id` in the session and only logs the user in after the code verifies. Whatever stores that pending state must not count as authenticated for the authenticated middleware.

### Remember-me: two models

| Model | How | Trade-off |
|---|---|---|
| **A. Long-lived session** (scs `Cookie.Persist=false` + `RememberMe(ctx, true)`) | Persistent cookie only when the box is ticked; otherwise a browser-session cookie. The Redis TTL is still `Lifetime`. | Simple. But scs `Lifetime` is a single global absolute expiry, so a 30-day remember-me means every session (ticked or not) can live 30 days server-side unless `SetDeadline` is used per session. Idle timeout and remember-me conflict. |
| **B. Separate remember token** (Laravel model) | Short session, plus a long-lived `remember` cookie containing `selector:validator`; the DB stores `sha256(validator)` per device. When the session is missing, validate (constant-time), rotate the token, start a new session. | More code and a Postgres table. Matches Laravel and lets sessions stay short. Supports per-device revocation and invalidation on password change (delete rows). Store the validator hashed; Laravel stores `remember_token` in plaintext. |

OWASP's session cheat sheet does not specify a remember-me design ("Use non-persistent cookies when authentication does not need to persist").

### Logout other devices / invalidate on password change

OWASP's Forgot Password Cheat Sheet: after reset, offer to or automatically end existing sessions. Neither scs nor gorilla keeps a user-to-sessions index. Options: (a) scs `Iterate` (SCAN all sessions; fine at small scale, O(all sessions)); (b) maintain a Redis set `user:<id>:sessions` beside the store; (c) store a per-user "session epoch" or password-hash fingerprint in the session and reject sessions whose value no longer matches (Laravel's `AuthenticateSession` middleware uses the password hash this way).

---

## 4. CSRF

### `http.CrossOriginProtection` (stdlib since Go 1.25)

Behaviour, read from [`net/http/csrf.go`](https://cs.opensource.google/go/go/+/refs/tags/go1.27.2:src/net/http/csrf.go) in Go 1.27.2:

1. `GET`, `HEAD`, `OPTIONS` always pass. The doc says "It's important that applications do not perform any state changing actions due to requests with safe methods."
2. `Sec-Fetch-Site: same-origin` or `none` passes. Any other value (`same-site`, `cross-site`) fails unless the `Origin` is added via `AddTrustedOrigin` or the route matches `AddInsecureBypassPattern`.
3. No `Sec-Fetch-Site`: if there's no `Origin` either, the request passes (treated as same-origin or non-browser). If `Origin`'s host equals `Host`, it passes. The source notes this cannot detect HTTP→HTTPS and says "We fail open... Sites can mitigate this with HTTP Strict Transport Security (HSTS)."
4. Otherwise it rejects with 403 (or a custom `SetDenyHandler`).

No tokens, no form instrumentation, no cookie. Design rationale: Filippo Valsorda, [*CSRF*](https://words.filippo.io/csrf/): `Sec-Fetch-Site` has been in all major browsers since 2023 and is only sent to trustworthy (HTTPS/localhost) origins; same-site sibling subdomains are rejected unless trusted; `Origin: null` counts as cross-origin.

Caveats that apply to this template:

- **Reverse proxies** that rewrite `Host` break the `Origin`/`Host` fallback for old browsers. Fix with `AddTrustedOrigin(APP_URL)`.
- **Local dev over plain HTTP on a non-localhost host**: browsers don't send `Sec-Fetch-Site` to non-trustworthy origins, so only the `Origin`/`Host` fallback runs.
- **Logout must be POST.** Any state change on GET bypasses it.

### OWASP's position

The [CSRF Prevention Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/Cross-Site_Request_Forgery_Prevention_Cheat_Sheet.html) calls Fetch Metadata "a lightweight and reliable method to block obvious cross-site requests" for software targeting modern browsers, and states that "a fallback to standard origin verification headers is a mandatory requirement for any Fetch Metadata implementation". `CrossOriginProtection` has that `Origin` fallback. The sheet also says SameSite cookies are defense in depth, not a replacement. The synchronizer token pattern is still "one of the most popular and recommended methods".

### Token-based libraries

| Library | Status | Notes |
|---|---|---|
| [gorilla/csrf](https://github.com/gorilla/csrf) | v1.7.3 (2025-04-14) | Fixed [CVE-2025-24358 / GHSA-rq77-p4h8-4crw](https://github.com/gorilla/csrf/security/advisories/GHSA-rq77-p4h8-4crw): its Referer check never ran because it read `r.URL.Scheme`, which is empty on server requests, so same-site origins could forge requests. Upgrading requires telling it whether the request is plaintext HTTP. |
| [justinas/nosurf](https://github.com/justinas/nosurf) | v1.2.0 (2025-05-13) | Security release for [CVE-2025-46721](https://github.com/justinas/nosurf-cve-2025-46721); adds same-origin checks that were not enforced before. |

Both had origin-check bugs in 2025, both are dependencies, and both need a hidden field in every form.

### Trade-offs

- **Stdlib `CrossOriginProtection` only**: zero dependencies, no template plumbing, matches Go's direction. Relies on browsers from 2023 onward, with the `Origin` fallback for older ones; HTTP→HTTPS gap on old browsers is closed by HSTS.
- **Stdlib + SameSite=Lax session cookie**: the usual pairing; SameSite as defense in depth, as OWASP says.
- **Add a synchronizer token as well** (stored in the Redis session, compared with `subtle.ConstantTimeCompare`): about 30 lines of stdlib code and a hidden field in every form. Covers browsers without `Sec-Fetch-Site`/`Origin`. Only matters if old-browser support is a requirement.

---

## 5. Rate limiting login and two-factor authentication

### Requirements

- RFC 4226 §7.3: OTP validation "needs to detect and stop brute force attacks"; set a throttle parameter T "as low as possible, while still ensuring that usability is not significantly impacted". This applies to TOTP, which is HOTP over time.
- OWASP MFA cheat sheet: "Apply strict attempt limits" for OTPs.
- OWASP Forgot Password: rate-limit reset requests per account to stop inbox flooding.
- Parity target: Fortify login 5/min per `email|ip`; two-factor challenge 5/min per pending login ID (section 1).

### Options (Redis-backed)

| Option | Status | Algorithm | Notes |
|---|---|---|---|
| [go-redis/redis_rate](https://github.com/go-redis/redis_rate) v10 | v10.0.1; last default-branch commit 2023-04; depends on go-redis v9 | GCRA via a Lua script | `Allow(ctx, key, redis_rate.PerMinute(5))` returns `Allowed`, `Remaining`, `RetryAfter`; `Reset(ctx, key)` clears a key after successful login. Small and stable, but effectively finished/unmaintained. |
| [ulule/limiter](https://github.com/ulule/limiter) v3 | v3.11.2 (2023-05) | Fixed window | Redis + memory stores, HTTP middleware. Old. |
| [sethvargo/go-limiter](https://github.com/sethvargo/go-limiter) | v1.2.0 (2026-07) | Token bucket | Redis store is a separate repo, [go-redisstore](https://github.com/sethvargo/go-redisstore), last push 2023-10, redigo. |
| Hand-rolled fixed window | n/a | `INCR key` + `EXPIRE key 60 NX` in a `MULTI` or Lua | About 20 lines on go-redis, the same model as Laravel's `RateLimiter` (counter + decay). `DEL` on success. |

### Points that matter more than the library

- **Key choice**: login keyed by normalised email + IP (Fortify). Per-email alone lets an attacker lock users out; per-IP alone misses distributed guessing. Many apps use both a tight `email|ip` key and a looser per-email key.
- **Two-factor key**: key by the pending user ID, not the IP, so rotating IPs doesn't reset the budget.
- **Recovery codes** share the two-factor challenge limiter.
- **Responses**: return `429` with `Retry-After`; keep the same error message for unknown email and wrong password (enumeration).
- **Fail-closed vs fail-open** when Redis is unavailable is a decision to make.

---

## 6. Password hashing

### OWASP [Password Storage Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/Password_Storage_Cheat_Sheet.html)

- **Argon2id preferred**: "minimum configuration of 19 MiB of memory, an iteration count of 2, and 1 degree of parallelism". Equivalent alternatives: `m=47104 (46 MiB), t=1, p=1`; `m=12288, t=3, p=1`; `m=9216, t=4, p=1`; `m=7168, t=5, p=1`.
- **scrypt** if Argon2id is unavailable: `N=2^17, r=8, p=1` and others.
- **bcrypt** "legacy systems only": work factor "minimum of 10", max input 72 bytes; pre-hashing is "dangerous" (null bytes, password shucking) unless done as `bcrypt(base64(hmac-sha384(password, pepper)))`.
- **PBKDF2** for FIPS-140: HMAC-SHA256 with 600,000 iterations.
- **Pepper**: optional defense in depth, stored outside the DB; can't be rotated without the user's password.
- **Upgrades**: rehash on next successful login; store algorithm and parameters in PHC string format.

### Go options

| Option | Status | Notes |
|---|---|---|
| [`golang.org/x/crypto/argon2`](https://pkg.go.dev/golang.org/x/crypto/argon2) `IDKey` | x/crypto v0.57.0 (2026-09-08) | Raw KDF only: `IDKey(password, salt, time, memory KiB, threads, keyLen)`. The caller handles salt (`crypto/rand`), PHC encoding/parsing, and constant-time comparison (`subtle.ConstantTimeCompare`). About 60 lines. The godoc example uses RFC 9106's 2 GiB parameters; those are far above OWASP's web-login minimum. |
| [alexedwards/argon2id](https://github.com/alexedwards/argon2id) | v1.0.0 (2023-10), pushed 2025-10 | Wraps x/crypto: `CreateHash`, `ComparePasswordAndHash`, `CheckHash` (returns parameters, for rehash detection), PHC string. **`DefaultParams` = 64 MiB, t=1, p=`runtime.NumCPU()`**; p depends on the host CPU, so hashes vary by machine. Pass explicit params. |
| [`golang.org/x/crypto/bcrypt`](https://pkg.go.dev/golang.org/x/crypto/bcrypt) | x/crypto v0.57.0 | Complete API (`GenerateFromPassword`, `CompareHashAndPassword`, `Cost` for rehash detection). Returns `ErrPasswordTooLong` for >72 bytes instead of truncating silently. |
| `crypto/pbkdf2` | stdlib since Go 1.24 | Only relevant if FIPS is a requirement. |

### Trade-offs

- **Argon2id via x/crypto + own PHC codec**: OWASP-preferred, only depends on x/crypto (Go-team maintained). The small codec is security-sensitive (parse untrusted parameter strings; cap memory/time to stop DoS from a tampered hash).
- **Argon2id via alexedwards/argon2id**: removes the codec code; one more small third-party dependency; parameters must be set explicitly.
- **bcrypt**: complete API in x/crypto, no codec needed, matches Laravel's default. OWASP calls it legacy. The 72-byte limit needs a validation rule (max password length in bytes, not runes).
- **Memory cost under load**: Argon2id at 19-46 MiB per hash means N concurrent logins use N times that. Login rate limiting (section 5) and/or a semaphore on hashing bound it.

---

## 7. Two-factor authentication (TOTP)

### Requirements

- **RFC 6238 §5.2**: "We RECOMMEND that at most one time step is allowed as the network delay"; default step 30 s.
- **RFC 6238 §5.2 (replay)**: "The verifier MUST NOT accept the second attempt of the OTP after the successful validation has been issued for the first OTP, which ensures one-time only use of an OTP."
- **RFC 4226 §4**: shared secret "MUST be at least 128 bits. This document RECOMMENDs a shared secret length of 160 bits."
- **RFC 6238 §3 R7**: keys "SHOULD be protected against unauthorized access and usage". Fortify encrypts the secret at rest.
- **RFC 4226 §7.3**: throttle validation attempts (section 5).

### Options

| Option | Status | Notes |
|---|---|---|
| [pquerna/otp](https://github.com/pquerna/otp) | v1.5.0 (2025-05-16), last push 2025-08; depends on `boombuler/barcode` for QR images | `totp.Generate` (default 20-byte = 160-bit secret, SHA1, 6 digits, 30 s), `Key.URL()` gives the `otpauth://` URI, `Key.Image()` the QR. `totp.Validate` uses `Skew: 1`. HOTP comparison uses `subtle.ConstantTimeCompare`. **v1.5.0 fixed a security bug**: secret generation could lose entropy on short random reads ([PR #100](https://github.com/pquerna/otp/pull/100)). Use ≥ v1.5.0. |
| [xlzd/gotp](https://github.com/xlzd/gotp) | v0.1.0 (2022), last push 2022-10 | Unmaintained. |
| Stdlib implementation | n/a | HOTP is `HMAC-SHA1(key, counter)` + dynamic truncation (RFC 4226 §5.3) + mod 10^6: `crypto/hmac`, `crypto/sha1`, `encoding/binary`, `encoding/base32`, `crypto/subtle`. About 50 lines, plus test vectors from RFC 6238 Appendix B. |

### Replay protection is not provided by pquerna/otp

`totp.ValidateCustom` returns only `bool`. It checks counters `t, t+1, t-1, …` and does not say **which time step matched** ([`totp/totp.go`](https://github.com/pquerna/otp/blob/master/totp/totp.go)). To meet the RFC 6238 MUST, the caller must:

- loop over the counters itself using `hotp.ValidateCustom` (or `totp.GenerateCodeCustom`) to learn the matched step, then
- reject if `matchedStep <= lastUsedStep` for that user, and persist `lastUsedStep` (Postgres column or Redis key with TTL ≥ window), atomically with acceptance.

Fortify does the same through `verifyKeyNewer` plus a cache entry. With replay handling written by hand anyway, the library's remaining value is secret generation, the `otpauth://` URI and the QR code.

### QR code

Stdlib has no QR encoder. Options: `boombuler/barcode` v1.1.0 (2025-07, active; pulled in by pquerna/otp already), `yeqown/go-qrcode` v2.2.5 (2025-02), `skip2/go-qrcode` (no release; last push 2024-03). An SVG QR can be inlined in the server-rendered page; always show the base32 secret beside it for manual entry.

### Secret storage

Encrypt the secret at rest in Postgres. A hash is not possible because verification needs the plaintext key. The stdlib has AES-GCM (`crypto/aes` + `crypto/cipher`); key from config, optionally derived per purpose with `crypto/hkdf`. Key rotation is something to plan for.

### Enrolment flow (Fortify parity)

Password confirmation → generate secret (not active) → show QR + secret → user submits a code → only then mark confirmed and issue recovery codes. Disabling also requires password confirmation.

---

## 8. Recovery codes

### What the sources say

- OWASP MFA cheat sheet: provide "a number of single-use recovery codes when they first setup MFA". It does not specify generation or storage.
- Fortify: 8 codes of the form `xxxxxxxxxx-xxxxxxxxxx` (20 alphanumerics, ~119 bits), stored **encrypted** as a JSON array; a used code is replaced with a new one; codes can be regenerated and are re-displayable in settings.

### Storage options

| Option | Pro | Con |
|---|---|---|
| **Encrypted (Fortify)** | Settings page can show the codes again (Livewire starter kit parity) | A DB read plus key compromise reveals all codes |
| **Hashed** (SHA-256 per code; codes have ≥100 bits of entropy, so a slow hash is unnecessary) | A DB leak reveals nothing usable | Codes can only be shown once at generation; "view recovery codes" becomes "regenerate" |
| Hashed with Argon2/bcrypt | Same | Wasted cost for high-entropy codes; N slow hashes per attempt |

### Pitfalls

- **Single use must be atomic**: mark-used and accept in one statement (`UPDATE … SET used_at = now() WHERE user_id = $1 AND code_hash = $2 AND used_at IS NULL RETURNING id`), or two concurrent requests can both succeed.
- **Comparison**: with hashed codes and a lookup by hash, timing leaks nothing useful. With encrypted codes compared in memory, use `subtle.ConstantTimeCompare` and check every code, not stopping at the first match.
- **Normalisation**: strip spaces and hyphens and fold case before comparing, consistently at generation and verification.
- **Generation**: `crypto/rand` only. Go 1.24+ `crypto/rand.Text()` returns ≥128 bits as base32 (26 chars); for a shorter, grouped format, encode `crypto/rand.Read` bytes yourself.
- Using a recovery code logs the user in and should notify them; regenerating codes requires password confirmation.

---

## 9. Signed URLs and tokens for email verification and password reset

### Requirements (OWASP [Forgot Password Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/Forgot_Password_Cheat_Sheet.html))

- Token from a CSPRNG, long enough to resist brute force, single use, expires, invalidated after use.
- Same response for existing and non-existing accounts, and the same response *time* (send mail asynchronously; no early return).
- Don't build the link from the request `Host` header; use a configured app URL (Host header injection).
- `Referrer-Policy: no-referrer` on the reset page.
- After reset: offer to end other sessions, or end them automatically.
- Rate-limit reset requests per account.

### Two mechanisms

| Mechanism | Fits | How (stdlib) | Properties |
|---|---|---|---|
| **Stateless signed URL** (Laravel email verification) | Email verification | `HMAC-SHA256(key, canonical(path + query incl. expires))`, appended as `signature`; verify with `hmac.Equal` and check `expires` | No DB row. **Not single-use** by itself; replay is harmless for verification (idempotent), but not for reset. Bind to the email (Laravel puts `sha1(email)` in the URL) so changing the email invalidates old links. |
| **Stored random token** (Laravel password reset) | Password reset | `crypto/rand.Text()` → email link; store `sha256(token)` + `expires_at` in Postgres; on use, look up by hash, check expiry, delete row in the same transaction as the password update | Single-use, revocable. A fast hash is enough for ≥128-bit tokens; Laravel uses the bcrypt hasher, which is unnecessary at that entropy. |

### Stdlib pieces and pitfalls

- `crypto/hmac` + `crypto/sha256` for signing; **`hmac.Equal`** for comparison (`==` or `bytes.Equal` leak timing).
- **Key separation**: don't reuse one secret for sessions, URL signing and encryption. `crypto/hkdf` (stdlib since Go 1.24) derives per-purpose keys from one app key.
- **Canonicalise before signing**: sign a canonical form (sorted query, signature removed), and include `expires` in the signed data so it can't be extended.
- Reset invalidates outstanding reset tokens and (per OWASP) sessions. A password change should also invalidate remember-me tokens (section 3).
- Laravel throttles reset-link sends to one per 60 s per user.

---

## 10. Stdlib primitives available in Go 1.27

All confirmed with `go doc` against Go 1.27.2.

| Need | Stdlib | Since |
|---|---|---|
| Random tokens | `crypto/rand.Text()` (≥128 bits, base32); `crypto/rand.Read` "never returns an error, and always fills b entirely" | 1.24 |
| Constant-time compare | `crypto/subtle.ConstantTimeCompare` (returns early on length mismatch only); `crypto/hmac.Equal` | long-standing |
| MAC / signing | `crypto/hmac`, `crypto/sha256` | long-standing |
| Key derivation | `crypto/hkdf`, `crypto/pbkdf2` | 1.24 |
| Encryption at rest | `crypto/aes` + `crypto/cipher` (GCM) | long-standing |
| CSRF | `net/http.CrossOriginProtection` | 1.25 |
| Cookies | `net/http.Cookie` incl. `Partitioned`, `SameSite` | long-standing |
| Password hashing | not in stdlib; `golang.org/x/crypto/argon2`, `.../bcrypt` (x/crypto v0.57.0) | n/a |
| TOTP | not in stdlib; buildable from `crypto/hmac`, `crypto/sha1`, `encoding/base32` | n/a |
| QR codes | not in stdlib | n/a |

The Go 1.27 release notes ([go.dev/doc/go1.27](https://go.dev/doc/go1.27)) contain nothing that changes sessions, CSRF or password hashing. Relevant to the HTTP layer: new `Server.MaxHeaderValueCount`; `httptest.NewTestServer` for `testing/synctest`; `crypto/tls` `Config.Rand` deprecated in favour of `testing/cryptotest.SetGlobalRandom`.

---

## 11. Pitfall checklist

| Pitfall | Where | Mitigation |
|---|---|---|
| Session fixation | Login, two-factor completion, logout, password change/reset | Regenerate the session token (`RenewToken`) on every privilege change; never promote the pre-login session |
| Half-authenticated two-factor state treated as logged in | Two-factor challenge | Pending state uses a distinct session key that the authenticated middleware ignores |
| TOTP replay within the window | Two-factor verify | Persist last accepted time step per user; reject `<=`; `Skew ≤ 1` |
| TOTP / recovery-code brute force | Two-factor challenge | Rate limit keyed by pending user ID (Fortify: 5/min) |
| Recovery code double-spend | Recovery code use | Atomic conditional `UPDATE … WHERE used_at IS NULL` |
| Recovery codes / TOTP secrets readable from a DB dump | Storage | Hash codes (if re-display isn't needed) or encrypt; always encrypt TOTP secrets |
| Timing leak in comparisons | Tokens, signatures, codes | `hmac.Equal` / `subtle.ConstantTimeCompare`; never `==` on secrets |
| User enumeration via timing | Login, reset request | On unknown email, still run a password hash against a dummy hash; send mail async; identical messages |
| bcrypt 72-byte limit | Registration, password change | Validate max length in bytes; x/crypto returns `ErrPasswordTooLong` |
| Argon2 memory DoS | Login | Rate limit; bound concurrent hashing; parse stored parameters defensively |
| Session keys readable in Redis | Session store | scs `HashTokenInStore = true` (or equivalent) |
| Cookie flags | Session, remember-me | `Secure`, `HttpOnly`, `SameSite=Lax` (Strict breaks inbound links), `__Host-` prefix |
| State change on GET | Logout, verification link | Logout is POST; the email verification GET is idempotent and signed |
| Host header injection in emailed links | Reset, verification emails | Build URLs from configured `APP_URL` |
| Token leakage via Referer | Reset page | `Referrer-Policy: no-referrer` |
| Stale sessions after password change | Password change/reset | Per-user session index or epoch check; delete remember-me tokens |
| CSRF behind a proxy | `CrossOriginProtection` | `AddTrustedOrigin(APP_URL)` if `Host` is rewritten |
| Old library versions with known CVEs | gorilla/csrf < 1.7.3, nosurf < 1.2.0, pquerna/otp < 1.5.0 | Pin at or above the fixed versions if used |

---

## 12. Open questions for the grilling ticket

1. Sessions: scs + goredisstore, or a hand-rolled store on go-redis with a per-user session index?
2. Remember-me: long-lived session (scs `RememberMe`) or a separate remember token table (Laravel model)?
3. CSRF: `http.CrossOriginProtection` alone (plus SameSite), or also a synchronizer token for pre-2023 browsers?
4. Rate limiting: redis_rate (GCRA, unmaintained but small) or a hand-rolled fixed window on go-redis? Fail-open or fail-closed when Redis is down?
5. Password hashing: Argon2id (x/crypto + own codec, or alexedwards/argon2id) vs bcrypt; which OWASP parameter set?
6. TOTP: pquerna/otp (plus own replay layer) or stdlib implementation; which QR library?
7. Recovery codes: encrypted (re-displayable, Fortify parity) or hashed (shown once)?
8. Where to store per-user last-used TOTP step: Postgres column or Redis key?
9. Encryption key management for TOTP secrets: single app key + HKDF-derived subkeys? Rotation story?
10. Session lifetimes: idle and absolute timeout values; password confirmation window (Laravel: 3 h).

## Sources

- RFC 4226 (HOTP): <https://www.rfc-editor.org/rfc/rfc4226>
- RFC 6238 (TOTP): <https://www.rfc-editor.org/rfc/rfc6238>
- OWASP Password Storage Cheat Sheet: <https://cheatsheetseries.owasp.org/cheatsheets/Password_Storage_Cheat_Sheet.html>
- OWASP Session Management Cheat Sheet: <https://cheatsheetseries.owasp.org/cheatsheets/Session_Management_Cheat_Sheet.html>
- OWASP CSRF Prevention Cheat Sheet: <https://cheatsheetseries.owasp.org/cheatsheets/Cross-Site_Request_Forgery_Prevention_Cheat_Sheet.html>
- OWASP Multifactor Authentication Cheat Sheet: <https://cheatsheetseries.owasp.org/cheatsheets/Multifactor_Authentication_Cheat_Sheet.html>
- OWASP Forgot Password Cheat Sheet: <https://cheatsheetseries.owasp.org/cheatsheets/Forgot_Password_Cheat_Sheet.html>
- Go `net/http.CrossOriginProtection`: <https://pkg.go.dev/net/http#CrossOriginProtection> and Go 1.27.2 `src/net/http/csrf.go`
- Filippo Valsorda, *CSRF*: <https://words.filippo.io/csrf/>
- Go 1.27 release notes: <https://go.dev/doc/go1.27>
- `golang.org/x/crypto` v0.57.0 (`argon2`, `bcrypt`): <https://pkg.go.dev/golang.org/x/crypto>
- alexedwards/scs v2.9.0: <https://github.com/alexedwards/scs>
- gorilla/sessions v1.4.0: <https://github.com/gorilla/sessions>; boj/redistore: <https://github.com/boj/redistore>; rbcervilla/redisstore: <https://github.com/rbcervilla/redisstore>
- redis/go-redis v9.23.0: <https://github.com/redis/go-redis>
- gorilla/csrf advisory GHSA-rq77-p4h8-4crw: <https://github.com/gorilla/csrf/security/advisories/GHSA-rq77-p4h8-4crw>
- justinas/nosurf v1.2.0: <https://github.com/justinas/nosurf/releases/tag/v1.2.0>
- go-redis/redis_rate v10: <https://github.com/go-redis/redis_rate>; ulule/limiter: <https://github.com/ulule/limiter>; sethvargo/go-limiter: <https://github.com/sethvargo/go-limiter>
- alexedwards/argon2id: <https://github.com/alexedwards/argon2id>
- pquerna/otp v1.5.0: <https://github.com/pquerna/otp>
- Laravel Fortify: <https://github.com/laravel/fortify>; Livewire starter kit: <https://github.com/laravel/livewire-starter-kit>; Laravel framework `SessionGuard`, `DatabaseTokenRepository`, `VerifyEmail`: <https://github.com/laravel/framework>

Library versions and activity were read from the GitHub API on 2026-10-09.
