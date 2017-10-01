package place

import (
	"net/http"
	"strconv"

	set "gopkg.in/fatih/set.v0"

	"github.com/pressly/chi/render"

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
		OrderBy("product_id").
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

	var products []*data.Product
	query := data.DB.Product.Find(
		db.Cond{"id": productIDs.List()},
		db.Cond{"place_id": place.ID},
	).OrderBy("id")
	query = cursor.UpdateQueryUpper(query)
	if err := query.All(&products); err != nil {
		render.Respond(w, r, err)
		return
	}

	presented := presenter.NewProductList(ctx, products)
	if err := render.RenderList(w, r, presented); err != nil {
		render.Respond(w, r, err)
	}

}
