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

func GenderCtx(next http.Handler) http.Handler {
	handler := func(w http.ResponseWriter, r *http.Request) {
		genderRaw := chi.URLParam(r, "gender")

		gender := new(data.ProductGender)
		if err := gender.UnmarshalText([]byte(genderRaw)); err != nil {
			render.Respond(w, r, api.ErrInvalidRequest(err))
			return
		}

		ctx := r.Context()
		ctx = context.WithValue(ctx, "product.gender", *gender)
		next.ServeHTTP(w, r.WithContext(ctx))
	}

	return http.HandlerFunc(handler)
}

func ExportProducts(w http.ResponseWriter, r *http.Request) {
	type exportProduct struct {
		ID            int64   `db:"id" json:"id"`
		Title         string  `db:"title" json:"title"`
		Merchant      string  `db:"merchant" json:"merchant"`
		ImageURL      *string `db:"image_url" json:"image_url"`
		Category      *string `db:"category" json:"category"`
		Price         float64 `db:"price" json:"price"`
		PreviousPrice float64 `db:"previous" json:"previous_price"`
	}

	query := data.DB.Select(
		"p.id",
		"p.title",
		"p.image_url",
		"pl.name as merchant",
		db.Raw("p.category->>'value' as category"),
		db.Raw("coalesce(max(price), 0) as price"),
		db.Raw("coalesce(max(prev_price), 0) previous")).
		From("products p").
		LeftJoin("places pl").On("pl.id = p.place_id").
		LeftJoin("product_variants pv").On("pv.product_id = p.id").
		GroupBy("p.id", "pl.name").
		OrderBy("id DESC")

	var products []*exportProduct
	if err := query.All(&products); err != nil {
		render.Respond(w, r, err)
		return
	}

	render.Respond(w, r, products)
}

func Routes() chi.Router {
	r := chi.NewRouter()

	r.Get("/export", ExportProducts)

	r.Get("/recent", ListRecentProduct)
	r.Get("/featured", ListFeaturedProduct)
	r.With(GenderCtx).
		With(CategoryCtx).
		Get("/gender/{gender}", ListGenderProduct)
	r.Route("/{productID}", func(r chi.Router) {
		r.Use(ProductCtx)
		r.Get("/", GetProduct)
		r.Get("/variant", GetVariant)
		r.Get("/related", ListRelatedProduct)
	})

	return r
}
