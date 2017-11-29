package category

import (
	"context"
	"net/http"
	"strings"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/data/presenter"
	"github.com/go-chi/chi"
	"github.com/go-chi/render"
	"github.com/pressly/lg"
)

func CategoryTypeCtx(next http.Handler) http.Handler {
	handler := func(w http.ResponseWriter, r *http.Request) {
		rawCategory := strings.TrimSpace(chi.URLParam(r, "categoryType"))
		categoryType := new(data.ProductCategoryType)
		if err := categoryType.UnmarshalText([]byte(rawCategory)); err != nil {
			render.Respond(w, r, err)
			return
		}
		ctx := r.Context()
		ctx = context.WithValue(ctx, "categoryType", *categoryType)
		lg.SetEntryField(ctx, "category", rawCategory)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
	return http.HandlerFunc(handler)
}

func ListCategory(w http.ResponseWriter, r *http.Request) {
	// TODO: with place context?
	render.Respond(w, r, data.ProductCategories)
}

func ListProductCategory(w http.ResponseWriter, r *http.Request) {
	// TODO: with place context?
	ctx := r.Context()
	categoryType := ctx.Value("categoryType").(data.ProductCategoryType)

	productCategories, err := data.DB.ProductCategory.FindByType(categoryType)
	if err != nil {
		render.Respond(w, r, err)
		return
	}
	render.Respond(w, r, presenter.NewCategory(ctx, productCategories))
}
