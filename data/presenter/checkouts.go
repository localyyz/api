package presenter

import (
	"context"
	"net/http"

	"bitbucket.org/moodie-app/moodie-api/data"
)

type Checkout struct {
	*data.Checkout

	ctx context.Context
}

func NewCheckout(ctx context.Context, checkout *data.Checkout) *Checkout {
	return &Checkout{
		Checkout: checkout,
		ctx:      ctx,
	}
}

func (c *Checkout) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}
