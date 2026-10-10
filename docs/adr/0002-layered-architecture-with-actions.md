# Layered architecture with actions and consumer-declared interfaces

The application is sliced by layer (domain, action, request, handler, job, adapter), each split into the same areas, because authentication and settings are two views onto one user and would not stand alone as feature slices. Business logic lives in actions, one use case each, behind thin handlers; every consumer declares its own small interfaces and every dependency with a plausible substitute (external systems, and libraries implementing a replaceable technique such as password hashing, TOTP or session management) sits behind an adapter, so the database, session store, mailer and queue can be swapped and every layer can be unit-tested with generated mocks. Libraries that are the medium the code is written in (templ for views, test-only tooling such as testify and mockgen) are imported directly, confined by architecture tests to the package that needs them ([ADR-0003](0003-templ-for-server-rendered-views.md)).

## Considered Options

- **Feature slices**: auth and settings share the user and its credentials, so slices would be thin shells around a shared core that does all the work.
- **One package per action**: compiler-enforced isolation, but about thirty tiny packages; one package per area with single-method interfaces composed per action avoids name collisions just as well and is closer to idiomatic Go.
- **Transactions inside adapter methods** (one method per multi-write operation): moves the order of business steps into the adapter. Instead, actions declare a `transactor` and the Postgres adapter carries the transaction in `ctx`; the cost is that passing the outer `ctx` inside the callback silently escapes the transaction.
- **Domain event bus**: looser coupling, but what happens after an action is no longer readable from its code; actions call follow-up actions directly.
- **DI container or code-generated wiring**: replaces about a hundred readable lines; wiring is by hand in one composition root.
- **`go` statements in actions** for slow work: cancelled request contexts, unrecovered panics, lost work on shutdown and harder tests; asynchrony is added by adapters and decorators, and mail goes through the job queue.

## Consequences

- An optional module cannot be a single folder: it is its own area in every layer it touches plus one wiring file, reached from core only through consumer-declared interfaces with a no-op adapter.
- Switching databases means a new adapter plus new migrations and queries; actions, requests and handlers stay untouched.
- The import rules and banned patterns are enforced mechanically, not by convention: go-arch-lint, depguard and architecture tests ([ADR-0007](0007-architecture-enforced-by-go-arch-lint-depguard-and-ast-tests.md)).
