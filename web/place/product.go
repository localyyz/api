package place

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/render"

	set "gopkg.in/fatih/set.v0"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/data/presenter"
	"bitbucket.org/moodie-app/moodie-api/web/api"
	"upper.io/db.v3"
)

func ListProductPrices(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	place := ctx.Value("place").(*data.Place)

	query := data.DB.
		Select(db.Raw("distinct round((etc->>'prc')::numeric,-1) as price")).
		From("product_variants").
		Where(db.Cond{"place_id": place.ID}).
		OrderBy("price").
		Iterator()

	var prices []float64
	for query.Next() {
		var price float64
		if err := query.Scan(&price); err != nil {
			render.Respond(w, r, err)
			return
		}
		prices = append(prices, price)
	}

	// less than the count distrubution, just return
	if len(prices) < 4 {
		render.Respond(w, r, prices)
		return
	}

	// find the distribution
	// 0-25%
	// 26-50%
	// 51-80%
	// 80-100%
	distribution := make([]float64, 4, 4)
	for i, p := range []float64{0.25, 0.5, 0.8} {
		distribution[i] = prices[int(p*float64(len(prices)))]
	}
	distribution[3] = prices[len(prices)-1]

	render.Respond(w, r, distribution)
}

func CategoryTypeCtx(next http.Handler) http.Handler {
	handler := func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		place := ctx.Value("place").(*data.Place)

		catType := r.URL.Query().Get("type")
		if len(catType) == 0 {
			render.Render(w, r, api.ErrBadID)
		}

		category, err := data.DB.ProductCategory.FindOne(
			db.Cond{
				"place_id": place.ID,
				"name":     catType,
			})
		if err != nil {
			render.Respond(w, r, err)
			return
		}

		ctx = context.WithValue(ctx, "category", category)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
	return http.HandlerFunc(handler)
}

//func ListProductCategory(w http.ResponseWriter, r *http.Request) {
//ctx := r.Context()
//place := ctx.Value("place").(*data.Place)
//category := ctx.Value("category").(*data.ProductCategory)

//cursor := api.NewPage(r)

//var list []struct {
//Value string `db:"value" json:"value"`
//}
//iter := data.DB.Iterator(
//`SELECT distinct(pt.value)
//FROM product_tags pt, product_variants pv
//WHERE (pt.place_id = ? AND pt.type = 6)
//AND pv.product_id = pt.product_id
//AND value IN ?
//GROUP BY value
//HAVING sum(pv.limits) > 1
//ORDER BY value ASC
//LIMIT ? OFFSET ?`,
//place.ID,
//category.Value,
//cursor.Limit,
//(cursor.Page-1)*cursor.Limit,
//)
//defer iter.Close()

//if err := iter.All(&list); err != nil {
//render.Respond(w, r, err)
//return
//}

//render.Respond(w, r, list)
//}

