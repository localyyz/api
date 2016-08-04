package utils

import "errors"

var (
	ErrBadID          = errors.New("bad or invalid id")
	ErrBadAction      = errors.New("can't do this")
	ErrInvalidSession = errors.New("linvalid session")
)
