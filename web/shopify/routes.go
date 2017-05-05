package shopify

import "github.com/pressly/chi"

func Routes() chi.Router {
	r := chi.NewRouter()

	r.Group(func(r chi.Router) {
		r.Use(CredCtx)
		r.Use(ClientCtx)

		r.Post("/sync", SyncProduct)
	})

	return r
}
