package category

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

func listCategoryProduct(ctx context.Context, categories []*data.ProductCategory) ([]*data.Product, error) {
	var categoryValues []string
	for _, v := range categories {
		categoryValues = append(categoryValues, v.Value)
	}

	// search the category tags that match the query term
	query := data.DB.
		Select(db.Raw("distinct pv.product_id")).
		From("product_variants pv").
		LeftJoin("product_tags pt").On("pv.product_id = pt.product_id").
		Where(db.Cond{
			"pt.value":    categoryValues,
			"pt.type":     data.ProductTagTypeCategory,
			"pv.limits >": 0,
		}).
		GroupBy("pv.product_id").
		OrderBy("-pv.product_id")

	cursor := ctx.Value("cursor").(*api.Page)
	paginator := cursor.UpdateQueryBuilder(query)
	var rows []*data.ProductVariant
	if err := paginator.All(&rows); err != nil {
		return nil, err
	}

	var productIDs []int64
	for _, row := range rows {
		productIDs = append(productIDs, row.ProductID)
	}
	var products []*data.Product
	if err := data.DB.Product.Find(
		db.Cond{"id": productIDs},
	).OrderBy("-id").All(&products); err != nil {
		return nil, err
	}
	cursor.Update(products)

	// count query
	row, _ := data.DB.
		Select(db.Raw("count(distinct pv.product_id)")).
		From("product_variants pv").
		LeftJoin("product_tags pt").On("pv.product_id = pt.product_id").
		Where(db.Cond{
			"pt.value":    categoryValues,
			"pt.type":     data.ProductTagTypeCategory,
			"pv.limits >": 0,
		}).QueryRow()
	row.Scan(&cursor.ItemTotal)

	return products, nil
}

type categoryFilterRequest struct {
	Filters []struct {
		Type  data.ProductTagType `json:"type,required"`
		Value string              `json:"value,required"`
	} `json:"filters"`
}

func (cf *categoryFilterRequest) Bind(*http.Request) error {
	return nil
}

// filter category product based on subcategories
func filterCategoryProduct(ctx context.Context, payload *categoryFilterRequest) ([]*data.Product, error) {
	var (
		tagFilters                           []db.Compound
		minPriceValue, maxPriceValue, fCount int
	)

	for _, f := range payload.Filters {
		fCount += 1
		var filterValue interface{}

		switch f.Type {
		case data.ProductTagTypeGender:
			filterValue = "man"
			if f.Value == "female" {
				filterValue = "woman"
			}
		case data.ProductTagTypePrice:
			if strings.HasPrefix(f.Value, "min") {
				minPriceValue, _ = strconv.Atoi(strings.TrimPrefix(f.Value, "min"))
			}
			if strings.HasPrefix(f.Value, "max") {
				maxPriceValue, _ = strconv.Atoi(strings.TrimPrefix(f.Value, "max"))
			}
			continue
		case data.ProductTagTypeCategory:
			mappings, err := data.DB.ProductCategory.FindByMapping(f.Value)
			if err != nil {
				continue
			}
			var tagValues []string
			for _, m := range mappings {
				tagValues = append(tagValues, m.Value)
			}
			filterValue = tagValues
		}

		tagFilters = append(
			tagFilters,
			db.Cond{
				"type":  f.Type,
				"value": filterValue,
			},
		)
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
	query := data.DB.Select(db.Raw("distinct pv.product_id")).
		From("product_variants pv").
		LeftJoin("product_tags pt").
		On("pv.product_id = pt.product_id").
		Where(db.Or(tagFilters...)).
		GroupBy("pv.product_id").
		Amend(func(query string) string {
			query = query + fmt.Sprintf(" HAVING count(distinct pt.type) = %d AND max(pv.limits) > 0", fCount)
			return query
		})
	var productTags []struct {
		ProductID int64 `db:"product_id"`
	}
	if err := query.All(&productTags); err != nil {
		return nil, err
	}

	productIDs := make([]int64, len(productTags))
	for i, t := range productTags {
		productIDs[i] = t.ProductID
	}

	productQuery := data.DB.Product.Find(
		db.Cond{"id": productIDs},
	).OrderBy("id DESC")
	cursor := ctx.Value("cursor").(*api.Page)
	productQuery = cursor.UpdateQueryUpper(productQuery)
	if err := productQuery.All(&products); err != nil {
		return nil, err
	}
	cursor.Update(products)

	return products, nil
}

// Return all products within a top level category
func ListCategoryProduct(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	categoryType := ctx.Value("categoryType").(data.ProductCategoryType)

	var payload categoryFilterRequest
	if err := render.Bind(r, &payload); err != nil {
		render.Render(w, r, api.ErrInvalidRequest(err))
		return
	}

	var (
		products []*data.Product
		err      error
	)
	if len(payload.Filters) > 0 {
		products, err = filterCategoryProduct(ctx, &payload)
	} else {
		// find category tag values
		categories, err := data.DB.ProductCategory.FindByType(categoryType)
		if err != nil {
			render.Respond(w, r, err)
			return
		}

		products, err = listCategoryProduct(ctx, categories)
	}

	if err != nil {
		render.Respond(w, r, err)
		return
	}
	render.RenderList(w, r, presenter.NewSearchProductList(ctx, products))
}
