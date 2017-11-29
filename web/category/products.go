package category

import (
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

func ListProductBrands(w http.ResponseWriter, r *http.Request) {
	category := r.URL.Query().Get("category")
	var tagProducts []struct {
		ProductID int64 `db:"product_id"`
	}
	data.DB.
		Select().
		Distinct("product_id").
		From("product_tags").
		Where(db.Cond{
			"type":  data.ProductTagTypeCategory,
			"value": category,
		}).
		All(&tagProducts)

	var productIDs []int64
	for _, t := range tagProducts {
		productIDs = append(productIDs, t.ProductID)
	}

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

func ListProductColors(w http.ResponseWriter, r *http.Request) {
	category := r.URL.Query().Get("category")

	var tagProducts []struct {
		ProductID int64 `db:"product_id"`
	}
	data.DB.
		Select().
		Distinct("product_id").
		From("product_tags").
		Where(db.Cond{
			"type":  data.ProductTagTypeCategory,
			"value": category,
		}).
		All(&tagProducts)
	var productIDs []int64
	for _, t := range tagProducts {
		productIDs = append(productIDs, t.ProductID)
	}

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
func ListProductSizes(w http.ResponseWriter, r *http.Request) {
	category := r.URL.Query().Get("category")

	var tagProducts []struct {
		ProductID int64 `db:"product_id"`
	}
	data.DB.
		Select().
		Distinct("product_id").
		From("product_tags").
		Where(db.Cond{
			"type":  data.ProductTagTypeCategory,
			"value": category,
		}).
		All(&tagProducts)

	var productIDs []int64
	for _, t := range tagProducts {
		productIDs = append(productIDs, t.ProductID)
	}

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

func ListCategoryProduct(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	categoryType := ctx.Value("categoryType").(data.ProductCategoryType)
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
		// find corresponding category
		categories, err := data.DB.ProductCategory.FindByType(categoryType)
		if err != nil {
			render.Respond(w, r, err)
			return
		}

		var categoryValues []string
		for _, v := range categories {
			categoryValues = append(categoryValues, v.Value)
		}

		// search the category tags that match the query term
		query := data.DB.
			Select(db.Raw("distinct on (product_id, created_at) *")).
			From("product_tags").
			Where(db.Cond{
				"value": categoryValues,
				"type":  data.ProductTagTypeCategory,
			}).
			GroupBy("id", "product_id", "created_at").
			OrderBy("-created_at")

		var counts []struct{}
		err = query.
			Limit(0).
			Offset(0).
			All(&counts)
		if err != nil {
			render.Respond(w, r, err)
			return
		}
		itemTotal = len(counts)

		query = cursor.UpdateQueryBuilder(query)
		var tags []*data.ProductTag
		if err := query.All(&tags); err != nil {
			render.Respond(w, r, err)
			return
		}
		productIDs := make([]int64, len(tags))
		for i, t := range tags {
			productIDs[i] = t.ProductID
		}

		// find the products
		products, err = data.DB.Product.FindAll(db.Cond{"id": productIDs})
		if err != nil {
			render.Respond(w, r, err)
			return
		}

	}
	w.Header().Add("X-Item-Total", fmt.Sprintf("%d", itemTotal))
	render.RenderList(w, r, presenter.NewSearchProductList(ctx, products))

}
