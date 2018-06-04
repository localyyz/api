package cart

import "github.com/pkg/errors"

var (
	ErrInvalidEmail    = errors.New("empty or invalid email")
	ErrInvalidShipping = errors.New("empty or invalid shipping address")
	ErrInvalidBilling  = errors.New("empty or invalid billing address")
	ErrInvalidStatus   = errors.New("invalid cart status, already completed.")
)
