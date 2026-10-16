package middleware

import (
	"context"
	"net/http"
	"net/netip"
)

type clientIPKey struct{}

func ClientIP(trusted []netip.Prefix) func(http.Handler) http.Handler {
	proxies := trustedProxies(trusted)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), clientIPKey{}, proxies.clientIP(r))))
		})
	}
}

func ClientIPFrom(ctx context.Context) netip.Addr {
	addr, _ := ctx.Value(clientIPKey{}).(netip.Addr)

	return addr
}
