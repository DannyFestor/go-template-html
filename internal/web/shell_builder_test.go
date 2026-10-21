package web_test

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/DannyFestor/go-template-html/internal/domain"
	"github.com/DannyFestor/go-template-html/internal/view"
	"github.com/DannyFestor/go-template-html/internal/web"
)

const appEntry = "internal/asset/src/app.ts"

var errManifestMissing = errors.New("manifest missing")

type fakeAssets struct {
	scripts     map[string]string
	stylesheets map[string][]string
	err         error
}

func (f fakeAssets) ScriptURL(source string) (string, error) {
	return f.scripts[source], f.err
}

func (f fakeAssets) StylesheetURLs(source string) ([]string, error) {
	return f.stylesheets[source], f.err
}

type fakeFlashes struct {
	message string
}

func (f *fakeFlashes) SetFlash(_ context.Context, message string) {
	f.message = message
}

func (f *fakeFlashes) TakeFlash(context.Context) string {
	message := f.message
	f.message = ""

	return message
}

type fakeCurrentUser struct {
	user     domain.User
	signedIn bool
}

func (f *fakeCurrentUser) CurrentUser(context.Context) (domain.User, bool) {
	return f.user, f.signedIn
}

type fixedTheme domain.Theme

func (f fixedTheme) Read(*http.Request) domain.Theme {
	return domain.Theme(f)
}

func guest() *fakeCurrentUser {
	var nobody domain.User

	return &fakeCurrentUser{user: nobody, signedIn: false}
}

func signedIn(name string) *fakeCurrentUser {
	return &fakeCurrentUser{
		user: domain.User{
			ID:              "019a1f6e-3c4b-7d2e-8f10-1234567890ab",
			Name:            name,
			Email:           "ada@example.com",
			PasswordHash:    "",
			EmailVerifiedAt: time.Time{},
			CreatedAt:       time.Date(2026, 10, 20, 9, 0, 0, 0, time.UTC),
		},
		signedIn: true,
	}
}

func resolvedAssets() fakeAssets {
	return fakeAssets{
		scripts:     map[string]string{appEntry: "/assets/app-nx4752wa.js"},
		stylesheets: map[string][]string{appEntry: {"/assets/app-fimgy6vt.css"}},
		err:         nil,
	}
}

func newShellBuilder(assets fakeAssets, flashes *fakeFlashes, users *fakeCurrentUser) *web.ShellBuilder {
	return web.NewShellBuilder("Acme Shop", assets, flashes, users, fixedTheme(domain.ThemeSystem))
}

func buildShellFor(t *testing.T, users *fakeCurrentUser) view.Shell {
	t.Helper()

	shell, err := newShellBuilder(resolvedAssets(), &fakeFlashes{message: ""}, users).Build(newRequest(t, nil))
	require.NoError(t, err)

	return shell
}

func TestBuildFillsTheShellWithTheAppNameTheAppEntrysAssetsTheFlashAndTheTheme(t *testing.T) {
	t.Parallel()

	flashes := &fakeFlashes{message: "Profile saved."}

	shell, err := newShellBuilder(resolvedAssets(), flashes, guest()).Build(newRequest(t, nil))

	require.NoError(t, err)
	assert.Equal(t, view.Shell{
		AppName:     "Acme Shop",
		Script:      "/assets/app-nx4752wa.js",
		Stylesheets: []string{"/assets/app-fimgy6vt.css"},
		User:        view.ShellUser{ID: "", Name: "", Email: "", Initials: ""},
		Status:      "Profile saved.",
		Theme:       "system",
	}, shell)
	assert.Empty(t, flashes.message, "the flash is still pending")
}

func TestBuildCarriesTheVisitorsChosenTheme(t *testing.T) {
	t.Parallel()

	builder := web.NewShellBuilder("Acme Shop", resolvedAssets(), &fakeFlashes{message: ""}, guest(),
		fixedTheme(domain.ThemeDark))

	shell, err := builder.Build(newRequest(t, nil))

	require.NoError(t, err)
	assert.Equal(t, "dark", shell.Theme)
}

func TestBuildShowsTheSignedInUser(t *testing.T) {
	t.Parallel()

	shell := buildShellFor(t, signedIn("Ada Lovelace"))

	assert.Equal(t, view.ShellUser{
		ID:       "019a1f6e-3c4b-7d2e-8f10-1234567890ab",
		Name:     "Ada Lovelace",
		Email:    "ada@example.com",
		Initials: "AL",
	}, shell.User)
}

func TestBuildTakesTheInitialsOfTheFirstTwoWordsOfTheName(t *testing.T) {
	t.Parallel()

	for name, initials := range map[string]string{
		"Ada Lovelace":          "AL",
		"ada":                   "A",
		"  grace   brewster  ":  "GB",
		"Augusta Ada King-Noel": "AA",
		"Émilie du Châtelet":    "ÉD",
		" Ada Byron  ":          "AB",
	} {
		assert.Equal(t, initials, buildShellFor(t, signedIn(name)).User.Initials, name)
	}
}

func TestBuildFailsWhenTheAssetsCannotBeResolved(t *testing.T) {
	t.Parallel()

	assets := fakeAssets{scripts: nil, stylesheets: nil, err: errManifestMissing}

	_, err := newShellBuilder(assets, &fakeFlashes{message: ""}, guest()).Build(newRequest(t, nil))

	assert.ErrorIs(t, err, errManifestMissing)
}
