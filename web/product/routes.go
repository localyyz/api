package product

import "github.com/pressly/chi"

func Routes() chi.Router {
	r := chi.NewRouter()

	r.Route("/:productID", func(r chi.Router) {
		r.Use(ProductCtx)
		r.Get("/variant", GetVariant)
	})

	return r
}
