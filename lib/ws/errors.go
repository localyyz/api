// Errors hides any internal error to a more friendly external error
// details are shown or hidden depending on the debug level

package ws

import (
	"errors"
	"net/http"

	"bitbucket.org/moodie-app/moodie-api/web/api"
	"github.com/goware/errorx"
	"upper.io/db.v3"
)

var (
	ErrUnrechable = errors.New("unreachable")

	mappings = map[error]*errorx.Errorx{
		db.ErrNoMoreRows:     errorx.New(http.StatusNotFound, "Ops... We couldn't find what you were looking for."),
		api.ErrClaimDistance: errorx.New(http.StatusBadRequest, "You're too far away. Get closer!"),
	}
)

// WrapError hides or shows error details depending on the log level
//  it also exchanges an internal error for an external friendly message
//
// ie:
//	InternalError: pg error no column 'name' found on table 'accounts'
//  ExternalError: code 2001: failed to update
//
//  returns wrapped errorx and proper http status code
func WrapError(status int, err error) (int, error) {

	// Mapped errors are well defined situations that
	// should provide a predefined helpful message to the user.
	if e, ok := mappings[err]; ok {
		e.Wrap(err)
		return e.Code, errors.New(e.Message)
	}

	return status, errorx.New(1000, err.Error())
}
