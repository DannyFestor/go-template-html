//go:build feature

package feature_test

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestALargePageIsCompressedForAClientAcceptingGzip(t *testing.T) {
	t.Parallel()

	fetched := startApp(t).send(t, acceptingEncoding(probeLargePage, "gzip"))

	assert.Equal(t, http.StatusOK, fetched.status)
	assert.Equal(t, "gzip", fetched.header.Get("Content-Encoding"))
	assert.Equal(t, largePage(), gunzipped(t, fetched.body))
}

func TestAPageThatSkipsCompressionIsSentPlain(t *testing.T) {
	t.Parallel()

	fetched := startApp(t).send(t, acceptingEncoding(probeSecretPage, "gzip, zstd"))

	assert.Equal(t, http.StatusOK, fetched.status)
	assert.Empty(t, fetched.header.Get("Content-Encoding"))
	assert.Empty(t, fetched.header.Values("No-Gzip-Compression"))
	assert.Equal(t, largePage(), string(fetched.body))
}

func TestAHashedAssetIsServedFromItsPrecompressedSibling(t *testing.T) {
	t.Parallel()

	running := startApp(t)
	script := builtAppEntry(t).File

	for encoding, extension := range map[string]string{"br": ".br", "gzip": ".gz"} {
		t.Run(encoding, func(t *testing.T) {
			t.Parallel()

			fetched := running.send(t, acceptingEncoding("/"+script, encoding))

			assert.Equal(t, http.StatusOK, fetched.status)
			assert.Equal(t, encoding, fetched.header.Get("Content-Encoding"))
			assert.Equal(t, builtFile(t, script+extension), fetched.body)
		})
	}
}

func acceptingEncoding(path, encodings string) outgoing {
	return outgoing{
		method:  http.MethodGet,
		path:    path,
		headers: map[string]string{"Accept-Encoding": encodings},
		body:    nil,
	}
}

func gunzipped(t *testing.T, compressed []byte) string {
	t.Helper()

	reader, err := gzip.NewReader(bytes.NewReader(compressed))
	require.NoError(t, err)

	plain, err := io.ReadAll(reader)
	require.NoError(t, err)

	return string(plain)
}
