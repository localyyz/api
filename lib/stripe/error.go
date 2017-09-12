package stripe

import (
	"encoding/json"
	"net/http"
)

type ErrorResponse struct {
	Errors map[string]interface{} `json:"error"` // more detail on individual errors
}

func (r *ErrorResponse) Error() string {
	e, ok := r.Errors["message"].(string)
	if !ok {
		return "unknown error"
	}
	return e
}

func CheckResponse(res *http.Response) error {
	if res.StatusCode <= http.StatusBadRequest {
		return nil
	}
	errorResponse := &ErrorResponse{}
	json.NewDecoder(res.Body).Decode(errorResponse)

	return errorResponse
}
