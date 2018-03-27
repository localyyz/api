package product

import (
	"context"
	"net/http"
	"strings"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/web/api"
	"github.com/go-chi/chi"
	"github.com/go-chi/render"
	db "upper.io/db.v3"
)

func CategoryCtx(next http.Handler) http.Handler {
	handler := func(w http.ResponseWriter, r *http.Request) {
		u := r.URL.Query()

		var cond db.Cond
		if category := strings.ToLower(u.Get("category")); category != "" {
			// selected a mapping value -> find all the values that it maps to
			var values []string
			mappings, _ := data.DB.Category.FindByMapping(category)
			for _, m := range mappings {
				values = append(values, m.Value)
			}
			cond = db.Cond{db.Raw("products.category->>'value'"): values}
		} else if rawType := strings.ToLower(u.Get("categoryType")); rawType != "" {
			categoryType := new(data.CategoryType)
			if err := categoryType.UnmarshalText([]byte(rawType)); err != nil {
				render.Respond(w, r, api.ErrInvalidRequest(err))
				return
			}
			cond = db.Cond{db.Raw("products.category->>'type'"): categoryType.String()}
		}

		ctx := r.Context()
		if len(cond) > 0 {
			ctx = context.WithValue(ctx, "product.filter", cond)
		}
		next.ServeHTTP(w, r.WithContext(ctx))
	}
	return http.HandlerFunc(handler)
}

func Routes() chi.Router {
	r := chi.NewRouter()

	r.Get("/featured", ListFeaturedProduct)
	r.Get("/onsale", ListOnsaleProduct)
	r.Route("/{productID}", func(r chi.Router) {
		r.Use(ProductCtx)
		r.Get("/", GetProduct)
		r.Get("/variant", GetVariant)
		r.Get("/related", ListRelatedProduct)
	})

	return r
}
