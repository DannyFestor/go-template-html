# Stdlib-built security primitives on HKDF-derived keys

Sessions, TOTP, rate limiting and signed URLs are written on the standard library and go-redis instead of imported, and every key the app uses is derived from one `APP_KEY` with `crypto/hkdf`, one subkey per purpose. Each primitive is small once the template's other decisions are in place: the typed session facade, credential fingerprint and remember tokens are ours regardless, TOTP replay protection needs the matched time step that no library returns, and a fixed-window limiter is a few Valkey commands. Owning them removes dependencies from the most security-sensitive path and lets each be tested exactly, against RFC vectors or a real Valkey. Per-purpose subkeys make separation between encryption, URL signatures, recovery-code digests and credential fingerprints cryptographic rather than a matter of payload format, as it is in Laravel, while keeping one secret to configure and rotate.

## Considered Options

- **alexedwards/scs + goredisstore**: maintained and battle-tested, but it would contribute only the store, encoding and the load-and-save middleware; we take on the `ResponseWriter` wrapper (cookie before headers flush, `Flush`/`Unwrap` for streaming) and its tests instead.
- **pquerna/otp**: `Validate` returns only a bool, so the RFC 6238 replay guard would still need our own validation loop.
- **go-redis/redis_rate**: GCRA is smoother than a fixed window, but the package has had no commits since 2023.
- **`APP_KEY` used directly for every purpose** (Laravel): one key, but an output of one mechanism could be accepted by another.

## Consequences

- The session middleware and rate limiter carry the risk a library would have absorbed; both need thorough tests against a real Valkey, including expiry and window edges.
- Changing the derivation labels or algorithm invalidates every ciphertext, signature, digest and fingerprint; rotation goes through `APP_PREVIOUS_KEYS`, and values move to the new key on their next write.
- The QR encoder (boombuler/barcode) and Argon2id (`x/crypto`) stay dependencies: encoding QR codes and the Argon2 KDF are substantial code, not glue.
