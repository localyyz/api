package presenter

import (
	"context"
	"net/http"
	"sort"
	"time"

	"github.com/go-chi/render"
	set "gopkg.in/fatih/set.v0"

	db "upper.io/db.v3"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/lib/apparelsorter"
	"bitbucket.org/moodie-app/moodie-api/lib/htmlx"
)

type Product struct {
	*data.Product
	Variants []*ProductVariant    `json:"variants"`
	Place    *Place               `json:"place"`
	Images   []*data.ProductImage `json:"images"`

	Sizes  []string `json:"sizes"`
	Colors []string `json:"colors"`

	CreateAt  *time.Time `json:"createdAt,omitempty"`
	UpdatedAt *time.Time `json:"updatedAt,omitempty"`
	DeleteAt  *time.Time `json:"deletedAt,omitempty"`

	HtmlDescription  string      `json:"htmlDescription"`
	NoTagDescription string      `json:"noTagDescription"`
	Description      interface{} `json:"description"`

	ctx context.Context
}

type VariantCache map[int64][]*ProductVariant
type VariantImageCache map[int64]int64
type PlaceCache map[int64]*data.Place
type ImageCache map[int64][]*data.ProductImage
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

	productIDSet := set.New()
	placeIDset := set.New()
	for _, product := range products {
		productIDSet.Add(int(product.ID))
		placeIDset.Add(int(product.PlaceID))
	}

	// fetch product variants
	variants, _ := data.DB.ProductVariant.FindAll(db.Cond{
		"product_id": set.IntSlice(productIDSet),
	})
	variantCache := make(VariantCache)
	variantIDSet := set.New()
	for _, v := range variants {
		variantCache[v.ProductID] = append(variantCache[v.ProductID], &ProductVariant{ProductVariant: v})
		variantIDSet.Add(int(v.ID))
	}

	// fetch variant image pivotes
	variantImageCache := make(VariantImageCache)
	variantImages, _ := data.DB.VariantImage.FindAll(db.Cond{
		"variant_id": set.IntSlice(variantIDSet),
	})
	for _, v := range variantImages {
		variantImageCache[v.VariantID] = v.ImageID
	}

	// fetch places
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

	// fetch product images
	images, _ := data.DB.ProductImage.FindAll(db.Cond{
		"product_id": set.IntSlice(productIDSet),
	})
	imageCache := make(ImageCache)
	for _, v := range images {
		imageCache[v.ProductID] = append(imageCache[v.ProductID], v)
	}

	ctx = context.WithValue(ctx, "variant.cache", variantCache)
	ctx = context.WithValue(ctx, "variant.image.cache", variantImageCache)
	ctx = context.WithValue(ctx, "image.cache", imageCache)
	ctx = context.WithValue(ctx, "place.cache", placeCache)

	for _, product := range products {
		presented := NewProduct(ctx, product)
		if len(presented.Variants) > 0 && len(presented.Images) > 0 {
			list = append(list, presented)
		}
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

	var variants []*ProductVariant
	if cache, _ := ctx.Value("variant.cache").(VariantCache); cache != nil {
		variants = cache[p.ID]
	} else {
		if vv, _ := data.DB.ProductVariant.FindByProductID(product.ID); vv != nil {
			for _, v := range vv {
				variants = append(p.Variants, &ProductVariant{ProductVariant: v})
			}
		}
	}

	// filter variants for ones that are in stock
	p.Variants = variants[:0]
	for _, v := range variants {
		if v.Limits > 0 {
			p.Variants = append(p.Variants, v)
		}
	}

	sizeSet := set.New()
	colorSet := set.New()
	for _, v := range p.Variants {
		if v.Etc.Color != "" {
			colorSet.Add(v.Etc.Color)
		}
		if v.Etc.Size != "" {
			sizeSet.Add(v.Etc.Size)
		}
	}

	// check and load product images
	if cache, _ := ctx.Value("image.cache").(ImageCache); cache != nil {
		p.Images = cache[p.ID]
	} else {
		p.Images, _ = data.DB.ProductImage.FindByProductID(p.ID)
	}
	for i, img := range p.Images {
		if i == 0 {
			if len(p.ImageURL) == 0 {
				p.ImageURL = img.ImageURL
			}
		}
		// TODO: backwards compart + REMOVE
		if p.Etc.Images == nil {
			p.Etc.Images = append(p.Etc.Images, img.ImageURL)
		}
	}
	if len(p.Images) == 0 {
		for i, imgURL := range p.Etc.Images {
			p.Images = append(
				p.Images,
				&data.ProductImage{ImageURL: imgURL, Ordering: int32(i)},
			)
		}
	}

	// update product variants with the variant image pivot value
	for _, v := range p.Variants {
		if cache, _ := ctx.Value("variant.image.cache").(VariantImageCache); cache != nil {
			v.ImageID = cache[v.ID]
		} else {
			if img, _ := data.DB.VariantImage.FindByVariantID(v.ID); img != nil {
				v.ImageID = img.ImageID
			}
		}
	}

	sizesorter := apparelsorter.New(set.StringSlice(sizeSet)...)
	sort.Sort(sizesorter)
	p.Sizes = sizesorter.StringSlice()
	p.Colors = set.StringSlice(colorSet)

	if cache, _ := ctx.Value("place.cache").(PlaceCache); cache != nil {
		p.Place = &Place{Place: cache[p.PlaceID]}
	} else {
		// if place is not in the context, return it as part of presenter
		if _, ok := ctx.Value("place").(*data.Place); !ok {
			if place, _ := data.DB.Place.FindByID(p.PlaceID); place != nil {
				p.Place = &Place{Place: place}
			}
		}
	}

	return p
}

func (p *Product) Render(w http.ResponseWriter, r *http.Request) error {
	p.HtmlDescription = htmlx.CaptionizeHtmlBody(p.Product.Description, -1)
	p.NoTagDescription = htmlx.StripTags(p.Product.Description)

	return nil
}
