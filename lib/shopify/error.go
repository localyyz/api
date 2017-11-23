package shopify

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
)

// Shopify errors usually have the form:
// {
//   "errors": {
//     "title": [
//       "something is wrong"
//     ]
//   }
// }
//

type ShopifyErrorer interface {
	Type() string
}

type LineItemError struct {
	Position string `json:"postition"`
	Quantity []struct {
		Message string `json:"message"`
		Options struct {
			Remaining int `json:"remaining"`
		} `json:"options"`
		Code string `json:"code"`
	} `json:"quantity"`

	ShopifyErrorer
}

type DiscountCodeError struct {
	Reason interface{} `json:"reason"`
	ShopifyErrorer
}

type ErrorResponse struct {
	Errors interface{} `json:"errors"`
}

var (
	// TODO: make this an unmarshall type
	ErrNotEnoughInStock = `not_enough_in_stock`
)

func (e *LineItemError) Error() string {
	for _, q := range e.Quantity {
		return fmt.Sprintf("[line_item] pos(%s) %s %s", e.Position, q.Code, q.Message)
	}
	return fmt.Sprintf("line_item at pos(%s) has errors", e.Position)
}

func (e *LineItemError) Type() string {
	return `line_items`
}

func (e *DiscountCodeError) Error() string {
	return fmt.Sprintf("%+v", e.Reason)
}

func (e *DiscountCodeError) Type() string {
	return `discount_code`
}

func (r *ErrorResponse) Error() string {
	if e, ok := r.Errors.(map[string]interface{}); ok {
		for k, v := range e {
			// value here can be a slice
			return fmt.Sprintf("%s: %+v", k, v)
		}
	}
	if e, ok := r.Errors.(string); ok {
		return e
	}
	return "unknown, unparsed error"
}

// CheckResponse checks the API response for errors, and returns them if
// present. A response is considered an error if it has a status code outside
// the 200 range or equal to 202 Accepted.
// API error responses are expected to have either no response
// body, or a JSON response body that maps to ErrorResponse. Any other
// response body will be silently ignored.
func CheckResponse(r *http.Response) error {
	if r.StatusCode == http.StatusAccepted {
		return nil
	}
	if r.StatusCode == http.StatusForbidden {
		return errors.New("forbidden")
	}
	if c := r.StatusCode; 200 <= c && c <= 299 {
		return nil
	}
	errorResponse := &ErrorResponse{}
	data, err := ioutil.ReadAll(r.Body)
	if err == nil && data != nil {
		json.Unmarshal(data, errorResponse)
	}
	return findFirstError(errorResponse)
}

func findFirstError(r *ErrorResponse) error {
	rr, ok := r.Errors.(map[string]interface{})
	if !ok {
		return r
	}

	// find the first error, and return
	for k, v := range rr {
		vv, ok := v.(map[string]interface{})
		if !ok {
			continue
		}

		switch k {
		//shipping_line: map[id:[map[code:expired message:has expired options:map[]]]]
		case "line_items":
			for pos, vvv := range vv {
				b, _ := json.Marshal(vvv)
				var e *LineItemError
				json.Unmarshal(b, &e)
				e.Position = pos
				return e
			}
		case "checkout":
			for kk, vvv := range vv {
				switch kk {
				case "discount_code":
					// list of fields
					if e, _ := vvv.([]interface{}); e != nil {
						for _, ee := range e {
							if ex, _ := ee.(map[string]interface{}); ex != nil {
								for _, reason := range ex {
									return &DiscountCodeError{Reason: reason}
								}
							}
						}
					}
				}
			}
		}
	}

	return r
}
