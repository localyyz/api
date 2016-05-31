// Errors hides any internal error to a more friendly external error
// details are shown or hidden depending on the debug level

package ws

import (
	"errors"

	"github.com/goware/errorx"
)

var (
	ErrUnrechable = errors.New("unreachable")
)

// WrapError hides or shows error details depending on the log level
//  it also exchanges an internal error for an external friendly message
//
// ie:
//	InternalError: pg error no column 'name' found on table 'accounts'
//  ExternalErorr: code 2001: failed to update
//
//  returns wrapped errorx and proper http status code
func WrapError(err error) error {
	// TODO: mapping with proper status code
	return errorx.New(1000, err.Error())
}
