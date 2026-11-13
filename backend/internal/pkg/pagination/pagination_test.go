package pagination

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ppabimanyu/sso/internal/domain"
)

func TestParseDefaults(t *testing.T) {
	params, err := Parse("", "")
	require.NoError(t, err)
	assert.Equal(t, Params{Limit: DefaultLimit, Offset: 0}, params)
}

func TestParseValues(t *testing.T) {
	params, err := Parse("50", "100")
	require.NoError(t, err)
	assert.Equal(t, Params{Limit: 50, Offset: 100}, params)
}

func TestParseRejectsOutOfRangeLimit(t *testing.T) {
	_, err := Parse("0", "")
	require.ErrorIs(t, err, domain.ErrValidation)

	_, err = Parse("101", "")
	require.ErrorIs(t, err, domain.ErrValidation)
}

func TestParseRejectsNegativeOffset(t *testing.T) {
	_, err := Parse("", "-1")
	require.ErrorIs(t, err, domain.ErrValidation)
}
