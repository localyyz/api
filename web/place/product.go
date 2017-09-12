package place

import (
	"net/http"

	set "gopkg.in/fatih/set.v0"

	"github.com/pressly/chi/render"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/data/presenter"
	"bitbucket.org/moodie-app/moodie-api/web/api"
	"upper.io/db.v3"
)

// List products at a given place
func ListProduct(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	place := ctx.Value("place").(*data.Place)
	cursor := api.NewPage(r)

	u := r.URL.Query()
	var tagFilters []db.Compound
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
			tagFilters = append(
				tagFilters,
				db.Cond{
					"value": tagValue,
					"type":  tagType,
				},
			)
		}

	}

	// TODO: handle empty tags
	var productTags []*data.ProductTag
	data.DB.
		Select("product_id").
		From("product_tags").
		Where(
			db.And(
				db.Cond{"place_id": place.ID},
				db.Or(tagFilters...),
			),
		).
		GroupBy("product_id").
		All(&productTags)

	productIDs := set.New()
	for _, pt := range productTags {
		productIDs.Add(pt.ProductID)
	}
	if productIDs.Size() == 0 {
		render.Respond(w, r, []struct{}{})
		return
	}

	var products []*data.Product
	query := data.DB.Product.Find(
		db.Cond{"id": productIDs.List()},
		db.Cond{"place_id": place.ID},
	)
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
