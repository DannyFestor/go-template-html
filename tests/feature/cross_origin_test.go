//go:build feature

package feature_test

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

const trustedOrigin = "https://partner.example"

func postWith(headers map[string]string) outgoing {
	return outgoing{method: http.MethodPost, path: probeRedirect, headers: headers, body: nil}
}

func TestACrossOriginPostIsRejectedWithTheForbiddenPage(t *testing.T) {
	t.Parallel()

	requests := map[string]map[string]string{
		"cross-site fetch": {"Sec-Fetch-Site": "cross-site"},
		"foreign origin":   {"Origin": "https://evil.example"},
	}

	for name, headers := range requests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			fetched := startApp(t).send(t, postWith(headers))

			assert.Equal(t, http.StatusForbidden, fetched.status)
			assert.Equal(t, "Forbidden", fetched.document(t).Find("main#public-main h1").Text())
		})
	}
}

func TestASameOriginPostGoesThrough(t *testing.T) {
	t.Parallel()

	fetched := startApp(t).send(t, postWith(map[string]string{"Sec-Fetch-Site": "same-origin"}))

	assert.Equal(t, http.StatusSeeOther, fetched.status)
}

func TestAPostFromATrustedOriginGoesThrough(t *testing.T) {
	t.Parallel()

	config := liveConfig(t)
	config.HTTP.CSRFTrustedOrigins = []string{trustedOrigin}

	fetched := startAppOn(t, config).send(t, postWith(map[string]string{
		"Sec-Fetch-Site": "cross-site",
		"Origin":         trustedOrigin,
	}))

	assert.Equal(t, http.StatusSeeOther, fetched.status)
}
