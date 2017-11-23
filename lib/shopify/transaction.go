package shopify

import (
	"time"
)

type Transaction struct {
	ID        int64  `json:"id"`
	Amount    string `json:"amount"`
	OrderID   int64  `json:"order_id"`
	ErrorCode string `json:"error_code"`
	Status    string `json:"status"`

	Test     bool   `json:"test"`
	Currency string `json:"currency"`

	CreatedAt *time.Time `json:"created_at"`
}

// List of error codes
//incorrect_number
//invalid_number
//invalid_expiry_date
//invalid_cvc
//expired_card
//incorrect_cvc
//incorrect_zip
//incorrect_address
//card_declined
//processing_error
//call_issuer
//pick_up_card
