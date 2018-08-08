package api

import (
	"errors"
	"net/http"

	"bitbucket.org/moodie-app/moodie-api/lib/shopify"

	db "upper.io/db.v3"

	"github.com/go-chi/render"
	"github.com/lib/pq"
	"github.com/pressly/lg"
)

type ApiError struct {
	Err error `json:"-"`

	StatusCode int    `json:"-"`
	StatusText string `json:"status"`

	AppCode   int64       `json:"code,omitempty"`
	ErrorText string      `json:"error,omitempty"`
	Cause     string      `json:"cause,omitempty"`
	Data      interface{} `json:"data,omitempty"`
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
	ErrConflictStore      = &ApiError{StatusCode: http.StatusConflict, ErrorText: "store already connected"}
	ErrInvalidChargeID    = &ApiError{StatusCode: http.StatusBadRequest, ErrorText: "invalid id was given by shopify, please contact support"}
	ErrInvalidBillingPlan = &ApiError{StatusCode: http.StatusBadRequest, ErrorText: "invalid plan type was selected, please contact support"}

	/* Permission */
	ErrPermissionDenied = &ApiError{StatusCode: http.StatusUnauthorized}

	/* Claims */
	ErrClaimDistance = errors.New("claim distance")

	/* Cart */
	ErrEmptyCart           = &ApiError{StatusCode: http.StatusOK, ErrorText: "empty cart"}
	ErrInvalidDiscountCode = &ApiError{StatusCode: http.StatusBadRequest, ErrorText: "discount code is invalid"}
	errOutOfStockCart      = &ApiError{StatusCode: http.StatusBadRequest, StatusText: "out-of-stock", ErrorText: "one or more items in your cart are out of stock"}
	errOutOfStockAdd       = &ApiError{StatusCode: http.StatusBadRequest, StatusText: "out-of-stock", ErrorText: "this variant is out of stock."}

	/* Lightning Section */
	ErrExpiredDeal         = &ApiError{StatusCode: http.StatusBadRequest, StatusText: "invalid interval", ErrorText: "this lightning deal has expired or is not available yet"}
	ErrLightningOutOfStock = &ApiError{StatusCode: http.StatusBadRequest, StatusText: "out of stock", ErrorText: "the products from this lightning collection have been sold out"}
	ErrMultiplePurchase    = &ApiError{StatusCode: http.StatusBadRequest, StatusText: "already purchased", ErrorText: "you have already purchased today's deal"}

	// generic api error
	errGeneric  = &ApiError{StatusCode: http.StatusInternalServerError, ErrorText: "Something went wrong"}
	errDatabase = &ApiError{StatusCode: http.StatusInternalServerError, ErrorText: "Failed. Please try again."}

	// error mapping
	mappings = map[error]*ApiError{
		db.ErrNoMoreRows: &ApiError{StatusCode: http.StatusNotFound, ErrorText: "Ops... We couldn't find what you were looking for."},

		ErrPasswordLength:   &ApiError{StatusCode: http.StatusBadRequest},
		ErrPasswordMismatch: &ApiError{StatusCode: http.StatusUnauthorized},
		ErrInvalidLogin:     &ApiError{StatusCode: http.StatusUnauthorized},

		ErrEncryptinError: &ApiError{StatusCode: http.StatusInternalServerError},
		ErrClaimDistance:  &ApiError{StatusCode: http.StatusBadRequest, ErrorText: "You're too far away. Get closer!"},
	}

	// db error mappings
	dbmappings = map[string]*ApiError{
		"unique_violation": &ApiError{StatusCode: http.StatusBadRequest, ErrorText: "Already exists!"},
	}
)

func ErrStripeProcess(err error) *ApiError {
	return &ApiError{
		Err:        err,
		StatusCode: http.StatusBadRequest,
		StatusText: "Payment error.",
		ErrorText:  err.Error(),
	}
}

func ErrCardVaultProcess(err error) *ApiError {
	return &ApiError{
		Err:        err,
		StatusCode: http.StatusBadRequest,
		StatusText: "Payment error.",
		ErrorText:  err.Error(),
	}
}

func ErrInvalidRequest(err error) *ApiError {
	return &ApiError{
		Err:        err,
		StatusCode: http.StatusBadRequest,
		StatusText: "Invalid request.",
		ErrorText:  err.Error(),
	}
}

func ErrOutOfStockCart(v interface{}) *ApiError {
	e := *errOutOfStockCart
	e.Data = v
	return &e
}

func ErrOutOfStockAdd(v interface{}) *ApiError {
	e := *errOutOfStockAdd
	e.Data = v
	return &e
}

func WrapErr(err error) *ApiError {
	if e, ok := err.(*ApiError); ok {
		lg.Errorf("encountered invalid request error: %s", e)
		return e
	}

	if e, ok := err.(shopify.ShopifyErrorer); ok {
		lg.Errorf("encountered shopify api error: %s", e)
		return errGeneric
	}

	if e, ok := err.(*pq.Error); ok {
		lg.Errorf("encountered database error: %s", e)

		a := *errDatabase
		if de, ok := dbmappings[e.Code.Name()]; ok {
			a = *de
		}

		a.Cause = e.Message
		a.StatusText = http.StatusText(a.StatusCode)
		return &a
	}

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

	lg.Errorf("encountered internal error: %s", err.Error())
	return errGeneric
}

func (e *ApiError) Error() string {
	return e.ErrorText
}

func (e *ApiError) Render(w http.ResponseWriter, r *http.Request) error {
	render.Status(r, e.StatusCode)
	return nil
}
