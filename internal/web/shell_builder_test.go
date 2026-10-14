package web_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

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

func TestBuildFillsTheShellWithTheAppNameAndTheAppEntrysAssets(t *testing.T) {
	t.Parallel()

	assets := fakeAssets{
		scripts:     map[string]string{appEntry: "/assets/app-nx4752wa.js"},
		stylesheets: map[string][]string{appEntry: {"/assets/app-fimgy6vt.css"}},
		err:         nil,
	}

	shell, err := web.NewShellBuilder("Acme Shop", assets).Build()

	require.NoError(t, err)
	assert.Equal(t, view.Shell{
		AppName:     "Acme Shop",
		Script:      "/assets/app-nx4752wa.js",
		Stylesheets: []string{"/assets/app-fimgy6vt.css"},
		User:        view.ShellUser{Name: "", Email: "", Initials: ""},
	}, shell)
}

func TestBuildFailsWhenTheAssetsCannotBeResolved(t *testing.T) {
	t.Parallel()

	assets := fakeAssets{scripts: nil, stylesheets: nil, err: errManifestMissing}

	_, err := web.NewShellBuilder("Acme Shop", assets).Build()

	assert.ErrorIs(t, err, errManifestMissing)
}
