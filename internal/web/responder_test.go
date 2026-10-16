package web_test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/DannyFestor/go-template-html/internal/view/layout"
	"github.com/DannyFestor/go-template-html/internal/web"
)

type text string

func (t text) Render(_ context.Context, w io.Writer) error {
	_, err := io.WriteString(w, string(t))

	return err //nolint:wrapcheck // a test double, whose callers wrap its error
}

type spyReporter struct {
	reported []error
}

func (s *spyReporter) Report(_ context.Context, err error) {
	s.reported = append(s.reported, err)
}

func newResponder(reporter *spyReporter, showErrorDetails bool) *web.Responder {
	return newFlashingResponder(reporter, showErrorDetails, &fakeFlashes{message: ""})
}

func newFlashingResponder(reporter *spyReporter, showErrorDetails bool, flashes *fakeFlashes) *web.Responder {
	assets := fakeAssets{
		scripts:     map[string]string{appEntry: "/assets/app.js"},
		stylesheets: map[string][]string{appEntry: nil},
		err:         nil,
	}

	return web.NewResponder(web.NewShellBuilder("Acme Shop", assets, flashes), flashes, reporter, showErrorDetails)
}

func newRequest(t *testing.T, headers map[string]string) *http.Request {
	t.Helper()

	request := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/", nil)
	for name, value := range headers {
		request.Header.Set(name, value)
	}

	return request
}

func renderPage(t *testing.T, headers map[string]string, page web.Page) *httptest.ResponseRecorder {
	t.Helper()

	recorder := httptest.NewRecorder()

	err := newResponder(&spyReporter{reported: nil}, false).Page(recorder, newRequest(t, headers), http.StatusOK, page)
	require.NoError(t, err)

	return recorder
}

func appPageWithFragment() web.Page {
	return web.Page{Layout: layout.KindApp, Body: text("<p>body</p>"), Fragment: text("<p>fragment</p>")}
}

func appPageWithoutFragment() web.Page {
	return web.Page{Layout: layout.KindApp, Body: text("<p>body</p>"), Fragment: nil}
}

func TestPageRendersTheBodyInsideItsLayout(t *testing.T) {
	t.Parallel()

	recorder := renderPage(t, nil, appPageWithFragment())

	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Equal(t, "text/html; charset=utf-8", recorder.Header().Get("Content-Type"))
	assert.Contains(t, recorder.Body.String(), `<main id="app-main" class="p-6 lg:p-10"><p>body</p></main>`)
	assert.NotContains(t, recorder.Body.String(), "fragment")
}

func TestPageRendersOnlyTheFragmentForPartialRequests(t *testing.T) {
	t.Parallel()

	recorder := renderPage(t, htmxPartial(), appPageWithFragment())

	assert.Equal(t, "text/html; charset=utf-8", recorder.Header().Get("Content-Type"))
	assert.Equal(t, "<p>fragment</p>", recorder.Body.String())
	assert.Equal(t, noSwapOverride(), swapHeaders(recorder.Header()))
}

func TestPageFallsBackToTheFullPageForPartialRequestsWithoutAFragment(t *testing.T) {
	t.Parallel()

	recorder := renderPage(t, htmxPartial(), appPageWithoutFragment())

	assert.Contains(t, recorder.Body.String(), `<main id="app-main" class="p-6 lg:p-10"><p>body</p></main>`)
}

func TestPageSwapsTheWholeBodyWhenABoostedRequestCrossesLayouts(t *testing.T) {
	t.Parallel()

	recorder := renderPage(t, boostedInto("main#public-main"), appPageWithFragment())

	assert.Equal(t, wholeBodySwap(), swapHeaders(recorder.Header()))
	assert.Contains(t, recorder.Body.String(), `<main id="app-main"`)
}

func TestPageKeepsTheTargetWhenABoostedRequestStaysInItsLayout(t *testing.T) {
	t.Parallel()

	recorder := renderPage(t, boostedInto("main#app-main"), appPageWithFragment())

	assert.Equal(t, noSwapOverride(), swapHeaders(recorder.Header()))
}

func swapHeaders(header http.Header) []string {
	return []string{header.Get("Hx-Retarget"), header.Get("Hx-Reselect"), header.Get("Hx-Reswap")}
}

