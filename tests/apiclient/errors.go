package apiclient

import (
	"encoding/json"
	"fmt"
	"net/http"

	"bitbucket.org/moodie-app/moodie-api/web/api"
)

type Err struct {
	Method     string
	RequestURL string
	StatusCode int
	Message    string
}

func (e Err) Error() string {
	return fmt.Sprintf("%v %v: HTTP %v: %s", e.Method, e.RequestURL, e.StatusCode, e.Message)
}

type Err400 struct{ Err }
type Err401 struct{ Err }
type Err403 struct{ Err }
type Err404 struct{ Err }
type Err409 struct{ Err }
type Err410 struct{ Err }
type Err422 struct{ Err }
type Err429 struct{ Err }
type Err500 struct{ Err }
type Err503 struct{ Err }
type ErrUnknown struct{ Err }

func NewError(resp *http.Response) error {
	var errMsg api.ApiError
	json.NewDecoder(resp.Body).Decode(&errMsg)
	resp.Body.Close()

	req := resp.Request
	errWrap := Err{
		req.Method,
		req.URL.String(),
		resp.StatusCode,
		errMsg.ErrorText,
	}

	switch resp.StatusCode {
	case 400:
		return Err400{errWrap}
	case 401:
		return Err401{errWrap}
	case 403:
		return Err403{errWrap}
	case 404:
		return Err404{errWrap}
	case 409:
		return Err409{errWrap}
	case 410:
		return Err410{errWrap}
	case 422:
		return Err422{errWrap}
	case 429:
		return Err429{errWrap}
	case 500:
		return Err500{errWrap}
	case 503:
		return Err503{errWrap}
	default:
		return ErrUnknown{errWrap}
	}
	panic("unreachable")
}
