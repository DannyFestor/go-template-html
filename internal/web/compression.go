package web

import "net/http"

const noCompressionHeader = "No-Gzip-Compression"

// SkipCompression guards a page that reveals a server-generated secret against BREACH: compressing a secret next to
// text an attacker can reflect into the page leaks the secret through the response size.
func SkipCompression(w http.ResponseWriter) {
	w.Header().Set(noCompressionHeader, "1")
}
