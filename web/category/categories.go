package category

import (
	"context"
	"net/http"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/data/presenter"
	"bitbucket.org/moodie-app/moodie-api/web/api"
	"github.com/go-chi/chi"
	"github.com/go-chi/render"
	"github.com/pressly/lg"
	db "upper.io/db.v3"
)

func CategoryTypeCtx(next http.Handler) http.Handler {
	handler := func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		input := chi.URLParam(r, "categoryType")
		var categoryType data.CategoryType
		if err := categoryType.UnmarshalText([]byte(input)); err != nil {
			// did not parse as category. attempt to parse as subcategory
			categories, err := data.DB.Category.FindByMapping(input)
			if err != nil {
				// did not parse either. return error
				render.Respond(w, r, api.ErrInvalidRequest(err))
				return
			}

			var values []string
			for _, c := range categories {
				values = append(values, c.Value)
			}
			ctx = context.WithValue(ctx, "category.value", values)
			lg.SetEntryField(ctx, "subcategory", input)
		} else {
			ctx = context.WithValue(ctx, "category.type", categoryType)
			lg.SetEntryField(ctx, "category", input)
		}

		next.ServeHTTP(w, r.WithContext(ctx))
	}
	return http.HandlerFunc(handler)
}

var (
	displayCategories = []data.CategoryType{
		data.CategorySale, // sale
		data.CategoryCollection,
		data.CategoryApparel,
		data.CategoryHandbag,
		data.CategoryShoe,
		data.CategoryJewelry,
		data.CategoryAccessory,
		data.CategoryBag,
		data.CategoryCosmetic,
		data.CategorySneaker,
		data.CategorySwimwear,
	}
)

func List(w http.ResponseWriter, r *http.Request) {
	presented := presenter.NewCategoryList(r.Context(), displayCategories)
	render.RenderList(w, r, presented)
}

func GetCategory(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	categoryType := ctx.Value("category.type").(data.CategoryType)
	render.Render(w, r, presenter.NewCategory(ctx, categoryType))
}

func ListProducts(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	cursor := ctx.Value("cursor").(*api.Page)
	filterSort := ctx.Value("filter.sort").(*api.FilterSort)

	cond := db.Cond{
		"p.status":     data.ProductStatusApproved,
		"p.deleted_at": nil,
	}
	if categoryType, ok := ctx.Value("category.type").(data.CategoryType); ok {
		cond[db.Raw("p.category->>'type'")] = categoryType.String()
	}
	if categoryValue, ok := ctx.Value("category.value").([]string); ok {
		cond[db.Raw("p.category->>'value'")] = categoryValue
	}
	query := data.DB.Select("p.*").
		From("products p").
		Where(cond).
		OrderBy("p.id DESC", "p.score DESC")
	query = filterSort.UpdateQueryBuilder(query)

	var products []*data.Product
	paginate := cursor.UpdateQueryBuilder(query)
	if err := paginate.All(&products); err != nil {
		render.Respond(w, r, err)
		return
	}
	cursor.Update(products)

	render.RenderList(w, r, presenter.NewProductList(ctx, products))
}
