package presenter

import (
	"context"

	"bitbucket.org/moodie-app/moodie-api/data"
)

type Category struct {
	Type   data.ProductCategoryType `json:"type"`
	Values []string                 `json:"values"`

	ctx context.Context
}

func NewCategory(ctx context.Context, productCategories []*data.ProductCategory) *Category {
	categoryType := ctx.Value("categoryType").(data.ProductCategoryType)
	category := &Category{
		Type: categoryType,
		ctx:  ctx,
	}
	for _, c := range productCategories {
		category.Values = append(category.Values, c.Value)
	}
	return category
}
