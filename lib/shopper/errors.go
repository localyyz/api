package shopper

import "github.com/pkg/errors"

var (
	ErrShippingRates  = errors.New("merchant does not ship to your address")
	ErrEmptyCheckout  = errors.New("nothing in your cart")
	ErrItemOutofStock = errors.New("item is out of stock")

	ErrEmptyPaymentTransaction = errors.New("payment missing transaction")
	ErrDiscountCode            = errors.New("discount code cannot be applied")
)

type CheckoutError struct {
	PlaceID int64
	ItemID  int64
	ErrCode CheckoutErrorCode
	Err     error
}

type CheckoutErrorCode uint32

const (
	_                                CheckoutErrorCode = iota // 0
	CheckoutErrorCodeGeneric                                  // 1
	CheckoutErrorCodeLineItem                                 // 2
	CheckoutErrorCodeNoShipping                               // 3
	CheckoutErrorCodeShippingAddress                          // 4
	CheckoutErrorCodeBillingAddress                           // 5
	CheckoutErrorCodeDiscountCode                             // 6
)

func (e *CheckoutError) Error() string {
	return e.Err.Error()
}
