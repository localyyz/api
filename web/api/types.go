package api

import (
	"net/http"

	"github.com/pressly/chi/render"
)

type emptyList []render.Renderer
type noContent struct{}

// empty list is used by endpoints
// to return a valid json empty list
var (
	EmptyListResp = emptyList{}
	NoContentResp = noContent{}
)

func (emptyList) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

func (noContent) Render(w http.ResponseWriter, r *http.Request) error {
	render.Status(r, http.StatusNoContent)
	return nil
}
