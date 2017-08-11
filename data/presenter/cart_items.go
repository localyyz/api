package presenter

import (
	"context"
	"net/http"

	"bitbucket.org/moodie-app/moodie-api/data"
	"github.com/pressly/chi/render"
)

type CartItem struct {
	*data.CartItem

	Cart    *Cart           `json:"cart,omitempty"`
	Place   *Place          `json:"place,omitempty"`
	Product *Product        `json:"product,omitempty"`
	Variant *ProductVariant `json:"variant,omitempty"`

	Price float64 `json:"price"`

	ctx context.Context
}

func NewCartItem(ctx context.Context, item *data.CartItem) *CartItem {
	resp := &CartItem{
		CartItem: item,
		ctx:      ctx,
	}
	if product, _ := data.DB.Product.FindByID(item.ProductID); product != nil {
		resp.Product = NewProduct(ctx, product)
		resp.Product.Render(nil, nil)
	}
	if variant, _ := data.DB.ProductVariant.FindByID(item.VariantID); variant != nil {
		resp.Variant = &ProductVariant{ProductVariant: variant}
		resp.Price = variant.Etc.Price
	}
	return resp
}

func (i *CartItem) Render(w http.ResponseWriter, r *http.Request) error {
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
