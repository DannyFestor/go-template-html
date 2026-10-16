package asset_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/fstest"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/DannyFestor/go-template-html/internal/asset"
)

const (
	entry    = "internal/asset/src/app.ts"
	manifest = `{
  "internal/asset/src/app.ts": {
    "file": "assets/app-nx4752wa.js",
    "name": "app",
    "src": "internal/asset/src/app.ts",
    "isEntry": true,
    "css": ["assets/app-fimgy6vt.css"]
  }
}`
)

func TestScriptURLResolvesTheEntrysHashedFile(t *testing.T) {
	t.Parallel()

	server := newImmutable(t, dist())

	url, err := server.ScriptURL(entry)

	require.NoError(t, err)
	assert.Equal(t, "/assets/app-nx4752wa.js", url)
}

func TestStylesheetURLsResolveTheEntrysHashedCSS(t *testing.T) {
	t.Parallel()

	server := newImmutable(t, dist())

	urls, err := server.StylesheetURLs(entry)

	require.NoError(t, err)
	assert.Equal(t, []string{"/assets/app-fimgy6vt.css"}, urls)
}

func TestURLsFailForAnEntryMissingFromTheManifest(t *testing.T) {
	t.Parallel()

	server := newImmutable(t, dist())

	_, scriptErr := server.ScriptURL("internal/asset/src/admin.ts")
	_, stylesheetErr := server.StylesheetURLs("internal/asset/src/admin.ts")

	require.ErrorContains(t, scriptErr, "internal/asset/src/admin.ts")
	require.ErrorContains(t, stylesheetErr, "internal/asset/src/admin.ts")
}

func TestNewImmutableFailsWithoutAManifest(t *testing.T) {
	t.Parallel()

	files := dist()
	delete(files, "manifest.json")

	_, err := asset.NewImmutable(files)

	require.ErrorContains(t, err, "manifest.json")
}

func TestImmutableServesTheCurrentFileForAYear(t *testing.T) {
	t.Parallel()

	response := get(t, newImmutable(t, dist()), "/assets/app-fimgy6vt.css")

	assert.Equal(t, http.StatusOK, response.Code)
	assert.Equal(t, "body{}", response.Body.String())
	assert.Equal(t, "text/css; charset=utf-8", response.Header().Get("Content-Type"))
	assert.Equal(t, "public, max-age=31536000, immutable", response.Header().Get("Cache-Control"))
}

func TestImmutableServesTheCurrentFileUncachedForAnOutdatedHash(t *testing.T) {
	t.Parallel()

	response := get(t, newImmutable(t, dist()), "/assets/app-0ld0hash.js")

	assert.Equal(t, http.StatusOK, response.Code)
	assert.Equal(t, "console.log('app')", response.Body.String())
	assert.Equal(t, "no-cache", response.Header().Get("Cache-Control"))
}

func TestServesThePrecompressedSiblingTheClientAccepts(t *testing.T) {
	t.Parallel()

	server := newImmutable(t, dist())

	for encoding, body := range map[string]string{"br": "brotli body{}", "gzip": "gzip body{}"} {
		t.Run(encoding, func(t *testing.T) {
			t.Parallel()

			response := getAccepting(t, server, "/assets/app-fimgy6vt.css", encoding)

			assert.Equal(t, http.StatusOK, response.Code)
			assert.Equal(t, encoding, response.Header().Get("Content-Encoding"))
			assert.Equal(t, body, response.Body.String())
			assert.Equal(t, "text/css; charset=utf-8", response.Header().Get("Content-Type"))
			assert.Equal(t, "public, max-age=31536000, immutable", response.Header().Get("Cache-Control"))
		})
	}
}

func TestServesTheUncompressedFileToAClientAcceptingNeitherEncoding(t *testing.T) {
	t.Parallel()

	response := getAccepting(t, newImmutable(t, dist()), "/assets/app-fimgy6vt.css", "zstd")

	assert.Empty(t, response.Header().Get("Content-Encoding"))
	assert.Equal(t, "body{}", response.Body.String())
}

