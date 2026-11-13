package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"net/http"
	"time"
)

// SessionCookieName is the host-only SSO cookie name. The __Host- prefix
// requires Secure, Path=/, and no Domain attribute.
const SessionCookieName = "__Host-sso_session"

const sessionTokenBytes = 32

// NewSessionToken returns an opaque 256-bit token for the browser and the
// SHA-256 hash that is safe to persist.
func NewSessionToken() (string, []byte, error) {
	buffer := make([]byte, sessionTokenBytes)
	if _, err := rand.Read(buffer); err != nil {
		return "", nil, fmt.Errorf("generate session token: %w", err)
	}

	raw := base64.RawURLEncoding.EncodeToString(buffer)
	return raw, HashSessionToken(raw), nil
}

// HashSessionToken returns the stored hash of an opaque Session token.
func HashSessionToken(raw string) []byte {
	sum := sha256.Sum256([]byte(raw))
	return sum[:]
}

// SessionCookie builds the host-only Session cookie. The browser keeps it until
// the absolute expiry; the server still enforces the shorter idle TTL.
func SessionCookie(token string, expiresAt, now time.Time) *http.Cookie {
	cookie := sessionCookieBase() // #nosec G124 -- sessionCookieBase sets Secure, HttpOnly, SameSite=Lax, and Path=/.
	cookie.Value = token
	cookie.MaxAge = int(expiresAt.Sub(now) / time.Second)
	cookie.Expires = expiresAt.UTC()
	return cookie
}

// ExpiredSessionCookie clears the Session cookie in the browser.
func ExpiredSessionCookie() *http.Cookie {
	cookie := sessionCookieBase() // #nosec G124 -- sessionCookieBase sets Secure, HttpOnly, SameSite=Lax, and Path=/.
	cookie.MaxAge = -1
	cookie.Expires = time.Unix(1, 0).UTC()
	return cookie
}

func sessionCookieBase() *http.Cookie {
	return &http.Cookie{
		Name:     SessionCookieName,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	}
}
