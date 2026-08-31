// Package validate provides a single shared validator instance used to
// check Options structs (struct-tagged with `validate:"..."`) across the
// codebase.
package validate

import (
	"fmt"

	"github.com/go-playground/validator/v10"
)

var v = validator.New(validator.WithRequiredStructEnabled())

// Struct validates s against its `validate` struct tags.
func Struct(s interface{}) error {
	if err := v.Struct(s); err != nil {
		return fmt.Errorf("validate: %w", err)
	}
	return nil
}
