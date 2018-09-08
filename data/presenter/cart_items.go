package presenter

import (
	"context"
	"net/http"

	"github.com/go-chi/render"

	"bitbucket.org/moodie-app/moodie-api/data"
)

type CartItem struct {
	*data.CartItem

	Cart    *Cart           `json:"cart,omitempty"`
	Place   *Place          `json:"-"`
	Product *Product        `json:"product,omitempty"`
	Variant *ProductVariant `json:"variant,omitempty"`

	Price    float64 `json:"price"`
	HasError bool    `json:"hasError"`
	Err      string  `json:"error"`

	ctx context.Context
}

func NewCartItem(ctx context.Context, item *data.CartItem) *CartItem {
	resp := &CartItem{
		CartItem: item,
		ctx:      ctx,
	}
	if variant, _ := data.DB.ProductVariant.FindByID(item.VariantID); variant != nil {
		resp.Variant = NewProductVariant(ctx, variant)
	}
	if product, _ := data.DB.Product.FindByID(item.ProductID); product != nil {
		resp.Product = NewProduct(ctx, product)
		resp.Variant.Place = resp.Product.Place.Place
	}
	return resp
}

func (i *CartItem) Render(w http.ResponseWriter, r *http.Request) error {
	if i.Product != nil {
		i.Product.Render(w, r)
	}
	if i.Variant != nil {
		i.Variant.Render(w, r)
		i.Price = i.Variant.Price
	}
	return nil
}

type CartItemList []*CartItem

func NewCartItemList(ctx context.Context, cartItems []*data.CartItem) []render.Renderer {
	list := []render.Renderer{}
	for _, item := range cartItems {
		list = append(list, NewCartItem(ctx, item))
	}
	return list
}

func (*CartItemList) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}
