package testmailpit

import (
	"encoding/json"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type Address struct {
	Name    string
	Address string
}

type Message struct {
	From    Address
	To      []Address
	Subject string
	Date    time.Time
	Text    string
	HTML    string
}

type searchResult struct {
	Messages []struct{ ID string }
}

func (i Instance) MessageTo(t *testing.T, recipient string) Message {
	t.Helper()

	var found searchResult

	i.get(t, "/search?"+url.Values{"query": {`to:"` + recipient + `"`}}.Encode(), &found)
	require.Len(t, found.Messages, 1, "mail to %s", recipient)

	var message Message

	i.get(t, "/message/"+url.PathEscape(found.Messages[0].ID), &message)

	return message
}

func (i Instance) get(t *testing.T, path string, into any) {
	t.Helper()

	request, err := http.NewRequestWithContext(t.Context(), http.MethodGet, i.apiURL+path, nil)
	require.NoError(t, err)

	response, err := http.DefaultClient.Do(request)
	require.NoError(t, err)

	defer func() { _ = response.Body.Close() }()

	require.Equal(t, http.StatusOK, response.StatusCode, path)
	require.NoError(t, json.NewDecoder(response.Body).Decode(into))
}
