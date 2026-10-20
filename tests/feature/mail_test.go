//go:build feature

package feature_test

import (
	"net/http"
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAQueuedMailIsSentWithBothPartsWhenTheTestRunsTheQueuedJobs(t *testing.T) {
	t.Parallel()

	running := startApp(t)

	queued := running.send(t, outgoing{method: http.MethodPost, path: probeMail, headers: nil, body: nil})
	require.Equal(t, http.StatusNoContent, queued.status)
	require.Empty(t, running.Outbox.Sent(), "a queued mail waits for the worker")

	running.RunQueuedJobs(t)

	sent := running.Outbox.Sent()
	require.Len(t, sent, 1)
	assert.Equal(t, probeMailRecipient, sent[0].To)
	assert.Equal(t, probeMailSubject, sent[0].Subject)
	assert.Equal(t, "Hello "+probeMailName+",\nthis mail was rendered when it was queued.\n", sent[0].Text)

	html, err := goquery.NewDocumentFromReader(strings.NewReader(sent[0].HTML))
	require.NoError(t, err)
	assert.Equal(t, probeMailSubject, html.Find("head title").Text())
	assert.Equal(t, "Hello "+probeMailName, html.Find("body p").Text())
}
