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

### Mail

**Mail**:
A message the application sends to an email address, rendered into an HTML and a plain-text part when it is queued and delivered by the worker.
_Avoid_: Email (for the message), notification

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

**Tooling test**:
A test that exercises the repo's developer tooling, such as its git hooks, against a throwaway git repository or a fake of the service it calls instead of the application.
_Avoid_: Script test, hook test

### Accounts

**Verified user**:
A signed-in user who has proven ownership of their current email address by following the link sent to it.
_Avoid_: Confirmed user, activated user

**Verification link**:
A signed link, valid for a limited time, that is mailed to a user's email address and makes them a verified user when they follow it while signed in.
_Avoid_: Activation link, confirmation link

**Remember token**:
A long-lived credential that one device keeps so the user is signed back in after their session ends.
_Avoid_: Remember-me cookie, persistent login, auth token

**Password reset link**:
A single-use link, valid for a limited time, mailed to a user who forgot their password so they can choose a new one without signing in.
_Avoid_: Reset token (in prose), recovery link

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

### Appearance

**Theme**:
The light, dark or system colour scheme a visitor picks on the appearance settings page, kept per browser in a cookie, not on the account. The system theme follows the browser's own preference.
_Avoid_: Colour mode, dark mode (for the setting), preference

### Realtime

**Realtime module**:
The optional module that pushes public and private channel broadcasts to pages a visitor has open.
_Avoid_: Websocket module, broadcasting

**Public channel**:
A realtime broadcast stream any visitor may subscribe to.

**Private channel**:
A realtime broadcast stream only authorized, signed-in users may subscribe to.
_Avoid_: Protected channel, auth channel

**Announcement**:
A message broadcast on the public `announcements` channel, shown as a banner on every open page.
_Avoid_: Notification, alert

**User notice**:
A message broadcast on one user's private channel, shown as a toast on every page that user has open.
_Avoid_: Notification, alert
