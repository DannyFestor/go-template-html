package testdb

import (
	"context"
	"crypto/rand"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

//nolint:gochecknoglobals // Every test in the process clones the one template, built at most once.
var preparedTemplate = sync.OnceValues(func() (template, error) {
	return prepareTemplate(context.Background())
})

func New(t *testing.T) Database {
	t.Helper()

	source, err := preparedTemplate()
	require.NoError(t, err, "prepare the template database")

	name := source.settings.name + "_t_" + strings.ToLower(rand.Text())

	err = withAdmin(t.Context(), source.settings, func(a admin) error {
		return a.create(t.Context(), name, source.name)
	})
	require.NoError(t, err, "clone the template database")

	t.Cleanup(func() { dropUnlessFailed(t, source.settings, name) })

	return source.settings.database(name)
}

func dropUnlessFailed(t *testing.T, s settings, name string) {
	t.Helper()

	if t.Failed() {
		t.Logf("kept the database %s of the failed test, `make test-db-clean` drops it", name)

		return
	}

	ctx := context.WithoutCancel(t.Context())

	err := withAdmin(ctx, s, func(a admin) error { return a.drop(ctx, name) })
	require.NoError(t, err, "drop the test database")
}
