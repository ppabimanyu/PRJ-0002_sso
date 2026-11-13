// Package validate turns struct validation failures into domain.ErrValidation.
package validate

import (
	"fmt"

	"github.com/go-playground/validator/v10"

	"github.com/ppabimanyu/sso/internal/domain"
)

var instance = validator.New(validator.WithRequiredStructEnabled())

// Struct validates a request DTO or any struct with `validate` tags.
func Struct(value any) error {
	if err := instance.Struct(value); err != nil {
		return fmt.Errorf("%w: %w", domain.ErrValidation, err)
	}
	return nil
}
