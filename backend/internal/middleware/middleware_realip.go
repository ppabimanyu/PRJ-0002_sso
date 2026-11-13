package middleware

import (
	"net"
	"net/http"
	"net/netip"
	"strings"
)

// RealIP replaces the peer address with the client address from
// X-Forwarded-For, but only when the direct peer is a configured trusted
// proxy. Without trusted proxies it leaves RemoteAddr alone, so a client
// cannot spoof its own address by sending the header.
func RealIP(trustedProxies []netip.Prefix) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if len(trustedProxies) == 0 {
				next.ServeHTTP(w, r)
				return
			}

			peer, err := netip.ParseAddr(hostOnly(r.RemoteAddr))
			if err != nil || !isTrusted(peer, trustedProxies) {
				next.ServeHTTP(w, r)
				return
			}

			if client := clientFromForwardedFor(r.Header.Values("X-Forwarded-For"), trustedProxies); client != "" {
				r.RemoteAddr = client
			}

			next.ServeHTTP(w, r)
		})
	}
}

func clientFromForwardedFor(headers []string, trustedProxies []netip.Prefix) string {
	if len(headers) == 0 {
		return ""
	}

	parts := strings.Split(strings.Join(headers, ","), ",")
	for index := len(parts) - 1; index >= 0; index-- {
		address, err := netip.ParseAddr(strings.TrimSpace(parts[index]))
		if err != nil {
			return ""
		}
		if isTrusted(address, trustedProxies) {
			continue
		}
		return address.Unmap().String()
	}
	return ""
}

func isTrusted(address netip.Addr, trustedProxies []netip.Prefix) bool {
	address = address.Unmap()
	for _, prefix := range trustedProxies {
		if prefix.Contains(address) {
			return true
		}
	}
	return false
}

func hostOnly(address string) string {
	host, _, err := net.SplitHostPort(address)
	if err != nil {
		return address
	}
	return host
}
