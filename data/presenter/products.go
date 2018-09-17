package presenter

import (
	"context"
	"net/http"
	"sort"

	"github.com/go-chi/render"
	set "gopkg.in/fatih/set.v0"

	db "upper.io/db.v3"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/data/stash"
	"bitbucket.org/moodie-app/moodie-api/lib/apparelsorter"
	"bitbucket.org/moodie-app/moodie-api/lib/htmlx"
	xchange "bitbucket.org/moodie-app/moodie-api/lib/xchanger"
)

type Product struct {
	*data.Product

	Variants []*ProductVariant    `json:"variants"`
	Place    *Place               `json:"place"`
	Images   []*data.ProductImage `json:"images"`

	ImageURL string   `json:"imageUrl"`
	Sizes    []string `json:"sizes"`
	Colors   []string `json:"colors"`

	// potentially (deal) modified price.
	Price float64 `json:"price"`

	ViewCount     int64 `json:"views"`
	PurchaseCount int64 `json:"purchased"`
	LiveViewCount int64 `json:"liveViews"`

	IsFavourite     bool `json:"isFavourite"`
	HasFreeShipping bool `json:"hasFreeShipping,omitempty"`

	CreateAt  interface{} `json:"createdAt,omitempty"`
	UpdatedAt interface{} `json:"updatedAt,omitempty"`
	DeleteAt  interface{} `json:"deletedAt,omitempty"`

	HtmlDescription  string      `json:"htmlDescription"`
	NoTagDescription string      `json:"noTagDescription"`
	Description      interface{} `json:"description"`

	ctx context.Context
}

type ProductEvent struct {
	*data.Product
	// the session user who viewed the product
	ViewerID int64 `json:"viewerId"`
	BuyerID  int64 `json:"buyerId"`
}

type ProductCache map[int64]*data.Product
type VariantCache map[int64][]*ProductVariant
type VariantImageCache map[int64]int64
type PlaceCache map[int64]*data.Place
type ImageCache map[int64][]*data.ProductImage
type FavoriteCache map[int64]bool
type ShippingZoneCache map[int64]bool

const PlaceCacheCtxKey = "place.cache"

