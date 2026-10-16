package middleware

import (
	"net/http"
	"net/netip"
	"slices"
	"strings"
)

type trustedProxies []netip.Prefix

// clientIP walks X-Forwarded-For from the right, because only the entries our own proxies appended can be believed.
func (t trustedProxies) clientIP(r *http.Request) netip.Addr {
	client := remoteAddr(r)
	if !t.trusts(client) {
		return client
	}

	for _, hop := range slices.Backward(forwardedFor(r.Header)) {
		addr, err := netip.ParseAddr(hop)
		if err != nil {
			return client
		}

		client = addr.Unmap()
		if !t.trusts(client) {
			return client
		}
	}

	return client
}

func (t trustedProxies) trusts(addr netip.Addr) bool {
	return slices.ContainsFunc(t, func(prefix netip.Prefix) bool { return prefix.Contains(addr) })
}

func remoteAddr(r *http.Request) netip.Addr {
	addrPort, err := netip.ParseAddrPort(r.RemoteAddr)
	if err != nil {
		return netip.Addr{}
	}

	return addrPort.Addr().Unmap()
}

func forwardedFor(header http.Header) []string {
	var hops []string

	for _, value := range header.Values("X-Forwarded-For") {
		for hop := range strings.SplitSeq(value, ",") {
			hops = append(hops, strings.TrimSpace(hop))
		}
	}

	return hops
}
