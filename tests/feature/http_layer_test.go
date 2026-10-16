//go:build feature

package feature_test

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/DannyFestor/go-template-html/internal/testsupport/testapp"
)

func TestUnknownPathRendersTheNotFoundPage(t *testing.T) {
	t.Parallel()

	fetched := startApp(t).get(t, "/no-such-page")

	assert.Equal(t, http.StatusNotFound, fetched.status)
	document := fetched.document(t)
	assert.Equal(t, appName, document.Find("head > title").Text())
	assert.Equal(t, "Not Found", document.Find("body > main#public-main h1").Text())
}

func TestWrongMethodRendersTheMethodNotAllowedPage(t *testing.T) {
	t.Parallel()

	fetched := startApp(t).send(t, outgoing{method: http.MethodDelete, path: "/", headers: nil, body: nil})

	assert.Equal(t, http.StatusMethodNotAllowed, fetched.status)
	assert.Contains(t, fetched.header.Get("Allow"), http.MethodGet)
	assert.Equal(t, "Method Not Allowed", fetched.document(t).Find("main#public-main h1").Text())
}

func TestUnexpectedErrorRendersAReportedServerError(t *testing.T) {
	t.Parallel()

	running := startApp(t)

	fetched := running.get(t, probeFailure)

	assert.Equal(t, http.StatusInternalServerError, fetched.status)
	document := fetched.document(t)
	assert.Equal(t, "Internal Server Error", document.Find("main#public-main h1").Text())
	assert.Zero(t, document.Find("pre").Length(), "error details outside local")

	reported := logLine(t, running.logs.String(), "unexpected error")
	assert.Equal(t, "probe failed", reported["error"])
	assert.Equal(t, fetched.header.Get("X-Request-ID"), reported["request_id"])
}

func TestHTMXRequestReceivesTheErrorPageAsAWholeBodySwap(t *testing.T) {
	t.Parallel()

	fetched := startApp(t).send(t, outgoing{
		method:  http.MethodGet,
		path:    "/no-such-page",
		headers: map[string]string{"HX-Request": "true", "HX-Target": "main#public-main"},
		body:    nil,
	})

	assert.Equal(t, http.StatusNotFound, fetched.status)
	assert.Equal(t, "body", fetched.header.Get("Hx-Retarget"))
	assert.Equal(t, "body", fetched.header.Get("Hx-Reselect"))
	assert.Equal(t, "outerSync", fetched.header.Get("Hx-Reswap"))
	assert.Equal(t, 1, fetched.document(t).Find("body > main#public-main").Length())
}

func TestRedirectSendsSeeOtherToPlainRequests(t *testing.T) {
	t.Parallel()

	fetched := startApp(t).send(t, outgoing{method: http.MethodPost, path: probeRedirect, headers: nil, body: nil})

	assert.Equal(t, http.StatusSeeOther, fetched.status)
	assert.Equal(t, "/", fetched.header.Get("Location"))
}

func TestRedirectSendsHXRedirectToPartialHTMXRequests(t *testing.T) {
	t.Parallel()

	fetched := startApp(t).send(t, outgoing{
		method:  http.MethodPost,
		path:    probeRedirect,
		headers: map[string]string{"HX-Request": "true", "HX-Request-Type": "partial"},
		body:    nil,
	})

	assert.Equal(t, http.StatusOK, fetched.status)
	assert.Equal(t, "/", fetched.header.Get("Hx-Redirect"))
	assert.Empty(t, fetched.header.Get("Location"))
}

func TestBodyWithinTheLimitIsAccepted(t *testing.T) {
	t.Parallel()

	fetched := startApp(t).send(t, formOfSize(testapp.MaxBodyBytes))

	assert.Equal(t, http.StatusNoContent, fetched.status)
}

func TestBodyOverTheLimitIsRejected(t *testing.T) {
	t.Parallel()

	fetched := startApp(t).send(t, formOfSize(testapp.MaxBodyBytes+1))

	assert.Equal(t, http.StatusRequestEntityTooLarge, fetched.status)
	assert.True(t, fetched.closesConnection, "the connection stays open after the 413")
	assert.Equal(t, "Request Entity Too Large", fetched.document(t).Find("main#public-main h1").Text())
}

func TestAccessLogCarriesTheRequestIDWithoutTheQueryString(t *testing.T) {
	t.Parallel()

	running := startApp(t)

	fetched := running.get(t, "/?token=hunter2")

	logged := logLine(t, running.logs.String(), "request")
	assert.Equal(t, fetched.header.Get("X-Request-ID"), logged["request_id"])
	assert.Equal(t, http.MethodGet, logged["method"])
	assert.Equal(t, "/", logged["path"])
	assert.InDelta(t, http.StatusOK, logged["status"], 0)
	assert.NotContains(t, running.logs.String(), "hunter2")
}

func formOfSize(size int) outgoing {
	field := "field="

	return outgoing{
		method:  http.MethodPost,
		path:    probeForm,
		headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		body:    strings.NewReader(field + strings.Repeat("a", size-len(field))),
	}
}

func logLine(t *testing.T, logs, message string) map[string]any {
	t.Helper()

	for line := range strings.Lines(logs) {
		var decoded map[string]any

		require.NoError(t, json.Unmarshal([]byte(line), &decoded))

		if decoded["msg"] == message {
			return decoded
		}
	}

	require.Failf(t, "no log line", "no %q line in:\n%s", message, logs)

	return nil
}
