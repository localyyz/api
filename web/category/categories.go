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
	db "upper.io/db.v3"
)

func CategoryTypeCtx(next http.Handler) http.Handler {
	handler := func(w http.ResponseWriter, r *http.Request) {
		rawCategory := strings.TrimSpace(chi.URLParam(r, "categoryType"))
		categoryType := new(data.CategoryType)
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
	ctx := r.Context()
	gender := ctx.Value("product.gender").(data.ProductGender)

	var categories []*data.Category
	err := data.DB.Category.
		Find(db.Cond{
			"gender": []data.ProductGender{
				gender,
				data.ProductGenderUnisex,
			},
		}).
		Select(db.Raw("distinct type")).
		OrderBy("type").
		All(&categories)
	if err != nil {
		render.Respond(w, r, err)
		return
	}
	types := make([]data.CategoryType, len(categories))
	for i, c := range categories {
		types[i] = c.Type
	}
	render.Respond(w, r, presenter.NewCategoryList(ctx, types))
}

func GetCategory(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	categoryType := ctx.Value("categoryType").(data.CategoryType)
	render.Respond(w, r, presenter.NewCategory(ctx, categoryType))
}
