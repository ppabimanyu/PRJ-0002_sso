package auth

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHashPasswordVerifiesTheSamePassword(t *testing.T) {
	password := "correct horse battery staple"

	hash, err := HashPassword(password)
	require.NoError(t, err)
	require.True(t, strings.HasPrefix(hash, "$argon2id$"))

	valid, err := VerifyPassword(password, hash)
	require.NoError(t, err)
	assert.True(t, valid)
}

func TestVerifyPasswordRejectsTheWrongPassword(t *testing.T) {
	hash, err := HashPassword("correct horse battery staple")
	require.NoError(t, err)

	valid, err := VerifyPassword("incorrect horse battery staple", hash)
	require.NoError(t, err)
	assert.False(t, valid)
}

func TestHashPasswordUsesAFreshSalt(t *testing.T) {
	first, err := HashPassword("correct horse battery staple")
	require.NoError(t, err)
	second, err := HashPassword("correct horse battery staple")
	require.NoError(t, err)

	assert.NotEqual(t, first, second)
}

func TestVerifyPasswordRejectsAMalformedHash(t *testing.T) {
	valid, err := VerifyPassword("correct horse battery staple", "not-a-phc-hash")

	assert.Error(t, err)
	assert.False(t, valid)
}

func TestValidatePassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
		email    string
		wantErr  string
	}{
		{name: "accepts twelve characters", password: "twelve-chars", email: "user@example.com"},
		{name: "rejects short password", password: "short", email: "user@example.com", wantErr: "at least 12"},
		{name: "rejects long password", password: strings.Repeat("a", 129), email: "user@example.com", wantErr: "at most 128"},
		{name: "rejects password equal to email", password: "user@example.com", email: "User@Example.com", wantErr: "must not match"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := ValidatePassword(test.password, test.email)
			if test.wantErr == "" {
				require.NoError(t, err)
				return
			}
			require.Error(t, err)
			assert.Contains(t, err.Error(), test.wantErr)
		})
	}
}
