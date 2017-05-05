package connect

import "errors"

var (
	ErrTokenExpired = errors.New("token expired")
	ErrInvalidState = errors.New("state invalid")
	ErrMismathShop  = errors.New("shop id mismatch")
	ErrInvalidToken = errors.New("token invalid")
)
