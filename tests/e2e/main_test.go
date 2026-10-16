//go:build e2e

package e2e_test

import (
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/mxschmitt/playwright-go"
)

const browserVariable = "E2E_BROWSER"

var errUnknownBrowser = errors.New("want chromium, firefox or webkit, which `make test-e2e` passes")

//nolint:gochecknoglobals // Launching a browser takes seconds, so every test shares the one TestMain launches.
var (
	browser playwright.Browser
	expect  = playwright.NewPlaywrightAssertions()
)

func TestMain(m *testing.M) {
	os.Exit(runInBrowser(m))
}

func runInBrowser(m *testing.M) int {
	driver, err := playwright.Run()
	if err != nil {
		return reportFailure(fmt.Errorf("start the Playwright driver, which `make e2e-install` installs: %w", err))
	}

	defer func() { _ = driver.Stop() }()

	name := os.Getenv(browserVariable)

	browserType, known := browserTypes(driver)[name]
	if !known {
		return reportFailure(fmt.Errorf("%s=%q: %w", browserVariable, name, errUnknownBrowser))
	}

	browser, err = browserType.Launch()
	if err != nil {
		return reportFailure(fmt.Errorf("launch %s: %w", browserType.Name(), err))
	}

	defer func() { _ = browser.Close() }()

	return m.Run()
}

func browserTypes(driver *playwright.Playwright) map[string]playwright.BrowserType {
	return map[string]playwright.BrowserType{
		"chromium": driver.Chromium,
		"firefox":  driver.Firefox,
		"webkit":   driver.WebKit,
	}
}

func reportFailure(err error) int {
	_, _ = fmt.Fprintln(os.Stderr, err)

	return 1
}
