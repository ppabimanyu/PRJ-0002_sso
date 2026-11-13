package httphandler

import (
	"log/slog"
	"net/http"
	"net/netip"

	"github.com/go-chi/chi/v5"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"

	"github.com/ppabimanyu/sso/internal/domain"
	"github.com/ppabimanyu/sso/internal/middleware"
	"github.com/ppabimanyu/sso/internal/pkg/problem"
)

// RouterOptions carries everything the router needs from the outside world.
// Repositories arrive as domain ports so this package never imports them.
type RouterOptions struct {
	Logger         *slog.Logger
	Readiness      map[string]domain.HealthChecker
	AllowedOrigins []string
	TrustedProxies []netip.Prefix
	RateLimit      func(http.Handler) http.Handler
	MaxBodyBytes   int64
}

// NewRouter builds the HTTP handler with the full middleware chain.
func NewRouter(opts RouterOptions) http.Handler {
	router := chi.NewRouter()

	router.Use(middleware.Recovery(opts.Logger))
	router.Use(otelhttp.NewMiddleware("http.server", otelhttp.WithSpanNameFormatter(spanName)))
	router.Use(middleware.InFlight("github.com/ppabimanyu/sso/internal/middleware"))
	router.Use(middleware.TraceID)
	router.Use(middleware.RequestID)
	router.Use(middleware.AccessLog(opts.Logger))
	router.Use(middleware.RealIP(opts.TrustedProxies))
	router.Use(middleware.BodyLimit(opts.MaxBodyBytes))
	if cors := middleware.CORS(opts.AllowedOrigins); cors != nil {
		router.Use(cors)
	}
	if opts.RateLimit != nil {
		router.Use(opts.RateLimit)
	}

	health := NewHealthHandler(opts.Readiness, opts.Logger)
	router.Get("/healthz", health.Live)
	router.Get("/readyz", health.Ready)

	router.Route("/api/v1", func(_ chi.Router) {
		// Feature routes mount here. Each feature gets its own
		// httphandler_<resource>.go file and a port in domain.
	})

	router.NotFound(func(w http.ResponseWriter, r *http.Request) {
		problem.WriteStatus(w, r, problem.New(http.StatusNotFound, "Not Found", "The requested resource does not exist."))
	})
	router.MethodNotAllowed(func(w http.ResponseWriter, r *http.Request) {
		problem.WriteStatus(w, r, problem.New(http.StatusMethodNotAllowed, "Method Not Allowed", "This method is not supported for the resource."))
	})

	return router
}

func spanName(_ string, r *http.Request) string {
	return r.Method + " " + r.URL.Path
}
