package shopify

import (
	"net/http"

	"bitbucket.org/moodie-app/moodie-api/data"
	lib "bitbucket.org/moodie-app/moodie-api/lib/shopify"
	"bitbucket.org/moodie-app/moodie-api/web/api"
	"github.com/go-chi/render"
	"github.com/pressly/lg"
)

type checkoutWrapper struct {
	*lib.Checkout
}

func (w *checkoutWrapper) Bind(r *http.Request) error {
	return nil
}

func CheckoutHandler(r *http.Request) error {
	ctx := r.Context()

	wrapper := new(checkoutWrapper)
	if err := render.Bind(r, wrapper); err != nil {
		return api.ErrInvalidRequest(err)
	}
	defer r.Body.Close()

	place := ctx.Value("place").(*data.Place)

	iterator := data.DB.IteratorContext(ctx, `
	SELECT id
	FROM (
	SELECT id, (jsonb_each(etc->'shopifyData')).value->>'token' as token
	FROM carts
	) x
	WHERE x.token = ?`, wrapper.Token)
	defer iterator.Close()

	var cartID int64
	if err := iterator.ScanOne(&cartID); err != nil || iterator.Err() != nil {
		return err
	}

	cart, err := data.DB.Cart.FindByID(cartID)
	if err != nil {
		return err
	}

	lg.Alertf("webhook: cart(%d) for place(%d) was updated", cart.ID, place.ID)
	return nil
}
