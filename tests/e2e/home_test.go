//go:build e2e

package e2e_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHomeGreetsTheVisitor(t *testing.T) {
	t.Parallel()

	page := openPage(t, "/")

	require.NoError(t, expect.Page(page.Page).ToHaveTitle(appName))
	require.NoError(t, expect.Locator(page.Locator("main h1")).ToHaveText("Let's get started"))
}

func TestHomeLoadsItsAssetsWithoutErrors(t *testing.T) {
	t.Parallel()

	page := openPage(t, "/")

	assert.Empty(t, page.errors.list())
}
