//go:build e2e

package e2e_test

import (
	"testing"

	"github.com/mxschmitt/playwright-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/DannyFestor/go-template-html/internal/route"
)

const (
	reloadMarker   = "window.e2eNotReloaded"
	appearanceForm = `form[action="` + route.SettingsAppearance + `"]`
)

func TestAUserSwitchesThemesWithoutAFullReload(t *testing.T) {
	t.Parallel()

	running := startApp(t)
	page := running.openSignedIn(t, running.createAda(t), route.SettingsAppearance)
	_, err := page.Evaluate(reloadMarker + " = true")
	require.NoError(t, err)

	require.NoError(t, page.Locator(themeButton("dark")).Click())

	require.NoError(t, expect.Locator(page.Locator("html")).ToContainClass("dark"))
	require.NoError(t, expect.Locator(page.Locator(appearanceForm)).ToHaveAttribute("data-theme", "dark"))
	require.NoError(t, expect.Locator(page.Locator(themeButton("dark"))).ToHaveAttribute("aria-pressed", "true"))
	notReloaded, err := page.Evaluate(reloadMarker)
	require.NoError(t, err)
	assert.Equal(t, true, notReloaded)

	_, err = page.Reload()
	require.NoError(t, err)
	require.NoError(t, expect.Locator(page.Locator("html")).ToContainClass("dark"))
	assert.Empty(t, page.errors.list())
}

func TestTheSystemThemeFollowsTheBrowsersColorScheme(t *testing.T) {
	t.Parallel()

	page := openPage(t, route.Home)

	assert.NotEqual(t,
		page.bodyBackgroundIn(t, playwright.ColorSchemeLight),
		page.bodyBackgroundIn(t, playwright.ColorSchemeDark),
	)
}

func (p browserPage) bodyBackgroundIn(t *testing.T, scheme *playwright.ColorScheme) any {
	t.Helper()

	require.NoError(t, p.EmulateMedia(playwright.PageEmulateMediaOptions{
		ColorScheme:   scheme,
		Contrast:      nil,
		ForcedColors:  nil,
		Media:         nil,
		ReducedMotion: nil,
	}))

	background, err := p.Evaluate("getComputedStyle(document.body).backgroundColor")
	require.NoError(t, err)

	return background
}

func themeButton(theme string) string {
	return appearanceForm + ` button[name="theme"][value="` + theme + `"]`
}
