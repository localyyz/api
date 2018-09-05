package category

import (
	"context"
	"net/http"
	"strconv"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/data/presenter"
	"bitbucket.org/moodie-app/moodie-api/web/api"
	"github.com/go-chi/chi"
	"github.com/go-chi/render"
	"github.com/pressly/lg"
	db "upper.io/db.v3"
)

func CategoryCtx(next http.Handler) http.Handler {
	handler := func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		categoryID, err := strconv.ParseInt(chi.URLParam(r, "categoryID"), 10, 64)
		if err != nil {
			render.Render(w, r, api.ErrBadID)
			return
		}

		// did not parse as category. attempt to parse as subcategory
		category, err := data.DB.Category.FindByID(categoryID)
		if err != nil {
			// did not parse either. return error
			render.Respond(w, r, api.ErrInvalidRequest(err))
			return
		}

		ctx = context.WithValue(ctx, "category", category)
		lg.SetEntryField(ctx, "category", category.Value)

		next.ServeHTTP(w, r.WithContext(ctx))
	}
	return http.HandlerFunc(handler)
}

func CategoryRootCtx(next http.Handler) http.Handler {
	handler := func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		filterSort := ctx.Value(api.FilterSortCtxKey).(*api.FilterSort)
		if f := filterSort.Gender(); f != nil {
			// use the filter value to find the root category node
			// did not parse as category. attempt to parse as subcategory
			var category *data.Category
			err := data.DB.Category.
				Find(db.Cond{"value": f.Value}).
				OrderBy("id").
				One(&category)
			if err != nil {
				// did not parse either. return error
				render.Respond(w, r, api.ErrInvalidRequest(err))
				return
			}
			ctx = context.WithValue(ctx, "category", category)
		}

		next.ServeHTTP(w, r.WithContext(ctx))
	}
	return http.HandlerFunc(handler)
}

func List(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var (
		categories []*data.Category
		err        error
	)
	if root, ok := ctx.Value("category").(*data.Category); ok {
		// NOTE/TODO: this is for backwards compatibility.
		// check if a root node is found
		// frontend is querying via FilterSort params
		// ie: /categories?filter=gender,val=man
		categories, err = data.DB.Category.FindDescendants(root.ID)
	} else {
		err = data.DB.Category.
			Find().
			OrderBy("id").
			All(&categories)
	}
	if err != nil {
		render.Respond(w, r, err)
		return
	}

	presented := presenter.NewCategoryList(ctx, categories)
	presented = append(presented, []render.Renderer{
		&presenter.Node{
			Category: &data.Category{
				ID:    10,
				Label: "Sales",
				Value: "sales",
			},
			Values: []*presenter.Node{
				{
					Category: &data.Category{
						ID:       13,
						Label:    "70% OFF",
						Value:    "70%+ OFF",
						ImageURL: "https://cdn.shopify.com/s/files/1/0052/8731/3526/files/70.png?17957505310432019141",
					},
				},
				{
					Category: &data.Category{
						ID:       12,
						Label:    "50%-70% OFF",
						Value:    "50%-70% OFF",
						ImageURL: "https://cdn.shopify.com/s/files/1/0052/8731/3526/files/50.png?5115785919598170614",
					},
				},
				{
					Category: &data.Category{
						ID:       11,
						Value:    "20% OFF",
						Label:    "20%-50% OFF",
						ImageURL: "https://cdn.shopify.com/s/files/1/0052/8731/3526/files/20.png?14969378164451378728",
					},
				},
			},
		},
		&presenter.Node{
			Category: &data.Category{
				ID:    20,
				Value: "collections",
				Label: "Collections",
			},
			Values: []*presenter.Node{
				{
					Category: &data.Category{
						ID:       21,
						Label:    "Under $50",
						ImageURL: "https://cdn.shopify.com/s/files/1/0835/3729/products/Oversized_Hoodies_-4_eda921cf-882d-479f-8d07-ed1c070b0a0a.jpg",
					},
				},
				{
					Category: &data.Category{
						ID:       22,
						Label:    "$50 - $200",
						ImageURL: "https://cdn.shopify.com/s/files/1/1066/9348/products/UNG85206_red_0.jpg",
					},
				},
				{
					Category: &data.Category{
						ID:       23,
						Label:    "$200 plus",
						ImageURL: "https://cdn.shopify.com/s/files/1/0444/7969/products/mens-jackets-coats-hexagon-stitch-brother-jacket-1.jpg",
					},
				},
			},
		},
	}...)

	render.RenderList(w, r, presented)
}

func GetCategory(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	category := ctx.Value("category").(*data.Category)
	presented := presenter.NewCategory(ctx, category)
	render.Render(w, r, presented)
}

func ListProducts(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	cursor := ctx.Value("cursor").(*api.Page)
	filterSort := ctx.Value("filter.sort").(*api.FilterSort)
	root := ctx.Value("category").(*data.Category)

	descendents, err := data.DB.Category.FindDescendants(root.ID)
	if err != nil {
		render.Respond(w, r, err)
		return
	}
	categoryIDs := []int64{root.ID}
	for _, d := range descendents {
		categoryIDs = append(categoryIDs, d.ID)
	}

	cond := db.Cond{
		"p.status":      data.ProductStatusApproved,
		"p.deleted_at":  nil,
		"p.category_id": categoryIDs,
	}
	query := data.DB.Select("p.*").
		From("products p").
		Where(cond).
		OrderBy("p.id DESC")
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
