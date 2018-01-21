package product

import (
	"context"
	"net/http"
	"strings"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/web/api"
	"github.com/go-chi/render"
	db "upper.io/db.v3"
)

func CategoryCtx(next http.Handler) http.Handler {
	handler := func(w http.ResponseWriter, r *http.Request) {
		u := r.URL.Query()

		var cond db.Cond
		if category := strings.ToLower(u.Get("category")); category != "" {
			// find the value => mapping in product categories
			var values []string
			mappings, _ := data.DB.Category.FindByMapping(category)
			for _, m := range mappings {
				values = append(values, m.Value)
			}
			cond = db.Cond{
				"pt.type":  data.ProductTagTypeCategory,
				"pt.value": values,
			}
		} else if rawType := strings.ToLower(u.Get("categoryType")); rawType != "" {
			categoryType := new(data.CategoryType)
			if err := categoryType.UnmarshalText([]byte(rawType)); err != nil {
				render.Respond(w, r, api.ErrInvalidRequest(err))
				return
			}

			// find the value => mapping in product categories
			var values []string
			categories, _ := data.DB.Category.FindByType(*categoryType)
			for _, c := range categories {
				values = append(values, c.Value)
			}
			cond = db.Cond{
				"pt.type":  data.ProductTagTypeCategory,
				"pt.value": values,
			}
		}

		ctx := r.Context()
		if len(cond) > 0 {
			ctx = context.WithValue(ctx, "product.filter", cond)
		}
		next.ServeHTTP(w, r.WithContext(ctx))
	}

	return http.HandlerFunc(handler)
}
