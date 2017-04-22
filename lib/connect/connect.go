package connect

import "errors"

var (
	ErrTokenExpired = errors.New("token expired")
	ErrInvalidState = errors.New("state invalid")
	ErrInvalidToken = errors.New("token invalid")
)
