# Go SSR Template

A clonable starting point for server-rendered Go web apps, giving each new project authentication, account settings, and production infrastructure from day one.

## Language

### Configuration

**Optional module**:
A self-contained capability shipped with the template that a developer can opt into or remove without touching unrelated code.
_Avoid_: Plugin, add-on, extension

### Architecture

**Action**:
A single use case the application performs, such as registering a user or enabling two-factor authentication, holding that use case's business logic.
_Avoid_: Use case, service, interactor, command

### Jobs

**Job**:
A unit of work the application performs in the background, outside the request that asked for it.
_Avoid_: Task, background job

**Periodic job**:
A job the application performs on a fixed schedule.
_Avoid_: Cron job, scheduled task

**Failed job**:
A job that used up its attempts without succeeding.
_Avoid_: Dead job, dead letter

**Worker**:
The process that performs jobs.
_Avoid_: Consumer, job runner

### Testing

**Unit test**:
A test of one piece of behaviour in isolation, with its collaborators replaced by test doubles.
_Avoid_: Small test

**Integration test**:
A test that exercises one adapter against the real external system it wraps, without going through HTTP.
_Avoid_: Database test, repository test

**Feature test**:
A test that exercises the whole application through real HTTP requests, without a browser.
_Avoid_: HTTP test, request test

**E2E test**:
A test that drives a real browser against the running application.
_Avoid_: Browser test, acceptance test

**Architecture test**:
A test that checks the code's structure against the architecture rules, not its behaviour.
_Avoid_: Structure test, fitness function

### Accounts

**Verified user**:
A signed-in user who has proven ownership of their current email address by following the link sent to it.
_Avoid_: Confirmed user, activated user

**Security setting**:
A choice an individual user makes about their own account's protection, such as enabling two-factor authentication.
_Avoid_: Preference

**Two-factor authentication**:
A second sign-in step requiring a time-based one-time code from the user's authenticator app.
_Avoid_: MFA, 2FA (in prose), OTP

**Recovery code**:
A single-use code that substitutes for a two-factor authentication code when the authenticator app is unavailable.
_Avoid_: Backup code

**Password confirmation**:
A re-entry of the current password required before a sensitive action, valid for a limited time.
_Avoid_: Sudo mode, re-authentication

### Realtime

**Realtime module**:
The optional module that pushes public and private channel broadcasts to pages a visitor has open.
_Avoid_: Websocket module, broadcasting

**Public channel**:
A realtime broadcast stream any visitor may subscribe to.

**Private channel**:
A realtime broadcast stream only authorized, signed-in users may subscribe to.
_Avoid_: Protected channel, auth channel