func TestLiveServesThePrecompressedSiblingOfTheLatestBuild(t *testing.T) {
	t.Parallel()

	files := dist()
	server := asset.NewLive(files)
	files["manifest.json"] = file(`{"internal/asset/src/app.ts": {"file": "assets/app-n3wb1ld0.js"}}`)
	files["assets/app-n3wb1ld0.js"] = file("console.log('rebuilt')")
	files["assets/app-n3wb1ld0.js.br"] = file("brotli rebuilt")

	response := getAccepting(t, server, "/assets/app-n3wb1ld0.js", "br")

	assert.Equal(t, http.StatusOK, response.Code)
	assert.Equal(t, "brotli rebuilt", response.Body.String())
}

func TestServesNothingTheBuildDidNotEmit(t *testing.T) {
	t.Parallel()

	files := dist()
	files[".gitkeep"] = file("")
	files["assets/admin-nx4752wa.js"] = file("console.log('admin')")
	server := newImmutable(t, files)

	for _, path := range []string{
		"/manifest.json",
		"/.gitkeep",
		"/assets/admin-nx4752wa.js",
		"/assets/app.js",
		"/assets/app-fimgy6vt.css.br",
	} {
		t.Run(path, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, http.StatusNotFound, get(t, server, path).Code)
		})
	}
}

func TestLiveResolvesURLsFromTheLatestBuild(t *testing.T) {
	t.Parallel()

	files := dist()
	server := asset.NewLive(files)
	files["manifest.json"] = file(`{"internal/asset/src/app.ts": {"file": "assets/app-n3wb1ld0.js"}}`)

	url, err := server.ScriptURL(entry)

	require.NoError(t, err)
	assert.Equal(t, "/assets/app-n3wb1ld0.js", url)
}

func TestLiveForbidsCachingAnyFile(t *testing.T) {
	t.Parallel()

	server := asset.NewLive(dist())

	for _, path := range []string{"/assets/app-nx4752wa.js", "/assets/app-0ld0hash.js"} {
		t.Run(path, func(t *testing.T) {
			t.Parallel()

			response := get(t, server, path)

			assert.Equal(t, http.StatusOK, response.Code)
			assert.Equal(t, "no-store", response.Header().Get("Cache-Control"))
		})
	}
}

func TestLiveFailsWhileTheManifestIsMissing(t *testing.T) {
	t.Parallel()

	files := dist()
	delete(files, "manifest.json")
	server := asset.NewLive(files)

	_, err := server.ScriptURL(entry)

	require.ErrorContains(t, err, "manifest.json")
	assert.Equal(t, http.StatusInternalServerError, get(t, server, "/assets/app-nx4752wa.js").Code)
}

func newImmutable(t *testing.T, files fstest.MapFS) *asset.Server {
	t.Helper()

	server, err := asset.NewImmutable(files)
	require.NoError(t, err)

	return server
}

func get(t *testing.T, server http.Handler, path string) *httptest.ResponseRecorder {
	t.Helper()

	return getAccepting(t, server, path, "")
}

func getAccepting(t *testing.T, server http.Handler, path, encodings string) *httptest.ResponseRecorder {
	t.Helper()

	request := httptest.NewRequestWithContext(t.Context(), http.MethodGet, path, nil)
	request.Header.Set("Accept-Encoding", encodings)

	recorder := httptest.NewRecorder()
	server.ServeHTTP(recorder, request)

	return recorder
}

func dist() fstest.MapFS {
	return fstest.MapFS{
		"manifest.json":              file(manifest),
		"assets/app-nx4752wa.js":     file("console.log('app')"),
		"assets/app-fimgy6vt.css":    file("body{}"),
		"assets/app-fimgy6vt.css.br": file("brotli body{}"),
		"assets/app-fimgy6vt.css.gz": file("gzip body{}"),
	}
}

func file(content string) *fstest.MapFile {
	return &fstest.MapFile{Data: []byte(content), Mode: 0, ModTime: time.Time{}, Sys: nil}
}
