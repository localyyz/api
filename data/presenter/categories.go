package presenter

import (
	"context"
	"net/http"

	"bitbucket.org/moodie-app/moodie-api/data"
	"github.com/go-chi/render"
	db "upper.io/db.v3"
)

type Category struct {
	Type   data.CategoryType `json:"type"`
	Values []string          `json:"values"`

	ctx context.Context
}

func NewCategory(ctx context.Context, categoryType data.CategoryType) *Category {
	category := &Category{
		Type: categoryType,
		ctx:  ctx,
	}
	cond := db.Cond{
		"type":       categoryType,
		"mapping !=": "",
	}
	if gender, ok := ctx.Value("product.gender").(data.ProductGender); ok {
		cond["gender"] = []data.ProductGender{gender, data.ProductGenderUnisex}
	}
	var categories []*data.Category
	data.DB.
		Select(db.Raw("distinct mapping")).
		From("product_categories").
		Where(cond).
		OrderBy("mapping").
		All(&categories)

	category.Values = []string{}
	for _, c := range categories {
		category.Values = append(category.Values, c.Mapping)
	}
	return category
}

func (c *Category) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

type CategoryList []*Category

func NewCategoryList(ctx context.Context, categories []data.CategoryType) []render.Renderer {
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
