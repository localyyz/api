package locale

import (
	"net/http"

	"github.com/pressly/chi"
)

func Routes() http.Handler {
	r := chi.NewRouter()

	r.Get("/", ListLocales)

	return r

}
