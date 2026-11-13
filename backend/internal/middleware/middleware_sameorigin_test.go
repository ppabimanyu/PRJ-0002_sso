package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSameOriginAllowsMatchingOrigin(t *testing.T) {
	request := httptest.NewRequest(http.MethodPost, "http://localhost:5173/api/v1/auth/login", nil)
	request.Header.Set("Origin", "http://localhost:5173")
	recorder := httptest.NewRecorder()

	SameOrigin(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})).ServeHTTP(recorder, request)

	assert.Equal(t, http.StatusNoContent, recorder.Code)
}

func TestSameOriginRejectsDifferentOrigin(t *testing.T) {
	request := httptest.NewRequest(http.MethodPost, "http://localhost:5173/api/v1/auth/login", nil)
	request.Header.Set("Origin", "https://attacker.example")
	recorder := httptest.NewRecorder()

	SameOrigin(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})).ServeHTTP(recorder, request)

	assert.Equal(t, http.StatusForbidden, recorder.Code)
}

func TestSameOriginRejectsCrossSiteFetchMetadata(t *testing.T) {
	request := httptest.NewRequest(http.MethodPost, "http://localhost:5173/api/v1/auth/logout", nil)
	request.Header.Set("Sec-Fetch-Site", "cross-site")
	recorder := httptest.NewRecorder()

	SameOrigin(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})).ServeHTTP(recorder, request)

	assert.Equal(t, http.StatusForbidden, recorder.Code)
}

func TestSameOriginAllowsSafeMethodsFromAnywhere(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "http://localhost:5173/api/v1/auth/session", nil)
	request.Header.Set("Origin", "https://attacker.example")
	recorder := httptest.NewRecorder()

	SameOrigin(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})).ServeHTTP(recorder, request)

	assert.Equal(t, http.StatusNoContent, recorder.Code)
}
