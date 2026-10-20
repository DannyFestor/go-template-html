//go:build e2e

package e2e_test

import (
	"net/http/httptest"
	"slices"
	"sync"
	"testing"

	"github.com/mxschmitt/playwright-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/DannyFestor/go-template-html/internal/factory"
	"github.com/DannyFestor/go-template-html/internal/testsupport/testapp"
	"github.com/DannyFestor/go-template-html/internal/testsupport/testdb"
	"github.com/DannyFestor/go-template-html/internal/testsupport/testvalkey"
)

const (
	appName          = "E2E Test App"
	consoleErrorType = "error"
	factorySeed      = 1
)

type browserPage struct {
	playwright.Page

	errors *pageErrors
}

type pageErrors struct {
	mutex    sync.Mutex
	messages []string
}

type app struct {
	*testapp.App

	server *httptest.Server
}

func openPage(t *testing.T, path string) browserPage {
	t.Helper()

	return startApp(t).open(t, path)
}

func startApp(t *testing.T) *app {
	t.Helper()

	testApp := testapp.New(t, testapp.Config(appName, testdb.New(t), testvalkey.New(t)), t.Output())

	return &app{App: testApp, server: testapp.Serve(t, testApp.Handler())}
}

func (a *app) open(t *testing.T, path string) browserPage {
	t.Helper()

	page := newPage(t)

	_, err := page.Goto(a.server.URL + path)
	require.NoError(t, err)

	return page
}

func (a *app) factory() *factory.Factory {
	return factory.New(a.Users, a.Passwords, factorySeed)
}

func newPage(t *testing.T) browserPage {
	t.Helper()

	browserContext, err := browser.NewContext()
	require.NoError(t, err)
	t.Cleanup(func() { assert.NoError(t, browserContext.Close()) })

	page, err := browserContext.NewPage()
	require.NoError(t, err)

	return browserPage{Page: page, errors: collectErrors(page)}
}

func collectErrors(page playwright.Page) *pageErrors {
	errors := &pageErrors{mutex: sync.Mutex{}, messages: nil}

	page.OnConsole(func(message playwright.ConsoleMessage) {
		if message.Type() == consoleErrorType {
			errors.add(message.Text())
		}
	})
	page.OnPageError(func(err error) { errors.add(err.Error()) })

	return errors
}

func (e *pageErrors) add(message string) {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	e.messages = append(e.messages, message)
}

func (e *pageErrors) list() []string {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	return slices.Clone(e.messages)
}
