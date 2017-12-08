package presenter

import (
	"context"
	"net/http"
	"time"

	"github.com/go-chi/render"

	db "upper.io/db.v3"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/lib/htmlx"
)

type Product struct {
	*data.Product
	Variants []*ProductVariant `json:"variants"`
	Place    *data.Place       `json:"place"`

	Sizes  []string `json:"sizes"`
	Colors []string `json:"colors"`

	CreateAt  *time.Time `json:"createdAt,omitempty"`
	UpdatedAt *time.Time `json:"updatedAt,omitempty"`
	DeleteAt  *time.Time `json:"deletedAt,omitempty"`

	HtmlDescription  string `json:"htmlDescription"`
	NoTagDescription string `json:"noTagDescription"`

	ctx context.Context
}

type SearchProductList []*Product

func (l SearchProductList) Render(w http.ResponseWriter, r *http.Request) error {
	for _, v := range l {
		if err := v.Render(w, r); err != nil {
			return err
		}
	}
	return nil
}

func NewSearchProductList(ctx context.Context, products []*data.Product) []render.Renderer {
	list := []render.Renderer{}
	for _, product := range products {
		list = append(list, NewProduct(ctx, product))
	}
	return list
}

type CartProductList []*Product

func (l CartProductList) Render(w http.ResponseWriter, r *http.Request) error {
	for _, v := range l {
		if err := v.Render(w, r); err != nil {
			return err
		}
	}
	return nil
}

func NewCartProductList(ctx context.Context, products []*data.Product) CartProductList {
	list := CartProductList{}
	variants := ctx.Value("variants").(map[int64]*data.ProductVariant)
	for _, product := range products {
		p := NewProduct(ctx, product)
		p.Variants = []*ProductVariant{NewProductVariant(ctx, variants[p.ID])}
		list = append(list, p)
	}
	return list
}

func NewProductList(ctx context.Context, products []*data.Product) []render.Renderer {
	list := []render.Renderer{}
	for _, product := range products {
		list = append(list, NewProduct(ctx, product))
	}
	return list
}

func NewProduct(ctx context.Context, product *data.Product) *Product {
	p := &Product{
		Product: product,
		ctx:     ctx,
	}

	p.Variants = []*ProductVariant{}
	if dbVariants, _ := data.DB.ProductVariant.FindByProductID(product.ID); dbVariants != nil {
		for _, v := range dbVariants {
			p.Variants = append(p.Variants, &ProductVariant{ProductVariant: v})
		}
	}

	p.Sizes = []string{}
	p.Colors = []string{}
	if tags, _ := data.DB.ProductTag.FindAll(
		db.Cond{
			"product_id": product.ID,
			"type":       []data.ProductTagType{data.ProductTagTypeSize, data.ProductTagTypeColor},
		}); tags != nil {
		for _, tt := range tags {
			switch tt.Type {
			case data.ProductTagTypeSize:
				p.Sizes = append(p.Sizes, tt.Value)
			case data.ProductTagTypeColor:
				p.Colors = append(p.Colors, tt.Value)
			}
		}
	}

	// if place is not in the context, return it as part of presenter
	if _, ok := ctx.Value("place").(*data.Place); !ok {
		if place, _ := data.DB.Place.FindByID(p.PlaceID); place != nil {
			p.Place = place
		}
	}

	return p
}

func (p *Product) Render(w http.ResponseWriter, r *http.Request) error {
	p.HtmlDescription = htmlx.CaptionizeHtmlBody(p.Description, -1)
	p.NoTagDescription = htmlx.StripTags(p.Description)

	return nil
}

type ProductCategory struct {
	*data.ProductCategory
	ImageURL string   `json:"imageUrl"`
	Values   []string `json:"values"`
}

func (c *ProductCategory) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

type ProductCategoryList []*ProductCategory

func (l ProductCategoryList) Render(w http.ResponseWriter, r *http.Request) error {
	for _, v := range l {
		if err := v.Render(w, r); err != nil {
			return err
		}
	}
	return nil
}

func NewProductCategoryList(ctx context.Context, categories []*data.ProductCategory) []render.Renderer {
	list := []render.Renderer{}
	for _, cat := range categories {
		list = append(list, &ProductCategory{ProductCategory: cat})
	}
	return list
}
