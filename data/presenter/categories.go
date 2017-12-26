package presenter

import (
	"context"
	"net/http"

	"bitbucket.org/moodie-app/moodie-api/data"
	"github.com/go-chi/render"
	db "upper.io/db.v3"
)

type Category struct {
	Type   data.ProductCategoryType `json:"type"`
	Values []string                 `json:"values"`

	ctx context.Context
}

func NewCategory(ctx context.Context, categoryType data.ProductCategoryType) *Category {
	category := &Category{
		Type: categoryType,
		ctx:  ctx,
	}
	var productCategories []*data.ProductCategory
	data.DB.
		Select(db.Raw("distinct mapping")).
		From("product_categories").
		Where(
			db.Cond{
				"type":       categoryType,
				"mapping !=": "",
			},
		).
		OrderBy("mapping").
		All(&productCategories)

	for _, c := range productCategories {
		category.Values = append(category.Values, c.Mapping)
	}
	return category
}

func (c *Category) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

type CategoryList []*Category

func NewCategoryList(ctx context.Context, categories []data.ProductCategoryType) []render.Renderer {
	list := []render.Renderer{}
	for _, c := range categories {
		list = append(list, NewCategory(ctx, c))
	}
	return list
}

func (l CategoryList) Render(w http.ResponseWriter, r *http.Request) error {
	for _, v := range l {
		if err := v.Render(w, r); err != nil {
			return err
		}
	}
	return nil
}
