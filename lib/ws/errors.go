// Errors hides any internal error to a more friendly external error
// details are shown or hidden depending on the debug level

package ws

import (
	"errors"
	"net/http"

	"upper.io/db"

	"github.com/goware/errorx"
)

var (
	ErrUnrechable = errors.New("unreachable")

	mappings = map[error]*errorx.Errorx{
		db.ErrNoMoreRows: errorx.New(http.StatusNotFound, "ops... we couldn't find what you were looking for"),
	}
)

// WrapError hides or shows error details depending on the log level
//  it also exchanges an internal error for an external friendly message
//
// ie:
//	InternalError: pg error no column 'name' found on table 'accounts'
//  ExternalErorr: code 2001: failed to update
//
//  returns wrapped errorx and proper http status code
func WrapError(status int, err error) (int, error) {

	// Look up error mapping
	if e, ok := mappings[err]; ok {
		e.Wrap(err)
		return e.Code, e
	}

	return status, errorx.New(1000, err.Error())
}
