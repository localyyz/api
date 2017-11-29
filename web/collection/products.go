package collection

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/data/presenter"
	"bitbucket.org/moodie-app/moodie-api/web/api"
	"github.com/go-chi/render"
	db "upper.io/db.v3"
)

func CollectionProductCtx(next http.Handler) http.Handler {
	handler := func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		collection := ctx.Value("collection").(*data.Collection)
		collectionProducts, err := data.DB.CollectionProduct.FindByCollectionID(collection.ID)
		if err != nil {
			render.Respond(w, r, err)
			return
		}

		productIDs := make([]int64, len(collectionProducts))
		for i, p := range collectionProducts {
			productIDs[i] = p.ProductID
		}

		ctx = context.WithValue(ctx, "product_ids", productIDs)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
	return http.HandlerFunc(handler)
}

func ListCollectionCategory(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	productIDs := ctx.Value("product_ids").([]int64)

	var categories []struct {
		Type   data.ProductCategoryType `db:"type" json:"type"`
		Values []string                 `db:"values" json:"values"`
	}
	err := data.DB.Select(db.Raw("distinct pc.type, array_agg(distinct pt.value) as values")).
		From("product_categories pc").
		LeftJoin("product_tags pt").
		On("pc.value = pt.value").
		Where(db.Cond{
			"pt.product_id": productIDs,
			"pt.type":       data.ProductTagTypeCategory,
		}).
		GroupBy("pc.type").
		All(&categories)
	if err != nil {
		render.Respond(w, r, err)
		return
	}
	render.Respond(w, r, categories)
}

func ListCollectionBrands(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	productIDs := ctx.Value("product_ids").([]int64)

	query := data.DB.
		Select().
		Distinct("value").
		From("product_tags").
		Where(db.Cond{
			"product_id": productIDs,
			"type":       data.ProductTagTypeBrand,
		}).
		OrderBy("value")
	var brands []struct {
		Value string `db:"value" json:"value"`
	}
	if err := query.All(&brands); err != nil {
		render.Respond(w, r, err)
		return
	}
	render.Respond(w, r, brands)
}

func ListCollectionColors(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	productIDs := ctx.Value("product_ids").([]int64)

	query := data.DB.
		Select().
		Distinct("value").
		From("product_tags").
		Where(db.Cond{
			"product_id": productIDs,
			"type":       data.ProductTagTypeColor,
		}).
		OrderBy("value")

	var colors []struct {
		Value string `db:"value" json:"value"`
	}
	if err := query.All(&colors); err != nil {
		render.Respond(w, r, err)
		return
	}
	render.Respond(w, r, colors)
}

func ListCollectionSizes(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	productIDs := ctx.Value("product_ids").([]int64)

	query := data.DB.
		Select().
		Distinct("value").
		From("product_tags").
		Where(db.Cond{
			"product_id": productIDs,
			"type":       data.ProductTagTypeSize,
		}).
		OrderBy("value")

	var sizes []struct {
		Value string `db:"value" json:"value"`
	}
	if err := query.All(&sizes); err != nil {
		render.Respond(w, r, err)
		return
	}
	render.Respond(w, r, sizes)
}

func ListCollectionProduct(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	cursor := api.NewPage(r)
	productIDs := ctx.Value("product_ids").([]int64)

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
		query := data.DB.Select("pt.product_id").
			From("product_tags pt").
			Where(
				db.And(
					db.Cond{"pt.product_id": productIDs},
					db.Or(tagFilters...),
				),
			).
			GroupBy("pt.product_id").
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

		filteredProductIDs := make([]int64, len(productTags))
		for i, t := range productTags {
			filteredProductIDs[i] = t.ProductID
		}

		productQuery := data.DB.Product.Find(db.Cond{"id": filteredProductIDs}).OrderBy("id DESC")
		productQuery = cursor.UpdateQueryUpper(productQuery)
		itemTotal = cursor.ItemTotal
		if err := productQuery.All(&products); err != nil {
			render.Respond(w, r, err)
			return
		}
	} else {
		query := data.DB.Product.Find(db.Cond{"id": productIDs}).OrderBy("id DESC")
		query = cursor.UpdateQueryUpper(query)
		itemTotal = cursor.ItemTotal
		if err := query.All(&products); err != nil {
			render.Respond(w, r, err)
			return
		}
	}

	presented := presenter.NewProductList(ctx, products)
	w.Header().Add("X-Item-Total", fmt.Sprintf("%d", itemTotal))
	if err := render.RenderList(w, r, presented); err != nil {
		render.Respond(w, r, err)
	}
}
