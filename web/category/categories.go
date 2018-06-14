package category

import (
	"context"
	"net/http"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/data/presenter"
	"bitbucket.org/moodie-app/moodie-api/web/api"
	"github.com/go-chi/chi"
	"github.com/go-chi/render"
	db "upper.io/db.v3"
)

func Routes() chi.Router {
	r := chi.NewRouter()

	r.Get("/", List)
	r.Route("/{categoryType}", func(r chi.Router) {
		r.Use(api.FilterSortCtx)
		r.Use(CategoryTypeCtx)
		r.Get("/", GetCategory)

		r.Route("/{subcategory}", func(r chi.Router) {
			r.Use(SubcategoryCtx)
			r.Get("/products", ListProducts)
		})
		r.Get("/products", ListProducts)
	})

	return r
}

func SubcategoryCtx(next http.Handler) http.Handler {
	handler := func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		subcategory := chi.URLParam(r, "subcategory")
		categories, err := data.DB.Category.FindByMapping(subcategory)
		if err != nil {
			render.Respond(w, r, err)
			return
		}

		var values []string
		for _, c := range categories {
			values = append(values, c.Value)
		}
		ctx = context.WithValue(ctx, "category.value", values)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
	return http.HandlerFunc(handler)
}

func CategoryTypeCtx(next http.Handler) http.Handler {
	handler := func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		rawCategoryType := chi.URLParam(r, "categoryType")

		var categoryType data.CategoryType
		if err := categoryType.UnmarshalText([]byte(rawCategoryType)); err != nil {
			render.Render(w, r, api.ErrInvalidRequest(err))
			return
		}

		ctx = context.WithValue(ctx, "category.type", categoryType)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
	return http.HandlerFunc(handler)
}

var (
	displayCategories = []data.CategoryType{
		data.CategoryApparel,
		data.CategoryHandbag,
		data.CategoryShoe,
		data.CategoryJewelry,
		data.CategoryAccessory,
		data.CategoryCosmetic,
		//data.CategoryFragrance,
		data.CategorySneaker,
		//data.CategoryLingerie,
		data.CategorySwimwear,
	}
)

func List(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if gender, ok := ctx.Value("session.gender").(data.UserGender); ok {
		ctx = context.WithValue(ctx, "product.gender", gender)
	}
	render.RenderList(w, r, presenter.NewCategoryList(ctx, displayCategories))
}

func GetCategory(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	categoryType := ctx.Value("category.type").(data.CategoryType)

	if gender, ok := ctx.Value("session.gender").(data.UserGender); ok {
		ctx = context.WithValue(ctx, "product.gender", gender)
	}
	render.Render(w, r, presenter.NewCategory(ctx, categoryType))
}

func ListProducts(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	categoryType := ctx.Value("category.type").(data.CategoryType)
	cursor := ctx.Value("cursor").(*api.Page)
	filterSort := ctx.Value("filter.sort").(*api.FilterSort)

	cond := db.Cond{
		"p.status":                    data.ProductStatusApproved,
		db.Raw("p.category->>'type'"): categoryType.String(),
		"p.deleted_at":                nil,
	}
	if categoryValue, ok := ctx.Value("category.value").([]string); ok {
		cond[db.Raw("p.category->>'value'")] = categoryValue
	}
	query := data.DB.Select("p.*").
		From("products p").
		Where(cond).
		OrderBy("p.score DESC")
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
