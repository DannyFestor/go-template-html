package middleware

import "net/http"

func LimitBody(maxBytes int64) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r.Body = http.MaxBytesReader(innermostWriter(w), r.Body, maxBytes)

			next.ServeHTTP(w, r)
		})
	}
}

// innermostWriter reaches the server's own writer, because MaxBytesReader closes the connection only through it.
func innermostWriter(w http.ResponseWriter) http.ResponseWriter {
	for {
		wrapper, isWrapper := w.(interface{ Unwrap() http.ResponseWriter })
		if !isWrapper {
			return w
		}

		w = wrapper.Unwrap()
	}
}
