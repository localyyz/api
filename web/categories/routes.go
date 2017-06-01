package categories

import "github.com/pressly/chi"

func Routes() chi.Router {
	r := chi.NewRouter()

	//r.Get("/", ListCategories)
	//r.Route("/:category", func(r chi.Router) {
	//r.Use(CategoryCtx)
	//r.Get("/", GetCategory)
	//r.Get("/places", ListPlaces)
	//})

	return r
}
