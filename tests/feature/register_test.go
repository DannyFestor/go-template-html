//go:build feature

package feature_test

import (
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/DannyFestor/go-template-html/internal/bootstrap"
	"github.com/DannyFestor/go-template-html/internal/route"
	"github.com/DannyFestor/go-template-html/internal/testsupport"
	"github.com/DannyFestor/go-template-html/internal/testsupport/testapp"
)

func registration(email, password string) url.Values {
	return url.Values{
		"name":                  {"Ada Lovelace"},
		"email":                 {email},
		"password":              {password},
		"password_confirmation": {password},
	}
}

func (a *app) register(t *testing.T, form url.Values) response {
	t.Helper()

	return a.send(t, outgoing{
		method:  http.MethodPost,
		path:    route.Register,
		headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		body:    strings.NewReader(form.Encode()),
	})
}

func TestTheRegisterPageShowsTheRegistrationForm(t *testing.T) {
	t.Parallel()

	document := startApp(t).page(t, route.Register)

	form := document.Find(`main#public-main form[method="post"][action="/register"]`)
	require.Equal(t, 1, form.Length())
	assert.Equal(t, "false", form.AttrOr("hx-boost", ""))

	for _, name := range []string{"name", "email", "password", "password_confirmation"} {
		assert.Equal(t, 1, form.Find(`input[name="`+name+`"]`).Length(), name)
	}

	assert.Equal(t, 1, document.Find(`main#public-main a[href="/login"]`).Length())
}

func TestAVisitorRegistersAndLandsOnTheDashboard(t *testing.T) {
	t.Parallel()

	running := startApp(t)

	registered := running.register(t, registration("ada@example.com", "correct horse"))

	require.Equal(t, http.StatusSeeOther, registered.status)
	require.Equal(t, route.Dashboard, registered.header.Get("Location"))

	dashboard := running.page(t, route.Dashboard)
	assert.Equal(t, "Dashboard", dashboard.Find("body main#app-main h1").Text())
	assert.Equal(t, "Ada Lovelace", dashboard.Find("#user-menu-name").Text())
}

func TestARegistrationWithATakenEmailReRendersWithTheError(t *testing.T) {
	t.Parallel()

	running := startApp(t)
	testsupport.Must(running.factory().User().WithEmail("ada@example.com").Create(t.Context()))(t)

	refused := running.register(t, registration("ADA@example.com", "correct horse"))

	assert.Equal(t, http.StatusUnprocessableEntity, refused.status)
	document := refused.document(t)
	assert.Equal(t, "This email address is already registered.", document.Find("#email-error").Text())
	assert.Equal(t, "Ada Lovelace", document.Find(`input[name="name"]`).AttrOr("value", ""))
	assert.Equal(t, "ADA@example.com", document.Find(`input[name="email"]`).AttrOr("value", ""))
	assert.Empty(t, document.Find(`input[name="password"]`).AttrOr("value", ""))
	assertStillAGuest(t, running)
}

func TestARegistrationWithAPasswordUnderEightCharactersReRendersWithTheError(t *testing.T) {
	t.Parallel()

	running := startApp(t)

	refused := running.register(t, registration("ada@example.com", "1234567"))

	assert.Equal(t, http.StatusUnprocessableEntity, refused.status)
	assert.Equal(t, "The password must be at least 8 characters.",
		refused.document(t).Find("#password-error").Text())
	assertStillAGuest(t, running)
}

func TestRegistrationIsRateLimitedPerClientIP(t *testing.T) {
	t.Parallel()

	running := startApp(t)
	invalid := registration("ada@example.com", "short")

	for range testapp.RateLimitAttempts {
		require.Equal(t, http.StatusUnprocessableEntity, running.register(t, invalid).status)
	}

	refused := running.register(t, invalid)

	assert.Equal(t, http.StatusTooManyRequests, refused.status)
	assert.NotEmpty(t, refused.header.Get("Retry-After"))
	assert.Equal(t, "Too many attempts. Please try again later.",
		refused.document(t).Find(`form [role="alert"]`).Text())
}

func TestASignedInUserRequestingTheRegisterPageIsSentToTheDashboard(t *testing.T) {
	t.Parallel()

	running := startApp(t)
	require.Equal(t, http.StatusSeeOther, running.register(t, registration("ada@example.com", "correct horse")).status)

	for name, fetched := range map[string]response{
		"page": running.get(t, route.Register),
		"form": running.register(t, registration("grace@example.com", "correct horse")),
	} {
		assert.Equal(t, http.StatusSeeOther, fetched.status, name)
		assert.Equal(t, route.Dashboard, fetched.header.Get("Location"), name)
	}

	assert.Equal(t, "Ada Lovelace", running.page(t, route.Dashboard).Find("#user-menu-name").Text())
}

func TestAGuestRequestingTheDashboardIsSentToLogin(t *testing.T) {
	t.Parallel()

	assertStillAGuest(t, startApp(t))
}

func TestSeedingRegistersTheTestUser(t *testing.T) {
	t.Parallel()

	config := liveConfig(t)
	running := startAppOn(t, config)

	require.NoError(
		t,
		bootstrap.Seed(t.Context(), config, bootstrap.NewLogger(running.logs, config.App.Env, config.Log)),
	)

	refused := running.register(t, registration("test@example.com", "correct horse"))

	assert.Equal(t, http.StatusUnprocessableEntity, refused.status)
	assert.Equal(t, "This email address is already registered.", refused.document(t).Find("#email-error").Text())
}

func assertStillAGuest(t *testing.T, running *app) {
	t.Helper()

	dashboard := running.get(t, route.Dashboard)

	assert.Equal(t, http.StatusSeeOther, dashboard.status)
	assert.Equal(t, route.Login, dashboard.header.Get("Location"))
}
