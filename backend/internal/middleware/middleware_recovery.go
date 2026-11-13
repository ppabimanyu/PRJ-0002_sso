package middleware

import (
	"errors"
	"log/slog"
	"net/http"
	"runtime/debug"

	"github.com/ppabimanyu/sso/internal/pkg/problem"
)

// Recovery logs a panic with its stack and returns a problem document.
// It re-panics on http.ErrAbortHandler so the server still aborts the response.
func Recovery(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				recovered := recover()
				if recovered == nil {
					return
				}
				if err, ok := recovered.(error); ok && errors.Is(err, http.ErrAbortHandler) {
					panic(recovered)
				}
				logger.ErrorContext(r.Context(), "panic recovered",
					"panic", recovered,
					"stack", string(debug.Stack()),
					"method", r.Method,
					"path", r.URL.Path,
				)
				problem.WriteStatus(w, r, problem.New(http.StatusInternalServerError, "Internal Server Error", ""))
			}()
			next.ServeHTTP(w, r)
		})
	}
}
