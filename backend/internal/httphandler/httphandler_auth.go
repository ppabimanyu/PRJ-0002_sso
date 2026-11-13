package httphandler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/ppabimanyu/sso/internal/auth"
	"github.com/ppabimanyu/sso/internal/domain"
	"github.com/ppabimanyu/sso/internal/pkg/problem"
	"github.com/ppabimanyu/sso/internal/pkg/validate"
)

// UserResponse is the public profile shape. Disabled and deletion flags stay
// server-side.
type UserResponse struct {
	ID            string    `json:"id"`
	Email         string    `json:"email"`
	Name          string    `json:"name"`
	PictureURL    string    `json:"pictureUrl,omitempty"`
	IsAdmin       bool      `json:"isAdmin"`
	EmailVerified bool      `json:"emailVerified"`
	CreatedAt     time.Time `json:"createdAt"`
}

// LoginRequest is one password login attempt.
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// LoginResponse returns the authenticated User. The Session token is only in
// the Set-Cookie header.
type LoginResponse struct {
	User UserResponse `json:"user"`
}

// SessionResponse describes the current cookie Session.
type SessionResponse struct {
	Authenticated bool          `json:"authenticated"`
	User          *UserResponse `json:"user,omitempty"`
	ExpiresAt     *time.Time    `json:"expiresAt,omitempty"`
}

type sessionContextKey struct{}

// SessionContext is the authenticated User and Session attached by RequireSession.
type SessionContext struct {
	User    domain.User
	Session domain.Session
}

// AuthHandler serves the password login, Session, logout, and profile routes.
type AuthHandler struct {
	service *auth.Service
	logger  *slog.Logger
	meta    RequestMetaFunc
}

// RequestMetaFunc supplies transport metadata from the router wiring, which is
// the only layer allowed to import middleware.
type RequestMetaFunc func(*http.Request) auth.RequestMeta

// NewAuthHandler builds the auth HTTP surface. meta may be nil in tests that do
// not care about Audit Event request metadata.
func NewAuthHandler(service *auth.Service, logger *slog.Logger, meta RequestMetaFunc) *AuthHandler {
	return &AuthHandler{service: service, logger: logger, meta: meta}
}

// Login godoc
// @Summary Password login
// @Description Verifies email and password, creates a Session, and sets the host-only Session cookie.
// @Tags auth
// @Accept json
// @Produce json
// @Param request body LoginRequest true "Email and password"
// @Success 200 {object} LoginResponse
// @Failure 400 {object} problem.Problem
// @Failure 401 {object} problem.Problem
// @Failure 429 {object} problem.Problem
// @Router /api/v1/auth/login [post]
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var request LoginRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&request); err != nil {
		problem.Write(w, r, fmt.Errorf("%w: request body must be a JSON object with email and password", domain.ErrValidation), h.logger)
		return
	}
	if err := validate.Struct(request); err != nil {
		problem.Write(w, r, err, h.logger)
		return
	}

	result, err := h.service.Login(r.Context(), auth.LoginInput{
		Email:       request.Email,
		Password:    request.Password,
		RequestMeta: h.requestMeta(r),
	})
	if err != nil {
		h.writeLoginError(w, r, err)
		return
	}

	http.SetCookie(w, auth.SessionCookie(result.Token, result.Session.AbsoluteExpiresAt, result.Session.CreatedAt))
	writeJSON(w, http.StatusOK, LoginResponse{User: newUserResponse(result.User)})
}

// Logout godoc
// @Summary Logout
// @Description Revokes the current Session and clears the Session cookie. The endpoint is idempotent.
// @Tags auth
// @Success 204
// @Router /api/v1/auth/logout [post]
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	if cookie, err := r.Cookie(auth.SessionCookieName); err == nil {
		if err := h.service.Logout(r.Context(), cookie.Value, h.requestMeta(r)); err != nil {
			http.SetCookie(w, auth.ExpiredSessionCookie())
			problem.Write(w, r, err, h.logger)
			return
		}
	}

	http.SetCookie(w, auth.ExpiredSessionCookie())
	w.WriteHeader(http.StatusNoContent)
}

