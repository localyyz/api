package api

import (
	"errors"
	"net/http"

	db "upper.io/db.v3"

	"github.com/goware/lg"
	"github.com/pressly/chi/render"
)

type ApiError struct {
	Err error `json:"-"`

	StatusCode int    `json:"-"`
	StatusText string `json:"status"`

	AppCode   int64  `json:"code,omitempty"`
	ErrorText string `json:"error,omitempty"`
}

var (
	ErrBadID          = &ApiError{StatusCode: http.StatusBadRequest, ErrorText: "bad or invalid id"}
	ErrBadAction      = errors.New("can't do this")
	ErrInvalidSession = &ApiError{StatusCode: http.StatusUnauthorized, ErrorText: "invalid session"}

	/** Auth **/
	// Password and Confirmation must match
	ErrPasswordMismatch = errors.New("password mismatch")
	// Minimum length requirement for password
	ErrPasswordLength = errors.New("password must be at least 8 characters long")
	// Unknown encryption error
	ErrEncryptinError = errors.New("internal error")
	// Invalid login
	ErrInvalidLogin = errors.New("invalid login credentials, check username and/or password")

	/* Shopify */
	ErrConflictStore = &ApiError{StatusCode: http.StatusConflict, ErrorText: "store already connected"}

	/* Permission */
	ErrPermissionDenied = &ApiError{StatusCode: http.StatusUnauthorized}

	/* Claims */
	ErrClaimDistance = errors.New("claim distance")

	// Cart
	ErrEmptyCart = &ApiError{StatusCode: http.StatusOK, ErrorText: "empty cart"}

	// generic api error
	errGeneric = &ApiError{StatusCode: http.StatusInternalServerError, ErrorText: "Something went wrong"}

	// error mapping
	mappings = map[error]*ApiError{
		db.ErrNoMoreRows: &ApiError{StatusCode: http.StatusNotFound, ErrorText: "Ops... We couldn't find what you were looking for."},

		ErrPasswordLength:   &ApiError{StatusCode: http.StatusBadRequest},
		ErrPasswordMismatch: &ApiError{StatusCode: http.StatusUnauthorized},
		ErrInvalidLogin:     &ApiError{StatusCode: http.StatusUnauthorized},

		ErrEncryptinError: &ApiError{StatusCode: http.StatusInternalServerError},
		ErrClaimDistance:  &ApiError{StatusCode: http.StatusBadRequest, ErrorText: "You're too far away. Get closer!"},
	}
)

func ErrInvalidRequest(err error) *ApiError {
	return &ApiError{
		Err:        err,
		StatusCode: http.StatusBadRequest,
		StatusText: "Invalid request.",
		ErrorText:  err.Error(),
	}
}

func WrapErr(err error) *ApiError {
	if e, ok := mappings[err]; ok {
		a := *e
		if len(a.StatusText) == 0 {
			a.StatusText = http.StatusText(a.StatusCode)
		}
		if len(a.ErrorText) == 0 {
			a.ErrorText = err.Error()
		}
		return &a
	}

	lg.Errorf("encountered internal error: %+v", err)
	return errGeneric
}

func (e *ApiError) Render(w http.ResponseWriter, r *http.Request) error {
	render.Status(r, e.StatusCode)
	return nil
}
