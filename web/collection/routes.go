package collection

import (
	"github.com/go-chi/chi"
)

func Routes() chi.Router {
	r := chi.NewRouter()

	r.With(FeaturedScopeCtx).Get("/featured", ListCollection)
	r.With(MaleScopeCtx).Get("/man", ListCollection)
	r.With(FemaleScopeCtx).Get("/woman", ListCollection)
	r.Route("/{collectionID}", func(r chi.Router) {
		r.Use(CollectionCtx)

		r.Get("/", GetCollection)
		r.Get("/products", ListProduct)
	})

	return r
}
