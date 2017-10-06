package presenter

import (
	"context"
	"net/http"

	"github.com/pressly/chi/render"

	"bitbucket.org/moodie-app/moodie-api/data"
)

type CartShippingMethod struct {
	*data.CartShippingMethod
}

func NewCartShippingMethod(ctx context.Context, m *data.CartShippingMethod) *CartShippingMethod {
	return &CartShippingMethod{m}
}

func (*CartShippingMethod) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

type CartShippingMethodList []*CartShippingMethod

func NewCartShippingMethodList(ctx context.Context, methods []*data.CartShippingMethod) []render.Renderer {
	list := []render.Renderer{}
	for _, item := range methods {
		list = append(list, NewCartShippingMethod(ctx, item))
	}
	return list
}

func (*CartShippingMethodList) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}
