package web

import (
	"math"
	"net/http"
	"strconv"
	"time"
)

func setRetryAfter(header http.Header, wait time.Duration) {
	header.Set("Retry-After", strconv.Itoa(int(math.Ceil(wait.Seconds()))))
}