// Session godoc
// @Summary Current Session
// @Description Returns the authenticated User and Session expiry, or authenticated false when there is no valid cookie.
// @Tags auth
// @Produce json
// @Success 200 {object} SessionResponse
// @Router /api/v1/auth/session [get]
func (h *AuthHandler) Session(w http.ResponseWriter, r *http.Request) {
	user, session, err := h.resolveSession(r)
	if err != nil {
		if !errors.Is(err, domain.ErrUnauthorized) {
			problem.Write(w, r, err, h.logger)
			return
		}
		http.SetCookie(w, auth.ExpiredSessionCookie())
		writeJSON(w, http.StatusOK, SessionResponse{Authenticated: false})
		return
	}

	profile := newUserResponse(user)
	expiresAt := session.ExpiresAt()
	writeJSON(w, http.StatusOK, SessionResponse{
		Authenticated: true,
		User:          &profile,
		ExpiresAt:     &expiresAt,
	})
}

// Me godoc
// @Summary Current profile
// @Description Returns the profile of the authenticated User.
// @Tags account
// @Produce json
// @Success 200 {object} UserResponse
// @Failure 401 {object} problem.Problem
// @Router /api/v1/me [get]
func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	current, ok := SessionFromContext(r.Context())
	if !ok {
		problem.WriteStatus(w, r, problem.New(http.StatusUnauthorized, "Unauthorized", "Authentication is required."))
		return
	}
	writeJSON(w, http.StatusOK, newUserResponse(current.User))
}

// RequireSession rejects requests without a live Session and puts the User and
// Session in the request context.
func (h *AuthHandler) RequireSession(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, session, err := h.resolveSession(r)
		if err != nil {
			if !errors.Is(err, domain.ErrUnauthorized) {
				problem.Write(w, r, err, h.logger)
				return
			}
			h.writeUnauthorized(w, r)
			return
		}

		ctx := context.WithValue(r.Context(), sessionContextKey{}, SessionContext{User: user, Session: session})
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (h *AuthHandler) resolveSession(r *http.Request) (domain.User, domain.Session, error) {
	cookie, err := r.Cookie(auth.SessionCookieName)
	if err != nil {
		return domain.User{}, domain.Session{}, domain.ErrUnauthorized
	}
	return h.service.AuthenticateSession(r.Context(), cookie.Value)
}

// SessionFromContext returns the authenticated User and Session for routed
// handlers mounted behind RequireSession.
func SessionFromContext(ctx context.Context) (SessionContext, bool) {
	current, ok := ctx.Value(sessionContextKey{}).(SessionContext)
	return current, ok
}

func (h *AuthHandler) writeUnauthorized(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, auth.ExpiredSessionCookie())
	problem.WriteStatus(w, r, problem.New(http.StatusUnauthorized, "Unauthorized", "Authentication is required."))
}

func (h *AuthHandler) writeLoginError(w http.ResponseWriter, r *http.Request, err error) {
	var blocked *auth.BlockedError
	switch {
	case errors.As(err, &blocked):
		retryAfter := int(blocked.RetryAfter.Seconds())
		if retryAfter < 1 {
			retryAfter = 1
		}
		w.Header().Set("Retry-After", strconv.Itoa(retryAfter))
		problem.WriteStatus(w, r, problem.New(http.StatusTooManyRequests, "Too Many Requests", "Too many failed login attempts. Try again later."))
	case errors.Is(err, domain.ErrInvalidCredentials):
		problem.WriteStatus(w, r, problem.New(http.StatusUnauthorized, "Unauthorized", "Email or password is incorrect."))
	default:
		problem.Write(w, r, err, h.logger)
	}
}

func newUserResponse(user domain.User) UserResponse {
	return UserResponse{
		ID:            user.ID.String(),
		Email:         user.Email,
		Name:          user.DisplayName(),
		PictureURL:    user.PictureURL,
		IsAdmin:       user.IsAdmin,
		EmailVerified: user.EmailVerifiedAt != nil,
		CreatedAt:     user.CreatedAt,
	}
}

func (h *AuthHandler) requestMeta(r *http.Request) auth.RequestMeta {
	if h.meta == nil {
		return auth.RequestMeta{}
	}
	return h.meta(r)
}
