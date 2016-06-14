package reward

import (
	"net/http"

	"github.com/pressly/chi"
)

func Routes() http.Handler {
	r := chi.NewRouter()
	return r
}
