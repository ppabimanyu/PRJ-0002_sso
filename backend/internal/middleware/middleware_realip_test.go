package middleware

import (
	"net/http"
	"net/http/httptest"
	"net/netip"
	"testing"

	"github.com/stretchr/testify/assert"
)

func runRealIP(trusted []netip.Prefix, remoteAddr string, forwardedFor ...string) string {
	var observed string
	handler := RealIP(trusted)(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		observed = r.RemoteAddr
	}))

	request := httptest.NewRequest(http.MethodGet, "/", nil)
	request.RemoteAddr = remoteAddr
	for _, value := range forwardedFor {
		request.Header.Add("X-Forwarded-For", value)
	}
	handler.ServeHTTP(httptest.NewRecorder(), request)
	return observed
}

func TestRealIPIgnoresForwardedForWithoutTrustedProxies(t *testing.T) {
	assert.Equal(t, "198.51.100.7:1234", runRealIP(nil, "198.51.100.7:1234", "203.0.113.9"))
}

func TestRealIPUsesForwardedForFromTrustedProxy(t *testing.T) {
	trusted := []netip.Prefix{netip.MustParsePrefix("10.0.0.0/8")}
	assert.Equal(t, "203.0.113.9", runRealIP(trusted, "10.0.0.5:443", "203.0.113.9"))
}

func TestRealIPIgnoresForwardedForFromUntrustedPeer(t *testing.T) {
	trusted := []netip.Prefix{netip.MustParsePrefix("10.0.0.0/8")}
	assert.Equal(t, "198.51.100.7:1234", runRealIP(trusted, "198.51.100.7:1234", "203.0.113.9"))
}

func TestRealIPSkipsTrustedHopsFromTheRight(t *testing.T) {
	trusted := []netip.Prefix{netip.MustParsePrefix("10.0.0.0/8")}
	assert.Equal(t, "203.0.113.9", runRealIP(trusted, "10.0.0.5:443", "203.0.113.9, 10.0.0.6"))
}

func TestRealIPKeepsPeerWhenHeaderIsInvalid(t *testing.T) {
	trusted := []netip.Prefix{netip.MustParsePrefix("10.0.0.0/8")}
	assert.Equal(t, "10.0.0.5:443", runRealIP(trusted, "10.0.0.5:443", "not-an-ip"))
}
