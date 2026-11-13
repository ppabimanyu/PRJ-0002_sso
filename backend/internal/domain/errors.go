// Package domain holds entities, constants, and the ports that handlers
// depend on. It imports no other internal package.
package domain

import "errors"

// Sentinel errors that handlers and repositories use to describe a failed
// operation without leaking infrastructure details.
var (
	ErrNotFound   = errors.New("not found")
	ErrConflict   = errors.New("conflict")
	ErrValidation = errors.New("validation failed")
)
