# Go SSR Template

A clonable starting point for server-rendered Go web apps, giving each new project authentication, account settings, and production infrastructure from day one.

## Language

### Configuration

**Feature flag**:
An app-wide switch, set by the developer, that turns an authentication feature on or off for every user.
_Avoid_: Toggle, setting, option

**Optional module**:
A self-contained capability shipped with the template that a developer can opt into or remove without touching unrelated code.
_Avoid_: Plugin, add-on, extension

### Architecture

**Action**:
A single use case the application performs, such as registering a user or enabling two-factor authentication, holding that use case's business logic.
_Avoid_: Use case, service, interactor, command

### Testing

**Feature test**:
A test that exercises the whole application through real HTTP requests, without a browser.
_Avoid_: Integration test, HTTP test

**E2E test**:
A test that drives a real browser against the running application.
_Avoid_: Browser test, acceptance test

**Architecture test**:
A test that checks the code's structure against the architecture rules, not its behaviour.
_Avoid_: Structure test, fitness function

### Accounts

**Security setting**:
A choice an individual user makes about their own account's protection, such as enabling two-factor authentication.
_Avoid_: Preference, feature flag

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

**Public channel**:
A realtime broadcast stream any visitor may subscribe to.

**Private channel**:
A realtime broadcast stream only authorized, signed-in users may subscribe to.
_Avoid_: Protected channel, auth channel
