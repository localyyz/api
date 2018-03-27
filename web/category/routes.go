package category

import (
	"net/http"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/data/presenter"
	"github.com/go-chi/chi"
	"github.com/go-chi/render"
	db "upper.io/db.v3"
)

func Routes() chi.Router {
	r := chi.NewRouter()

	return r
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