// List available top level product categories for a given place
func ListProductCategory(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	place := ctx.Value("place").(*data.Place)

	var categories []struct {
		*data.ProductCategory
		Values []string `db:"values" json:"values"`
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

func ListProductBrands(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	place := ctx.Value("place").(*data.Place)
	cursor := api.NewPage(r)

	var brands []struct {
		Value string `db:"value" json:"value"`
	}
	iter := data.DB.Iterator(
		`SELECT distinct(pt.value)
		FROM product_tags pt, product_variants pv
		WHERE (pt.place_id = ? AND pt.type = 7)
		and pv.product_id = pt.product_id
		group by value
		having sum(pv.limits) > 1
		ORDER BY value ASC
		LIMIT ? OFFSET ?`,
		place.ID,
		cursor.Limit,
		(cursor.Page-1)*cursor.Limit,
	)
	defer iter.Close()

	if err := iter.All(&brands); err != nil {
		render.Respond(w, r, err)
		return
	}

	render.Respond(w, r, brands)
}

func ListProductTags(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	place := ctx.Value("place").(*data.Place)

	var tagType data.ProductTagType
	if err := tagType.UnmarshalText([]byte(r.URL.Query().Get("t"))); err != nil {
		render.Respond(w, r, api.ErrInvalidRequest(err))
		return
	}

	query := data.DB.
		Select(db.Raw("distinct(value)")).
		From("product_tags").
		Where(
			db.Cond{
				"place_id": place.ID,
				"type":     tagType,
			},
		).
		OrderBy("value")

	var brands []struct {
		Value string `db:"value" json:"value"`
	}
	if err := query.All(&brands); err != nil {
		render.Respond(w, r, err)
		return
	}

	render.Respond(w, r, brands)
	return
}

// List products at a given place
func ListProduct(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	place := ctx.Value("place").(*data.Place)
	cursor := api.NewPage(r)

	u := r.URL.Query()
	u.Del("limit")
	u.Del("page")
	var tagFilters []db.Compound
	var fCount int
	for k, v := range u {
		// find filterings
		var tagType data.ProductTagType
		if err := tagType.UnmarshalText([]byte(k)); err != nil {
			// unrecognized filter tag type
			continue
		}

		if v == nil || len(v) == 0 {
			continue
		}

		fCount += 1
		// values can be multiple, ie: {gender: [male female]}
		for _, vv := range v {
			tagValue := vv
			// translate gender tag value
			if tagType == data.ProductTagTypeGender {
				tagValue = "man"
				if vv == "female" {
					tagValue = "woman"
				}
			}

			if tagType == data.ProductTagTypePrice {
				// if it's price, need to do some numeric conversions
				tagNumValue, _ := strconv.Atoi(tagValue)
				tagFilters = append(
					tagFilters,
					db.And(
						db.Cond{"type": tagType},
						db.Raw("value::numeric <= ?", tagNumValue),
					),
				)
				continue
			}

			tagFilters = append(
				tagFilters,
				db.Cond{
					"type":  tagType,
					"value": tagValue,
				},
			)
		}
	}

	var products []*data.Product
	if len(tagFilters) > 0 {
		// TODO: handle empty tags
		var productTags []struct {
			*data.ProductTag
			Count int `db:"c"`
		}
		data.DB.
			Select("product_id", db.Raw("count(*) as c")).
			From("product_tags").
			Where(
				db.And(
					db.Cond{"place_id": place.ID},
					db.Or(tagFilters...),
				),
			).
			GroupBy("product_id").
			//Having("count(*) > ?", len(tagFilters)).
			All(&productTags)
		// TODO: having.

		productIDs := set.New()
		for _, pt := range productTags {
			if len(tagFilters) == 0 || pt.Count == fCount {
				productIDs.Add(pt.ProductID)
			}
		}
		if productIDs.Size() == 0 {
			render.Respond(w, r, []struct{}{})
			return
		}

		query := data.DB.Product.Find(
			db.Cond{"id": productIDs.List()},
			db.Cond{"place_id": place.ID},
		).OrderBy("-created_at")
		query = cursor.UpdateQueryUpper(query)
		if err := query.All(&products); err != nil {
			render.Respond(w, r, err)
			return
		}
	} else {
		iter := data.DB.Iterator(`
			SELECT p.*
			FROM products p
			LEFT JOIN product_variants pv
			ON p.id = pv.product_id
			WHERE p.place_id = ?
			GROUP BY p.id
			HAVING sum(pv.limits) > 0
			ORDER BY p.created_at DESC
			LIMIT ? OFFSET ?`, place.ID, cursor.Limit, (cursor.Page-1)*cursor.Limit)
		defer iter.Close()

		if err := iter.All(&products); err != nil {
			render.Respond(w, r, err)
			return
		}
	}

	w.Header().Add("X-Item-Total", fmt.Sprintf("%d", cursor.ItemTotal))
	presented := presenter.NewProductList(ctx, products)
	if err := render.RenderList(w, r, presented); err != nil {
		render.Respond(w, r, err)
	}
}
