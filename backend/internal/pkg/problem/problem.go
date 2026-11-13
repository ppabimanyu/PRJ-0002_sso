// Package problem builds RFC 7807 problem documents and maps domain errors to
// HTTP status codes in one place.
package problem

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"github.com/ppabimanyu/sso/internal/domain"
)

// ContentType is the media type defined by RFC 7807.
const ContentType = "application/problem+json"

// Problem is an RFC 7807 problem detail document.
type Problem struct {
	Type     string `json:"type"`
	Title    string `json:"title"`
	Status   int    `json:"status"`
	Detail   string `json:"detail,omitempty"`
	Instance string `json:"instance,omitempty"`
}

// New builds a problem with the default type "about:blank".
func New(status int, title, detail string) *Problem {
	return &Problem{Type: "about:blank", Title: title, Status: status, Detail: detail}
}

func (p *Problem) Error() string {
	if p.Detail == "" {
		return p.Title
	}
	return p.Title + ": " + p.Detail
}

// FromError maps a domain error to a problem. Unknown errors become a generic
// 500 so internal details never reach the client.
func FromError(err error) *Problem {
	switch {
	case errors.Is(err, domain.ErrNotFound):
		return New(http.StatusNotFound, "Not Found", "The requested resource does not exist.")
	case errors.Is(err, domain.ErrConflict):
		return New(http.StatusConflict, "Conflict", "The request conflicts with the current state.")
	case errors.Is(err, domain.ErrValidation):
		return New(http.StatusBadRequest, "Bad Request", strings.TrimPrefix(err.Error(), domain.ErrValidation.Error()+": "))
	default:
		return New(http.StatusInternalServerError, "Internal Server Error", "")
	}
}

// Write maps err, logs server-side failures with the request context, and
// writes the problem document.
func Write(w http.ResponseWriter, r *http.Request, err error, logger *slog.Logger) {
	p := FromError(err)
	if p.Status >= http.StatusInternalServerError {
		logger.ErrorContext(r.Context(), "request failed", "error", err, "method", r.Method, "path", r.URL.Path)
	}
	WriteStatus(w, r, p)
}

// WriteStatus writes a problem document. Instance is filled from the request
// path when the caller left it empty.
func WriteStatus(w http.ResponseWriter, r *http.Request, p *Problem) {
	if p.Instance == "" {
		p.Instance = r.URL.Path
	}
	w.Header().Set("Content-Type", ContentType)
	w.WriteHeader(p.Status)
	_ = json.NewEncoder(w).Encode(p)
}
