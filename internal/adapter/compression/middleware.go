package compression

import (
	"fmt"
	"net/http"

	"github.com/klauspost/compress/gzhttp"
)

const minCompressedBytes = 1 << 10

func New() (func(http.Handler) http.Handler, error) {
	wrap, err := gzhttp.NewWrapper(
		gzhttp.ContentTypes(compressibleContentTypes()),
		gzhttp.MinSize(minCompressedBytes),
	)
	if err != nil {
		return nil, fmt.Errorf("configure response compression: %w", err)
	}

	return func(next http.Handler) http.Handler {
		return wrap(next)
	}, nil
}

func compressibleContentTypes() []string {
	return []string{
		"text/html",
		"text/plain",
		"text/css",
		"text/javascript",
		"application/javascript",
		"application/json",
		"image/svg+xml",
		"application/xml",
	}
}
