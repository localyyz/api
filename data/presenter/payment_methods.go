package presenter

import (
	"context"
	"net/http"

	"github.com/go-chi/render"

	"bitbucket.org/moodie-app/moodie-api/data"
)

type PaymentMethod struct {
	*data.PaymentMethod
	ctx context.Context
}

func (c *PaymentMethod) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

func NewPaymentMethod(ctx context.Context, payment *data.PaymentMethod) *PaymentMethod {
	resp := &PaymentMethod{
		PaymentMethod: payment,
		ctx:           ctx,
	}
	return resp
}

type PaymentMethodList []*PaymentMethod

func (l PaymentMethodList) Render(w http.ResponseWriter, r *http.Request) error {
	for _, v := range l {
		if err := v.Render(w, r); err != nil {
			return err
		}
	}
	return nil
}

func NewPaymentMethodList(ctx context.Context, payments []*data.PaymentMethod) []render.Renderer {
	list := []render.Renderer{}
	for _, p := range payments {
		list = append(list, NewPaymentMethod(ctx, p))
	}
	return list
}
