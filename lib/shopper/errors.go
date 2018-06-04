package shopper

import "github.com/pkg/errors"

var (
	ErrShippingRates  = errors.New("merchant does not ship to your address")
	ErrEmptyCheckout  = errors.New("nothing in your cart")
	ErrItemOutofStock = errors.New("item is out of stock")

	ErrEmptyPaymentTransaction = errors.New("payment missing transaction")
)

type CheckoutError struct {
	PlaceID int64
	ItemID  int64
	ErrCode CheckoutErrorCode
	Err     error
}

type CheckoutErrorCode uint32

const (
	_ CheckoutErrorCode = iota
	CheckoutErrorCodeGeneric
	CheckoutErrorCodeLineItem
	CheckoutErrorCodeNoShipping
	CheckoutErrorCodeShippingAddress
)

func (e *CheckoutError) Error() string {
	return e.Err.Error()
}
