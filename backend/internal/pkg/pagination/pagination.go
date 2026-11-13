// Package pagination holds the list response shape and query parsing that
// feature handlers share.
package pagination

import (
	"fmt"
	"strconv"

	"github.com/ppabimanyu/sso/internal/domain"
)

// Defaults for a list request that omits limit and offset.
const (
	DefaultLimit = 20
	MaxLimit     = 100
)

// Params are the validated limit and offset for one list request.
type Params struct {
	Limit  int
	Offset int
}

// Response is the standard list envelope.
type Response[T any] struct {
	Items  []T   `json:"items"`
	Total  int64 `json:"total"`
	Limit  int   `json:"limit"`
	Offset int   `json:"offset"`
}

// Parse reads limit and offset query values. Empty values fall back to the
// defaults. Out of range values return domain.ErrValidation.
func Parse(limitValue, offsetValue string) (Params, error) {
	params := Params{Limit: DefaultLimit, Offset: 0}

	if limitValue != "" {
		limit, err := strconv.Atoi(limitValue)
		if err != nil || limit < 1 || limit > MaxLimit {
			return Params{}, fmt.Errorf("%w: limit must be a number from 1 to %d", domain.ErrValidation, MaxLimit)
		}
		params.Limit = limit
	}

	if offsetValue != "" {
		offset, err := strconv.Atoi(offsetValue)
		if err != nil || offset < 0 {
			return Params{}, fmt.Errorf("%w: offset must be a number of zero or more", domain.ErrValidation)
		}
		params.Offset = offset
	}

	return params, nil
}
