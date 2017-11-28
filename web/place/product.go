package place

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"

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
	cursor := api.NewPage(r)

	u := r.URL.Query()
	u.Del("limit")
	u.Del("page")
	var tagFilters []db.Compound
	var minPriceValue int
	var maxPriceValue int
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
				if strings.HasPrefix(tagValue, "min") {
					minPriceValue, _ = strconv.Atoi(strings.TrimPrefix(tagValue, "min"))
				}
				if strings.HasPrefix(tagValue, "max") {
					maxPriceValue, _ = strconv.Atoi(strings.TrimPrefix(tagValue, "max"))
				}
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

	// add min/max price values if they are set
	if minPriceValue > 0 && maxPriceValue > 0 {
		tagFilters = append(
			tagFilters,
			db.And(
				db.Cond{"type": data.ProductTagTypePrice},
				db.Raw("value::numeric BETWEEN ? AND ?", minPriceValue, maxPriceValue),
			),
		)
	}

	var products []*data.Product
	var itemTotal int
	if len(tagFilters) > 0 {
		query := data.DB.Select("pv.product_id").
			From("product_variants pv").
			LeftJoin("product_tags pt").
			On("pv.product_id = pt.product_id").
			Where(
				db.And(
					db.Cond{
						"pv.place_id": place.ID,
						"pv.limits >": 0,
					},
					db.Or(tagFilters...),
				),
			).
			GroupBy("pv.product_id").
			Amend(func(query string) string {
				query = query + fmt.Sprintf(" HAVING count(distinct pt.type) = %d", fCount)
				return query
			})

		var productTags []struct {
			ProductID int64 `db:"product_id"`
		}
		if err := query.All(&productTags); err != nil {
			render.Respond(w, r, err)
			return
		}

		productIDs := make([]int64, len(productTags))
		for i, t := range productTags {
			productIDs[i] = t.ProductID
		}

		productQuery := data.DB.Product.Find(db.Cond{"id": productIDs}).OrderBy("id DESC")
		productQuery = cursor.UpdateQueryUpper(productQuery)
		itemTotal = cursor.ItemTotal
		if err := productQuery.All(&products); err != nil {
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

	w.Header().Add("X-Item-Total", fmt.Sprintf("%d", itemTotal))
	presented := presenter.NewProductList(ctx, products)
	if err := render.RenderList(w, r, presented); err != nil {
		render.Respond(w, r, err)
	}
}
