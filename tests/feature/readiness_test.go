//go:build feature

package feature_test

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/DannyFestor/go-template-html/internal/route"
	"github.com/DannyFestor/go-template-html/internal/testsupport/testapp"
	"github.com/DannyFestor/go-template-html/internal/testsupport/testdb"
	"github.com/DannyFestor/go-template-html/internal/testsupport/testvalkey"
)

func TestReadyzReportsOKWhileTheDatabaseIsUp(t *testing.T) {
	t.Parallel()

	app := startApp(t)

	got := app.get(t, route.Readyz)

	assert.Equal(t, http.StatusOK, got.status)
	assert.Equal(t, "ok\n", string(got.body))
}

func TestReadyzReportsUnavailableWhileTheDatabaseIsDown(t *testing.T) {
	t.Parallel()

	app := startAppOn(t, testapp.Config(appName, testdb.Unreachable(t), testvalkey.New(t)))

	got := app.get(t, route.Readyz)

	assert.Equal(t, http.StatusServiceUnavailable, got.status)
	assert.Equal(t, "unavailable\n", string(got.body))
}

func TestReadyzReportsUnavailableWhileValkeyIsDown(t *testing.T) {
	t.Parallel()

	app := startAppOn(t, testapp.Config(appName, testdb.New(t), testvalkey.Unreachable(t)))

	got := app.get(t, route.Readyz)

	assert.Equal(t, http.StatusServiceUnavailable, got.status)
	assert.Equal(t, "unavailable\n", string(got.body))
}

func TestReadyzIsLeftOutOfTheInfoAccessLog(t *testing.T) {
	t.Parallel()

	app := startApp(t)

	app.get(t, route.Readyz)

	assert.NotContains(t, app.logs.String(), route.Readyz)
}
