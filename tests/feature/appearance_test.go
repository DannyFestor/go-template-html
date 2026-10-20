//go:build feature

package feature_test

import (
	"net/http"
	"net/url"
	"testing"

	"github.com/PuerkitoBio/goquery"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/DannyFestor/go-template-html/internal/factory"
	"github.com/DannyFestor/go-template-html/internal/route"
)

func startAppAsAda(t *testing.T) *app {
	t.Helper()

	running := startApp(t)
	running.createAda(t)
	require.Equal(t, http.StatusSeeOther, running.login(t, credentials(adaEmail, factory.Password)).status)

	return running
}

func (a *app) chooseTheme(t *testing.T, theme string) response {
	t.Helper()

	return a.post(t, route.SettingsAppearance, url.Values{"theme": {theme}})
}

func TestTheChosenThemeIsRenderedOnTheNextPageLoad(t *testing.T) {
	t.Parallel()

	running := startAppAsAda(t)

	chosen := running.chooseTheme(t, "dark")

	require.Equal(t, http.StatusSeeOther, chosen.status)
	require.Equal(t, route.SettingsAppearance, chosen.header.Get("Location"))
	assert.True(t, running.page(t, route.Dashboard).Find("html").HasClass("dark"))
}

func TestTheAppearancePageOffersEachThemeWithSystemChosenUntilTheVisitorPicksOne(t *testing.T) {
	t.Parallel()

	document := startAppAsAda(t).page(t, route.SettingsAppearance)

	assert.True(t, document.Find("html").HasClass("system"))
	assert.Equal(t, map[string]string{"light": "false", "dark": "false", "system": "true"}, pressedThemes(document))
}

func TestTheAppearancePageShowsTheChosenThemeAsPressed(t *testing.T) {
	t.Parallel()

	running := startAppAsAda(t)
	running.chooseTheme(t, "light")

	document := running.page(t, route.SettingsAppearance)

	assert.True(t, document.Find("html").HasClass("light"))
	assert.Equal(t, map[string]string{"light": "true", "dark": "false", "system": "false"}, pressedThemes(document))
}

func TestTheChosenThemeAppliesToPublicPagesAfterLoggingOut(t *testing.T) {
	t.Parallel()

	running := startAppAsAda(t)
	running.chooseTheme(t, "dark")
	require.Equal(t, http.StatusSeeOther, running.post(t, route.Logout, url.Values{}).status)

	assert.True(t, running.page(t, route.Home).Find("html").HasClass("dark"))
}

func TestAnUnknownThemeReRendersWithTheErrorAndKeepsTheChosenTheme(t *testing.T) {
	t.Parallel()

	running := startAppAsAda(t)
	running.chooseTheme(t, "dark")

	refused := running.chooseTheme(t, "sepia")

	assert.Equal(t, http.StatusUnprocessableEntity, refused.status)
	document := refused.document(t)
	assert.Equal(t, "Choose light, dark or system.", document.Find("#theme-error").Text())
	assert.True(t, document.Find("html").HasClass("dark"))
	assert.True(t, running.page(t, route.Dashboard).Find("html").HasClass("dark"))
}

func TestAGuestRequestingTheAppearanceSettingsIsSentToLogin(t *testing.T) {
	t.Parallel()

	running := startApp(t)

	for name, fetched := range map[string]response{
		"page": running.get(t, route.SettingsAppearance),
		"form": running.chooseTheme(t, "dark"),
	} {
		assert.Equal(t, http.StatusSeeOther, fetched.status, name)
		assert.Equal(t, route.Login, fetched.header.Get("Location"), name)
	}
}

func pressedThemes(document *goquery.Document) map[string]string {
	pressed := map[string]string{}

	document.Find(`main#app-main form[action="/settings/appearance"] button[name="theme"]`).
		Each(func(_ int, button *goquery.Selection) {
			pressed[button.AttrOr("value", "")] = button.AttrOr("aria-pressed", "")
		})

	return pressed
}
