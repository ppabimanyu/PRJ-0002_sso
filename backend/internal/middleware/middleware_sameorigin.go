package middleware

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/ppabimanyu/sso/internal/pkg/problem"
)

var safeMethods = map[string]bool{
	http.MethodGet:     true,
	http.MethodHead:    true,
	http.MethodOptions: true,
	http.MethodTrace:   true,
}

// SameOrigin rejects state-changing browser requests that do not come from the
// auth origin. The deployment is single-origin, so anything else is a CSRF
// attempt regardless of the Session cookie.
func SameOrigin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if safeMethods[r.Method] {
			next.ServeHTTP(w, r)
			return
		}

		if site := r.Header.Get("Sec-Fetch-Site"); site != "" && site != "same-origin" && site != "none" {
			problem.WriteStatus(w, r, problem.New(http.StatusForbidden, "Forbidden", "Cross-origin request rejected."))
			return
		}

		if origin := r.Header.Get("Origin"); origin != "" {
			parsed, err := url.Parse(origin)
			if err != nil || !strings.EqualFold(parsed.Host, r.Host) {
				problem.WriteStatus(w, r, problem.New(http.StatusForbidden, "Forbidden", "Cross-origin request rejected."))
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}
