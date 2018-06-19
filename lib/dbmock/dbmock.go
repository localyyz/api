package dbmock

import (
	"context"

	"bitbucket.org/moodie-app/moodie-api/data"
)

func NewDB(ctx context.Context) *data.Database {
	t := NewTable(ctx)
	return &data.Database{
		Product:        data.ProductStore{t},
		ProductVariant: data.ProductVariantStore{t},
		ProductImage:   data.ProductImageStore{t},
		VariantImage:   data.VariantImageStore{t},
		Place:          data.PlaceStore{t},
		Cart:           data.CartStore{t},
		CartItem:       data.CartItemStore{t},
		Checkout:       data.CheckoutStore{t},
	}
}
