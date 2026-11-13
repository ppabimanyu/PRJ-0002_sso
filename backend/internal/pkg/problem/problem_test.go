package problem

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ppabimanyu/sso/internal/domain"
)

func TestFromErrorMapsDomainErrors(t *testing.T) {
	tests := []struct {
		name   string
		err    error
		status int
	}{
		{name: "not found", err: domain.ErrNotFound, status: http.StatusNotFound},
		{name: "conflict", err: fmt.Errorf("%w: duplicate email", domain.ErrConflict), status: http.StatusConflict},
		{name: "validation", err: fmt.Errorf("%w: email is required", domain.ErrValidation), status: http.StatusBadRequest},
		{name: "unknown", err: fmt.Errorf("connection refused"), status: http.StatusInternalServerError},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.status, FromError(test.err).Status)
		})
	}
}

func TestFromErrorHidesInternalDetail(t *testing.T) {
	p := FromError(fmt.Errorf("password=hunter2"))

	assert.Equal(t, http.StatusInternalServerError, p.Status)
	assert.Empty(t, p.Detail)
}

func TestFromErrorKeepsValidationDetail(t *testing.T) {
	p := FromError(fmt.Errorf("%w: email is required", domain.ErrValidation))

	assert.Equal(t, "email is required", p.Detail)
}

func TestWriteStatusSetsContentTypeAndInstance(t *testing.T) {
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/widgets", nil)

	WriteStatus(recorder, request, New(http.StatusNotFound, "Not Found", ""))

	require.Equal(t, http.StatusNotFound, recorder.Code)
	assert.Equal(t, ContentType, recorder.Header().Get("Content-Type"))
	assert.Contains(t, recorder.Body.String(), `"instance":"/api/v1/widgets"`)
}
