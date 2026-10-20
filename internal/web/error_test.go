package web_test

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/DannyFestor/go-template-html/internal/web"
)

var errDatabaseDown = errors.New("database down")

type errorResponse struct {
	recorder *httptest.ResponseRecorder
	reporter *spyReporter
}

func respondWithError(t *testing.T, headers map[string]string, showErrorDetails bool, err error) errorResponse {
	t.Helper()

	recorder := httptest.NewRecorder()
	reporter := &spyReporter{reported: nil}

	newResponder(reporter, showErrorDetails).Error(recorder, newRequest(t, headers), err)

	return errorResponse{recorder: recorder, reporter: reporter}
}

func (r errorResponse) main(t *testing.T) *goquery.Selection {
	t.Helper()

	document, err := goquery.NewDocumentFromReader(bytes.NewReader(r.recorder.Body.Bytes()))
	require.NoError(t, err)

	return document.Find("body > main#public-main")
}

func TestErrorRendersAStatusErrorsPageWithoutReportingIt(t *testing.T) {
	t.Parallel()

	response := respondWithError(t, nil, true, web.ErrNotFound)

	assert.Equal(t, http.StatusNotFound, response.recorder.Code)
	assert.Equal(t, "text/html; charset=utf-8", response.recorder.Header().Get("Content-Type"))
	assert.Equal(t, "Not Found", response.main(t).Find("h1").Text())
	assert.Zero(t, response.main(t).Find("pre").Length())
	assert.Empty(t, response.reporter.reported)
}

func TestErrorExplainsABadRequestWithoutReportingIt(t *testing.T) {
	t.Parallel()

	response := respondWithError(t, nil, false, web.ErrBadRequest)

	assert.Equal(t, http.StatusBadRequest, response.recorder.Code)
	assert.Equal(t, "Bad Request", response.main(t).Find("h1").Text())
	assert.Equal(t, "This request is malformed.", response.main(t).Find("h1 + p").Text())
	assert.Empty(t, response.reporter.reported)
}

func TestErrorReportsAnUnexpectedErrorAsAServerError(t *testing.T) {
	t.Parallel()

	response := respondWithError(t, nil, false, errDatabaseDown)

	assert.Equal(t, http.StatusInternalServerError, response.recorder.Code)
	assert.Equal(t, "Internal Server Error", response.main(t).Find("h1").Text())
	assert.Zero(t, response.main(t).Find("pre").Length())
	assert.Equal(t, []error{errDatabaseDown}, response.reporter.reported)
}

func TestErrorShowsTheErrorAndItsStackWhenDetailsAreShown(t *testing.T) {
	t.Parallel()

	response := respondWithError(t, nil, true, errDatabaseDown)

	detail := response.main(t).Find("pre").Text()
	assert.Contains(t, detail, "database down")
	assert.Contains(t, detail, "goroutine")
}

func TestErrorSwapsTheWholeBodyForHTMXRequests(t *testing.T) {
	t.Parallel()

	response := respondWithError(t, boostedInto("main#public-main"), false, web.ErrForbidden)

	assert.Equal(t, http.StatusForbidden, response.recorder.Code)
	assert.Equal(t, wholeBodySwap(), swapHeaders(response.recorder.Header()))
}

func TestErrorTellsHowLongToWaitAfterTooManyRequests(t *testing.T) {
	t.Parallel()

	response := respondWithError(t, nil, false, web.ErrTooManyRequests(90*time.Second))

	assert.Equal(t, http.StatusTooManyRequests, response.recorder.Code)
	assert.Equal(t, "90", response.recorder.Header().Get("Retry-After"))
}

func TestErrorRejectsAnOversizedBodyWithoutReportingIt(t *testing.T) {
	t.Parallel()

	response := respondWithError(t, nil, true, fmt.Errorf("parse the form: %w", &http.MaxBytesError{Limit: 1}))

	assert.Equal(t, http.StatusRequestEntityTooLarge, response.recorder.Code)
	assert.Empty(t, response.reporter.reported)
}