func wholeBodySwap() []string {
	return []string{"body", "body", "outerSync"}
}

func noSwapOverride() []string {
	return []string{"", "", ""}
}

func TestRateLimitedReRendersThePageWithRetryAfterForTooManyRequests(t *testing.T) {
	t.Parallel()

	recorder := httptest.NewRecorder()

	err := newResponder(&spyReporter{reported: nil}, false).
		RateLimited(recorder, newRequest(t, nil), web.ErrTooManyRequests(30*time.Second), appPageWithFragment())

	require.NoError(t, err)
	assert.Equal(t, http.StatusTooManyRequests, recorder.Code)
	assert.Equal(t, "30", recorder.Header().Get("Retry-After"))
	assert.Contains(t, recorder.Body.String(), "<p>body</p>")
}

func TestRateLimitedHandsBackAFailureToCountForItsErrorPage(t *testing.T) {
	t.Parallel()

	recorder := httptest.NewRecorder()
	unavailable := web.ErrServiceUnavailable(errDatabaseDown)

	err := newResponder(&spyReporter{reported: nil}, false).
		RateLimited(recorder, newRequest(t, nil), unavailable, appPageWithFragment())

	assert.Same(t, unavailable, err)
	assert.Empty(t, recorder.Body.String())
}

func TestRedirectSendsSeeOther(t *testing.T) {
	t.Parallel()

	recorder := httptest.NewRecorder()

	newResponder(&spyReporter{reported: nil}, false).Redirect(recorder, newRequest(t, nil), "/dashboard")

	assert.Equal(t, http.StatusSeeOther, recorder.Code)
	assert.Equal(t, "/dashboard", recorder.Header().Get("Location"))
}

func TestRedirectAsksHTMXToNavigateForPartialRequests(t *testing.T) {
	t.Parallel()

	recorder := httptest.NewRecorder()

	newResponder(&spyReporter{reported: nil}, false).Redirect(recorder, newRequest(t, htmxPartial()), "/dashboard")

	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Equal(t, "/dashboard", recorder.Header().Get("Hx-Redirect"))
	assert.Empty(t, recorder.Header().Get("Location"))
}

func TestRedirectWithStatusFlashesTheMessageForTheNextPage(t *testing.T) {
	t.Parallel()

	recorder := httptest.NewRecorder()
	flashes := &fakeFlashes{message: ""}

	newFlashingResponder(&spyReporter{reported: nil}, false, flashes).
		RedirectWithStatus(recorder, newRequest(t, nil), "/dashboard", "Profile saved.")

	assert.Equal(t, http.StatusSeeOther, recorder.Code)
	assert.Equal(t, "/dashboard", recorder.Header().Get("Location"))
	assert.Equal(t, "Profile saved.", flashes.message)
}

func TestPageShowsTheFlashAsAStatusInsideMainOnce(t *testing.T) {
	t.Parallel()

	recorder := httptest.NewRecorder()
	flashes := &fakeFlashes{message: "Profile saved."}

	err := newFlashingResponder(&spyReporter{reported: nil}, false, flashes).
		Page(recorder, newRequest(t, nil), http.StatusOK, appPageWithFragment())

	require.NoError(t, err)
	assert.Contains(t, recorder.Body.String(), `<main id="app-main" class="p-6 lg:p-10"><div role="status"`)
	assert.Contains(t, recorder.Body.String(), "Profile saved.")
	assert.Empty(t, flashes.message)
}

func TestPartialPageLeavesTheFlashForTheNextFullPage(t *testing.T) {
	t.Parallel()

	recorder := httptest.NewRecorder()
	flashes := &fakeFlashes{message: "Profile saved."}

	err := newFlashingResponder(&spyReporter{reported: nil}, false, flashes).
		Page(recorder, newRequest(t, htmxPartial()), http.StatusOK, appPageWithFragment())

	require.NoError(t, err)
	assert.NotContains(t, recorder.Body.String(), "Profile saved.")
	assert.Equal(t, "Profile saved.", flashes.message)
}

func boostedInto(target string) map[string]string {
	return map[string]string{"HX-Request": "true", "HX-Request-Type": "full", "HX-Target": target}
}

func htmxPartial() map[string]string {
	return map[string]string{"HX-Request": "true", "HX-Request-Type": "partial", "HX-Target": "div#results"}
}
