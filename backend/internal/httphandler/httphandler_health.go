// Package httphandler wires HTTP routes and translates requests into domain
// calls. Handler logic imports domain, pkg, and the auth module; the router
// also wires middleware.
package httphandler

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/ppabimanyu/sso/internal/domain"
)

// HealthResponse is the body of a liveness or readiness response.
type HealthResponse struct {
	Status string            `json:"status" example:"ok"`
	Checks map[string]string `json:"checks,omitempty"`
}

// HealthHandler serves the liveness and readiness endpoints.
type HealthHandler struct {
	checks map[string]domain.HealthChecker
	logger *slog.Logger
}

// NewHealthHandler builds the health handler. Each named checker is reported
// separately on the readiness endpoint.
func NewHealthHandler(checks map[string]domain.HealthChecker, logger *slog.Logger) *HealthHandler {
	return &HealthHandler{checks: checks, logger: logger}
}

// Live godoc
// @Summary Liveness check
// @Description Returns 200 while the process is running.
// @Tags health
// @Produce json
// @Success 200 {object} HealthResponse
// @Router /healthz [get]
func (h *HealthHandler) Live(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, HealthResponse{Status: "ok"})
}

// Ready godoc
// @Summary Readiness check
// @Description Returns 200 when every dependency answers, otherwise 503.
// @Tags health
// @Produce json
// @Success 200 {object} HealthResponse
// @Failure 503 {object} HealthResponse
// @Router /readyz [get]
func (h *HealthHandler) Ready(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	response := HealthResponse{Status: "ok", Checks: make(map[string]string, len(h.checks))}
	healthy := true

	for name, check := range h.checks {
		if err := check.Ping(ctx); err != nil {
			response.Checks[name] = "error"
			healthy = false
			h.logger.WarnContext(ctx, "readiness check failed", "dependency", name, "error", err)
			continue
		}
		response.Checks[name] = "ok"
	}

	if !healthy {
		response.Status = "degraded"
		writeJSON(w, http.StatusServiceUnavailable, response)
		return
	}

	writeJSON(w, http.StatusOK, response)
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}
