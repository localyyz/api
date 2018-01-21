package presenter

import (
	"context"
	"net/http"
	"time"

	"github.com/go-chi/render"
	set "gopkg.in/fatih/set.v0"

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

	ThumbURL         string      `json:"thumbUrl"`
	HtmlDescription  string      `json:"htmlDescription"`
	NoTagDescription string      `json:"noTagDescription"`
	Description      interface{} `json:"description"`

	ctx context.Context
}

type VariantCache map[int64][]*ProductVariant
type TagCache map[int64][]*data.ProductTag
type PlaceCache map[int64]*data.Place

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
	return newProductList(ctx, products)
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

func newProductList(ctx context.Context, products []*data.Product) []render.Renderer {
	list := []render.Renderer{}

	productIDs := make([]int64, len(products))
	placeIDset := set.New()
	for i, product := range products {
		productIDs[i] = product.ID
		placeIDset.Add(int(product.PlaceID))
	}

	// fetch product variants
	variantCache := make(VariantCache)
	if variants, _ := data.DB.ProductVariant.FindAll(db.Cond{"product_id": productIDs}); variants != nil {
		for _, v := range variants {
			variantCache[v.ProductID] = append(variantCache[v.ProductID], &ProductVariant{ProductVariant: v})
		}
	}

	placeCache := make(PlaceCache)
	if place, _ := ctx.Value("place").(*data.Place); place != nil {
		placeCache[place.ID] = place
	} else {
		if places, _ := data.DB.Place.FindAll(db.Cond{"id": set.IntSlice(placeIDset)}); places != nil {
			for _, p := range places {
				placeCache[p.ID] = p
			}
		}
	}

	//ctx = context.WithValue(ctx, "tag.cache", tagCache)
	ctx = context.WithValue(ctx, "variant.cache", variantCache)
	ctx = context.WithValue(ctx, "place.cache", placeCache)

	for _, product := range products {
		list = append(list, NewProduct(ctx, product))
	}
	return list
}

func NewProductList(ctx context.Context, products []*data.Product) []render.Renderer {
	return newProductList(ctx, products)
}

func NewProduct(ctx context.Context, product *data.Product) *Product {
	p := &Product{
		Product: product,
		ctx:     ctx,
	}

	p.Variants = []*ProductVariant{}
	if cache, _ := ctx.Value("variant.cache").(VariantCache); cache != nil {
		p.Variants = cache[p.ID]
	} else {
		if dbVariants, _ := data.DB.ProductVariant.FindByProductID(product.ID); dbVariants != nil {
			for _, v := range dbVariants {
				p.Variants = append(p.Variants, &ProductVariant{ProductVariant: v})
			}
		}
	}

	sizeSet := set.New()
	colorSet := set.New()
	for _, v := range p.Variants {
		colorSet.Add(v.Etc.Color)
		sizeSet.Add(v.Etc.Size)
	}
	p.Sizes = set.StringSlice(sizeSet)
	p.Colors = set.StringSlice(colorSet)

	if cache, _ := ctx.Value("place.cache").(PlaceCache); cache != nil {
		p.Place = cache[p.PlaceID]
	} else {
		// if place is not in the context, return it as part of presenter
		if _, ok := ctx.Value("place").(*data.Place); !ok {
			if place, _ := data.DB.Place.FindByID(p.PlaceID); place != nil {
				p.Place = place
			}
		}
	}

	return p
}

func (p *Product) Render(w http.ResponseWriter, r *http.Request) error {
	p.HtmlDescription = htmlx.CaptionizeHtmlBody(p.Product.Description, -1)
	p.NoTagDescription = htmlx.StripTags(p.Product.Description)
	p.ThumbURL = thumbImage(p.Product.ImageUrl)

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
