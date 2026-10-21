//go:build e2e

package e2e_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/DannyFestor/go-template-html/internal/route"
)

const headerLoginLink = `header a[href="` + route.Login + `"]`

type failureRecorder struct {
	failures []string
}

func TestANavigatingClickReturnsOnlyOnceThePageIsAtTheDestination(t *testing.T) {
	t.Parallel()

	running := startApp(t)
	page := running.open(t, route.Home)

	page.clickTo(t, headerLoginLink, route.Login)

	assert.Equal(t, running.server.URL+route.Login, page.URL())
}

func TestANavigatingClickFailsNamingBothURLsWhenItLandsElsewhere(t *testing.T) {
	t.Parallel()

	running := startApp(t)
	page := running.open(t, route.Home)
	recorder := &failureRecorder{failures: nil}

	page.clickTo(recorder, headerLoginLink, route.Register)

	require.Len(t, recorder.failures, 1)
	assert.Contains(t, recorder.failures[0], running.server.URL+route.Register)
	assert.Contains(t, recorder.failures[0], running.server.URL+route.Login)
}

func (r *failureRecorder) Errorf(format string, args ...any) {
	r.failures = append(r.failures, fmt.Sprintf(format, args...))
}

func (*failureRecorder) FailNow() {}

func (*failureRecorder) Helper() {}
