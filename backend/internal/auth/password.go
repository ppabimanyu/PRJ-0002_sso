// Package auth owns the multi-step authentication flows, Session lifecycle,
// and the Password hashing rules. Handlers call it; it never imports handlers
// or repository implementations.
package auth

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"strings"
	"unicode/utf8"

	"golang.org/x/crypto/argon2"

	"github.com/ppabimanyu/sso/internal/domain"
)

// Password policy from the v1 spec.
const (
	MinPasswordLength = 12
	MaxPasswordLength = 128

	argonMemory      = 64 * 1024
	argonIterations  = 3
	argonParallelism = 2
	argonSaltLength  = 16
	argonKeyLength   = 32
)

// ValidatePassword enforces length and the email comparison rule. It does not
// require composition rules on purpose.
func ValidatePassword(password, email string) error {
	length := utf8.RuneCountInString(password)
	if length < MinPasswordLength {
		return fmt.Errorf("%w: password must be at least %d characters", domain.ErrValidation, MinPasswordLength)
	}
	if length > MaxPasswordLength {
		return fmt.Errorf("%w: password must be at most %d characters", domain.ErrValidation, MaxPasswordLength)
	}
	if strings.EqualFold(strings.TrimSpace(email), password) {
		return fmt.Errorf("%w: password must not match the email address", domain.ErrValidation)
	}
	return nil
}

// HashPassword returns a PHC-formatted Argon2id hash with a fresh random salt.
func HashPassword(password string) (string, error) {
	salt := make([]byte, argonSaltLength)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("generate password salt: %w", err)
	}

	key := argon2.IDKey([]byte(password), salt, argonIterations, argonMemory, argonParallelism, argonKeyLength)

	return fmt.Sprintf(
		"$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version,
		argonMemory,
		argonIterations,
		argonParallelism,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(key),
	), nil
}

// VerifyPassword reports whether password matches a PHC Argon2id hash. A
// malformed hash returns an error so callers can distinguish bad input from a
// bad password.
func VerifyPassword(password, encoded string) (bool, error) {
	params, salt, expected, err := parsePasswordHash(encoded)
	if err != nil {
		return false, err
	}

	actual := argon2.IDKey([]byte(password), salt, params.iterations, params.memory, params.parallelism, uint32(len(expected))) // #nosec G115 -- parsePasswordHash bounds the key length to 16..64 bytes.
	return subtle.ConstantTimeCompare(actual, expected) == 1, nil
}

type argonParams struct {
	memory      uint32
	iterations  uint32
	parallelism uint8
}

func parsePasswordHash(encoded string) (argonParams, []byte, []byte, error) {
	parts := strings.Split(encoded, "$")
	if len(parts) != 6 || parts[1] != "argon2id" {
		return argonParams{}, nil, nil, fmt.Errorf("password hash is not argon2id")
	}

	var version int
	if _, err := fmt.Sscanf(parts[2], "v=%d", &version); err != nil || version != argon2.Version {
		return argonParams{}, nil, nil, fmt.Errorf("password hash version is not supported")
	}

	var memory, iterations uint32
	var parallelism uint8
	if _, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &memory, &iterations, &parallelism); err != nil {
		return argonParams{}, nil, nil, fmt.Errorf("password hash parameters are malformed: %w", err)
	}
	if memory == 0 || iterations == 0 || parallelism == 0 || memory > 1024*1024 || iterations > 10 {
		return argonParams{}, nil, nil, fmt.Errorf("password hash parameters are out of range")
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return argonParams{}, nil, nil, fmt.Errorf("password salt is malformed: %w", err)
	}
	expected, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return argonParams{}, nil, nil, fmt.Errorf("password hash is malformed: %w", err)
	}
	if len(expected) < 16 || len(expected) > 64 {
		return argonParams{}, nil, nil, fmt.Errorf("password hash length %d is out of range", len(expected))
	}

	return argonParams{memory: memory, iterations: iterations, parallelism: parallelism}, salt, expected, nil
}
