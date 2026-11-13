package middleware

import (
	"net/http"
	"time"

	"github.com/go-chi/cors"
)

// CORS allows the configured browser origins. An empty list means no CORS
// middleware, which is the same-origin deployment default.
func CORS(allowedOrigins []string) func(http.Handler) http.Handler {
	if len(allowedOrigins) == 0 {
		return nil
	}
	return cors.Handler(cors.Options{
		AllowedOrigins: allowedOrigins,
		AllowedMethods: []string{
			http.MethodGet,
			http.MethodPost,
			http.MethodPut,
			http.MethodPatch,
			http.MethodDelete,
			http.MethodOptions,
		},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Request-Id"},
		ExposedHeaders:   []string{RequestIDHeader, TraceIDHeader},
		AllowCredentials: false,
		MaxAge:           int((5 * time.Minute).Seconds()),
	})
}
