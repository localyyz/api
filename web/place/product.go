package place

import (
	"context"
	"net/http"

	"github.com/go-chi/render"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/data/presenter"
	"bitbucket.org/moodie-app/moodie-api/web/api"
	"upper.io/db.v3"
	"upper.io/db.v3/lib/sqlbuilder"
)

// List available top level product categories for a given place
func ListProductCategory(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	place := ctx.Value("place").(*data.Place)

	var categories []struct {
		Type   data.ProductCategoryType `db:"type" json:"type"`
		Values []string                 `db:"values" json:"values"`
	}
	err := data.DB.Select(db.Raw("distinct pc.type, array_agg(distinct pt.value) as values")).
		From("product_categories pc").
		LeftJoin("product_tags pt").
		On("pc.value = pt.value").
		Where(db.Cond{
			"pt.place_id": place.ID,
			"pt.type":     data.ProductTagTypeCategory,
		}).
		GroupBy("pc.type").
		All(&categories)
	if err != nil {
		render.Respond(w, r, err)
		return
	}
	render.Respond(w, r, categories)
}

func ListProductColors(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	place := ctx.Value("place").(*data.Place)
	categoryType, _ := ctx.Value("categoryType").(string)

	var query sqlbuilder.Selector
	if len(categoryType) != 0 {

		var tagProducts []struct {
			ProductID int64 `db:"product_id"`
		}
		data.DB.
			Select().
			Distinct("product_id").
			From("product_tags").
			Where(db.Cond{
				"type":     data.ProductTagTypeCategory,
				"value":    categoryType,
				"place_id": place.ID,
			}).
			All(&tagProducts)

		var productIDs []int64
		for _, t := range tagProducts {
			productIDs = append(productIDs, t.ProductID)
		}

		query = data.DB.
			Select().
			Distinct("value").
			From("product_tags").
			Where(db.Cond{
				"product_id": productIDs,
				"type":       data.ProductTagTypeColor,
			}).
			OrderBy("value")
	} else {
		// depending on if category is selected...
		query = data.DB.Select().
			Distinct("value").
			From("product_tags").
			Where(
				db.Cond{
					"place_id": place.ID,
					"type":     data.ProductTagTypeColor,
				},
			).
			OrderBy("value")
	}

	var colors []struct {
		Value string `db:"value" json:"value"`
	}
	if err := query.All(&colors); err != nil {
		render.Respond(w, r, err)
		return
	}

	render.Respond(w, r, colors)
}

func ListProductSizes(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	place := ctx.Value("place").(*data.Place)
	categoryType, _ := ctx.Value("categoryType").(string)

	var query sqlbuilder.Selector
	if len(categoryType) != 0 {

		var tagProducts []struct {
			ProductID int64 `db:"product_id"`
		}
		data.DB.
			Select().
			Distinct("product_id").
			From("product_tags").
			Where(db.Cond{
				"type":     data.ProductTagTypeCategory,
				"value":    categoryType,
				"place_id": place.ID,
			}).
			All(&tagProducts)

		var productIDs []int64
		for _, t := range tagProducts {
			productIDs = append(productIDs, t.ProductID)
		}

		query = data.DB.
			Select().
			Distinct("value").
			From("product_tags").
			Where(db.Cond{
				"product_id": productIDs,
				"type":       data.ProductTagTypeSize,
			}).
			OrderBy("value")
	} else {
		// depending on if category is selected...
		query = data.DB.Select().
			Distinct("value").
			From("product_tags").
			Where(
				db.Cond{
					"place_id": place.ID,
					"type":     data.ProductTagTypeSize,
				},
			).
			OrderBy("value")
	}
	var sizes []struct {
		Value string `db:"value" json:"value"`
	}
	if err := query.All(&sizes); err != nil {
		render.Respond(w, r, err)
		return
	}
	render.Respond(w, r, sizes)
}

func ListProductBrands(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	place := ctx.Value("place").(*data.Place)
	categoryType, _ := ctx.Value("categoryType").(string)

	var query sqlbuilder.Selector
	if len(categoryType) != 0 {

		var tagProducts []struct {
			ProductID int64 `db:"product_id"`
		}
		data.DB.
			Select().
			Distinct("product_id").
			From("product_tags").
			Where(db.Cond{
				"type":     data.ProductTagTypeCategory,
				"value":    categoryType,
				"place_id": place.ID,
			}).
			All(&tagProducts)

		var productIDs []int64
		for _, t := range tagProducts {
			productIDs = append(productIDs, t.ProductID)
		}

		query = data.DB.
			Select().
			Distinct("value").
			From("product_tags").
			Where(db.Cond{
				"product_id": productIDs,
				"type":       data.ProductTagTypeBrand,
			}).
			OrderBy("value")
	} else {
		// depending on if category is selected...
		query = data.DB.Select().
			Distinct("value").
			From("product_tags").
			Where(
				db.Cond{
					"place_id": place.ID,
					"type":     data.ProductTagTypeBrand,
				},
			).
			OrderBy("value")
	}

	var brands []struct {
		Value string `db:"value" json:"value"`
	}
	if err := query.All(&brands); err != nil {
		render.Respond(w, r, err)
		return
	}

	render.Respond(w, r, brands)
}

func ProductCategoryCtx(next http.Handler) http.Handler {
	handler := func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		for k, v := range r.URL.Query() {
			if v == nil || len(v) == 0 {
				continue
			}
			// find filterings
			var filterType data.ProductTagType
			if err := filterType.UnmarshalText([]byte(k)); err != nil {
				// unrecognized filter tag type
				continue
			}
			if filterType == data.ProductTagTypeCategory {
				ctx = context.WithValue(ctx, "categoryType", v[0])
				break
			}
		}
		next.ServeHTTP(w, r.WithContext(ctx))
	}
	return http.HandlerFunc(handler)
}

// List products at a given place
func ListProduct(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	place := ctx.Value("place").(*data.Place)

	query := data.DB.
		Select(db.Raw("distinct p.*")).
		From("products p").
		LeftJoin("product_variants pv").
		On("pv.product_id = p.id").
		Where(
			db.Cond{
				"pv.limits >": 0,
				"p.place_id":  place.ID,
			},
		).
		GroupBy("p.id").
		OrderBy("-p.id")
	cursor := ctx.Value("cursor").(*api.Page)
	paginate := cursor.UpdateQueryBuilder(query)

	var products []*data.Product
	if err := paginate.All(&products); err != nil {
		render.Respond(w, r, err)
		return
	}
	cursor.Update(products)

	// count query
	row, _ := data.DB.
		Select(db.Raw("count(distinct pv.product_id)")).
		From("product_variants pv").
		Where(db.Cond{
			"pv.limits >": 0,
			"pv.place_id": place.ID,
		}).QueryRow()
	row.Scan(&cursor.ItemTotal)

	presented := presenter.NewProductList(ctx, products)
	if err := render.RenderList(w, r, presented); err != nil {
		render.Respond(w, r, err)
	}
}
