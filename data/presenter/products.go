package presenter

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"

	db "upper.io/db.v3"

	"github.com/pressly/chi/render"

	"bitbucket.org/moodie-app/moodie-api/data"
)

type Product struct {
	*data.Product
	Variants []*ProductVariant `json:"variants"`
	Place    *Place            `json:"place"`

	Sizes  []string `json:"sizes"`
	Colors []string `json:"colors"`

	ShopUrl string `json:"shopUrl"`

	CreateAt  *time.Time `json:"createdAt,omitempty"`
	UpdatedAt *time.Time `json:"updatedAt,omitempty"`
	DeleteAt  *time.Time `json:"deletedAt,omitempty"`

	ctx context.Context
}

type SearchProductList []*Product
type CartProductList []*Product

func (l SearchProductList) Render(w http.ResponseWriter, r *http.Request) error {
	for _, v := range l {
		if err := v.Render(w, r); err != nil {
			return err
		}
	}
	return nil
}

func (l CartProductList) Render(w http.ResponseWriter, r *http.Request) error {
	for _, v := range l {
		if err := v.Render(w, r); err != nil {
			return err
		}
	}
	return nil
}

func NewSearchProductList(ctx context.Context, products []*data.Product) SearchProductList {
	list := SearchProductList{}
	for _, product := range products {
		list = append(list, NewProduct(ctx, product))
	}
	return list
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

	place, ok := p.ctx.Value("place").(*data.Place)
	if !ok {
		place, _ = data.DB.Place.FindByID(p.PlaceID)
	}
	p.Place = &Place{Place: place}

	if dbVariants, _ := data.DB.ProductVariant.FindByProductID(product.ID); dbVariants != nil {
		variants := make([]*ProductVariant, len(dbVariants))
		for i, v := range dbVariants {
			variants[i] = &ProductVariant{ProductVariant: v}
		}
		p.Variants = variants
	}

	p.Sizes = []string{}
	if sizes, _ := data.DB.ProductTag.FindAll(db.Cond{"product_id": product.ID, "type": data.ProductTagTypeSize}); sizes != nil {
		for _, s := range sizes {
			p.Sizes = append(p.Sizes, s.Value)
		}
	}

	p.Colors = []string{}
	if colors, _ := data.DB.ProductTag.FindAll(db.Cond{"product_id": product.ID, "type": data.ProductTagTypeColor}); colors != nil {
		for _, c := range colors {
			p.Colors = append(p.Colors, c.Value)
		}
	}

	return p
}

func (p *Product) Render(w http.ResponseWriter, r *http.Request) error {
	var u *url.URL
	if p.Place.ShopifyID != "" {
		u = &url.URL{
			Host: fmt.Sprintf("%s.myshopify.com", p.Place.ShopifyID),
		}
	} else if p.Place.Website != "" {
		u, _ = url.Parse(p.Place.Website)
	}

	u.Scheme = "https"
	u.Path = fmt.Sprintf("products/%s", p.ExternalID)

	p.ShopUrl = u.String()

	return nil
}

type ProductCategory struct {
	*data.ProductTag
	ImageURL string `json:"imageUrl"`

	ID        interface{} `json:"id,omitempty"`
	ProductID interface{} `json:"productId,omitempty"`
	Type      interface{} `json:"type,omitempty"`
	CreatedAt interface{} `json:"createdAt,omitempty"`
}

func (c *ProductCategory) Render(w http.ResponseWriter, r *http.Request) error {
	if tag, _ := data.DB.ProductTag.FindOne(db.Cond{"value": c.Value}); tag != nil {
		if product, _ := data.DB.Product.FindByID(tag.ProductID); product != nil {
			c.ImageURL = product.ImageUrl
		}
	}
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

func NewProductCategoryList(ctx context.Context, tags []*data.ProductTag) []render.Renderer {
	list := []render.Renderer{}
	for _, tag := range tags {
		list = append(list, &ProductCategory{ProductTag: tag})
	}
	return list
}