func newProductList(ctx context.Context, products []*data.Product) []*Product {
	list := []*Product{}

	productCache := make(ProductCache)
	productIDSet := set.New()
	placeIDset := set.New()
	for _, product := range products {
		productIDSet.Add(int(product.ID))
		placeIDset.Add(int(product.PlaceID))
		productCache[product.ID] = product
	}

	if productIDSet.Size() == 0 {
		return list
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
	ctx = context.WithValue(ctx, PlaceCacheCtxKey, placeCache)

	// fetch free shipping zones on merchant
	zones, _ := data.DB.ShippingZone.FindAll(db.Cond{
		"place_id": set.IntSlice(placeIDset),
		"type":     data.ShippingZoneTypeByPrice,
		"price":    db.Eq(0),
	})
	zoneCache := make(ShippingZoneCache)
	for _, n := range zones {
		zoneCache[n.PlaceID] = true
	}

	// fetch product variants
	variants, _ := data.DB.ProductVariant.FindAll(db.Cond{
		"product_id": set.IntSlice(productIDSet),
	})
	variantCache := make(VariantCache)
	variantIDSet := set.New()
	for _, v := range variants {
		vv := NewProductVariant(ctx, v)
		vv.Product = productCache[v.ProductID]
		vv.Place = placeCache[vv.Product.PlaceID]
		variantCache[v.ProductID] = append(
			variantCache[v.ProductID],
			vv,
		)
		variantIDSet.Add(int(v.ID))
	}

	// fetch variant image pivotes
	// NOTE: IMG -> 1:M -> VAR
	variantImageCache := make(VariantImageCache)
	variantImages, _ := data.DB.VariantImage.FindAll(db.Cond{
		"variant_id": set.IntSlice(variantIDSet),
	})
	for _, v := range variantImages {
		variantImageCache[v.VariantID] = v.ImageID
	}

	// fetch product images
	images, _ := data.DB.ProductImage.FindAll(db.Cond{
		"product_id": set.IntSlice(productIDSet),
	})
	imageCache := make(ImageCache)
	for _, v := range images {
		imageCache[v.ProductID] = append(imageCache[v.ProductID], v)
	}

	// fetch users favourites
	favouriteCache := make(FavoriteCache)
	if user, ok := ctx.Value("session.user").(*data.User); ok {
		favouriteProducts, _ := data.DB.FavouriteProduct.FindAll(db.Cond{
			"user_id":    user.ID,
			"product_id": set.IntSlice(productIDSet),
		})
		for _, f := range favouriteProducts {
			favouriteCache[f.ProductID] = true
		}
	}

	ctx = context.WithValue(ctx, "variant.cache", variantCache)
	ctx = context.WithValue(ctx, "variant.image.cache", variantImageCache)
	ctx = context.WithValue(ctx, "image.cache", imageCache)
	ctx = context.WithValue(ctx, "favourite.cache", favouriteCache)
	ctx = context.WithValue(ctx, "shippingzone.cache", zoneCache)

	for _, product := range products {
		list = append(list, NewProduct(ctx, product))
	}
	return list
}

func NewProductList(ctx context.Context, products []*data.Product) []render.Renderer {
	list := []render.Renderer{}
	for _, p := range newProductList(ctx, products) {
		list = append(list, p)
	}
	return list
}

func NewProduct(ctx context.Context, product *data.Product) *Product {
	p := &Product{
		Product: product,
		Price:   product.Price,
		ctx:     ctx,
	}

	if cache, _ := ctx.Value("variant.cache").(VariantCache); cache != nil {
		p.Variants = cache[p.ID]
	} else {
		if vv, _ := data.DB.ProductVariant.FindByProductID(product.ID); vv != nil {
			for _, v := range vv {
				p.Variants = append(p.Variants, NewProductVariant(ctx, v))
			}
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

	// product image thumb
	for _, v := range p.Variants {
		for i, vi := range p.Images {
			if i == 0 {
				p.ImageURL = vi.ImageURL
			} else if vi.ID == v.ImageID {
				p.ImageURL = vi.ImageURL
				break
			}
		}

		// use the first variant as the preview image
		break
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

	if cache, _ := ctx.Value("variant.cache").(VariantCache); cache != nil {
		p.Variants = cache[p.ID]
	} else {
		if variants, _ := data.DB.ProductVariant.FindByProductID(product.ID); variants != nil {
			for _, v := range variants {
				vv := NewProductVariant(ctx, v)
				vv.Product = p.Product
				vv.Place = p.Place.Place
				p.Variants = append(p.Variants, vv)
			}
		}
	}

	if cache, _ := ctx.Value("favourite.cache").(FavoriteCache); cache != nil {
		p.IsFavourite = cache[p.ID]
	} else {
		if user, ok := ctx.Value("session.user").(*data.User); ok {
			p.IsFavourite, _ = data.DB.FavouriteProduct.Find(db.Cond{
				"product_id": p.ID,
				"user_id":    user.ID,
			}).Exists()
		}
	}

	if cache, _ := ctx.Value("shippingzone.cache").(ShippingZoneCache); cache != nil {
		p.HasFreeShipping, _ = cache[p.PlaceID]
	} else {
		p.HasFreeShipping, _ = data.DB.ShippingZone.Find(db.Cond{
			"place_id": p.PlaceID,
			"type":     data.ShippingZoneTypeByPrice,
			"price":    db.Eq(0),
		}).Exists()
	}

	p.ViewCount, _ = stash.GetProductViews(p.ID)
	p.LiveViewCount, _ = stash.GetProductLiveViews(p.ID)
	p.PurchaseCount, _ = stash.GetProductPurchases(p.ID)

	// modify product price if deal is active
	if deal, ok := ctx.Value(DealCtxKey).(*data.Deal); ok {
		// NOTE: deal value here is negative because the type is fixed amount only for now
		p.Price += deal.Value
	}

	return p
}

func (p *Product) Render(w http.ResponseWriter, r *http.Request) error {
	p.HtmlDescription = htmlx.CaptionizeHtmlBody(p.Product.Description, -1)
	p.NoTagDescription = htmlx.StripTags(p.Product.Description)

	for _, v := range p.Variants {
		v.Render(w, r)
	}

	// conver to usd if applicable
	if p.Place != nil && p.Place.Currency != "USD" {
		p.Price = xchange.ToUSD(p.Price, p.Place.Place.Currency)
	}
	p.Place.Currency = "USD" // NOTE: change it up to USD

	return nil
}
