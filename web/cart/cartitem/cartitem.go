package cartitem

import (
	"fmt"
	"net/http"
	"strings"

	db "upper.io/db.v3"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/data/presenter"
	"bitbucket.org/moodie-app/moodie-api/lib/connect"
	"bitbucket.org/moodie-app/moodie-api/lib/events"
	"bitbucket.org/moodie-app/moodie-api/lib/shopify"
	"bitbucket.org/moodie-app/moodie-api/web/api"
	"github.com/go-chi/render"
	"github.com/pkg/errors"
	"github.com/pressly/lg"
)

type CartItemRequest struct {
	ProductID *int64 `json:"productId,omitempty"`
	Color     string `json:"color"`
	Size      string `json:"size"`
	Quantity  uint32 `json:"quantity"`

	VariantID *int64 `json:"variantId,omitempty"`
}

func (c *CartItemRequest) Bind(r *http.Request) error {
	if c.ProductID == nil && c.VariantID == nil {
		return errors.New("invalid add item")
	}
	if c.Quantity < 1 {
		c.Quantity = 1
	}
	c.Color = strings.ToLower(c.Color)
	c.Size = strings.ToLower(c.Size)
	return nil
}

func CreateCartItem(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	cart := ctx.Value("cart").(*data.Cart)

	var payload CartItemRequest
	if err := render.Bind(r, &payload); err != nil {
		render.Render(w, r, api.ErrInvalidRequest(err))
		return
	}

	switch cart.Status {
	case data.CartStatusInProgress:
	case data.CartStatusCheckout:
		cart.Status = data.CartStatusInProgress
		if err := data.DB.Cart.Save(cart); err != nil {
			render.Render(w, r, api.ErrInvalidRequest(err))
			return
		}
	default:
		render.Render(w, r, api.ErrInvalidRequest(errors.New("invalid cart")))
		return
	}

	var (
		variant *data.ProductVariant
		err     error
	)
	if payload.VariantID != nil {
		variant, err = data.DB.ProductVariant.FindByID(*payload.VariantID)
	} else {
		// TODO: remove, here for bkwards compat
		variant, err = data.DB.ProductVariant.FindOne(
			db.Cond{
				"product_id":                   payload.ProductID,
				db.Raw("lower(etc->>'color')"): payload.Color,
				db.Raw("lower(etc->>'size')"):  payload.Size,
			},
		)
	}
	if err != nil {
		render.Render(w, r, api.ErrInvalidRequest(err))
		return
	}

	if variant.Limits == 0 {
		render.Render(w, r, api.ErrOutOfStockAdd(err))
		return
	}

	// check if an checkout exists
	checkout, err := data.DB.Checkout.FindOne(
		db.Cond{
			"cart_id":  cart.ID,
			"place_id": variant.PlaceID,
		},
	)
	if err != nil {
		if err == db.ErrNoMoreRows {
			checkout = &data.Checkout{
				CartID:  &cart.ID,
				UserID:  cart.UserID,
				PlaceID: variant.PlaceID,
				Status:  data.CheckoutStatusPending,
			}
			if err := data.DB.Checkout.Save(checkout); err != nil {
				lg.Alertf("checkout create: %s", err)
				render.Respond(w, r, err)
				return
			}
		} else {
			render.Respond(w, r, err)
			return
		}
	}

	// check if this product is part of a deal?
	var deal *data.Deal
	data.DB.Select("d.*").
		From("deals d").
		LeftJoin("deal_products dp").On("dp.deal_id = d.id").
		Where(db.Cond{"dp.product_id": variant.ProductID}).
		One(&deal)
	if deal != nil {
		// TODO: check deal usage limit
		// TODO: check once per customer limit
		if deal.Status == data.DealStatusActive {
			checkout.DiscountCode = deal.Code
			checkout.AppliedDiscount = &data.CheckoutAppliedDiscount{
				AppliedDiscount: &shopify.AppliedDiscount{
					Title:  deal.Code,
					Amount: fmt.Sprintf("%.2f", deal.Value),
				},
			}
		} else {
			// check if user deal exists + active
			userDeal, _ := data.DB.Deal.FindOne(db.Cond{
				"parent_id": deal.ID,
				"user_id":   cart.UserID,
				"status":    data.DealStatusActive,
			})
			if userDeal != nil && userDeal.Status == data.DealStatusActive {
				checkout.DiscountCode = userDeal.Code
				checkout.AppliedDiscount = &data.CheckoutAppliedDiscount{
					AppliedDiscount: &shopify.AppliedDiscount{
						Title:  deal.Code,
						Amount: fmt.Sprintf("%.2f", userDeal.Value),
					},
				}
			}
		}
		if err := data.DB.Checkout.Save(checkout); err != nil {
			render.Respond(w, r, err)
			return
		}
	}

	newItem := &data.CartItem{
		CartID:     cart.ID,
		ProductID:  variant.ProductID,
		VariantID:  variant.ID,
		CheckoutID: &checkout.ID,
		PlaceID:    variant.PlaceID,
		Quantity:   uint32(payload.Quantity),
	}
	if err := data.DB.CartItem.Save(newItem); err != nil {
		render.Respond(w, r, err)
		return
	}
	lg.SetEntryField(ctx, "variant_id", variant.ID)

	// emit event
	connect.NATS.Emit(
		events.EvProductAddedToCart,
		presenter.ProductEvent{
			Product:  &data.Product{ID: variant.ProductID},
			ViewerID: cart.UserID,
		},
	)

	render.Status(r, http.StatusCreated)
	render.Render(w, r, presenter.NewCartItem(ctx, newItem))
}

func RemoveCartItem(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	cart := ctx.Value("cart").(*data.Cart)
	cartItem := ctx.Value("cart_item").(*data.CartItem)

	// Remove the cart item
	if err := data.DB.CartItem.Delete(cartItem); err != nil {
		render.Respond(w, r, api.ErrInvalidRequest(err))
		return
	}

	// check if this is the last item in cart for a specific checkout
	numItems, err := data.DB.CartItem.Find(
		db.Cond{
			"cart_id":  cart.ID,
			"place_id": cartItem.PlaceID,
		},
	).Count()
	if err != nil {
		render.Respond(w, r, api.ErrInvalidRequest(err))
		return
	}

	if numItems == 0 && cartItem.CheckoutID != nil {
		if err := data.DB.Checkout.Delete(&data.Checkout{
			ID: *cartItem.CheckoutID,
		}); err != nil {
			lg.Alertf("checkout delete: %s", err)
		}
	}

	// Reset cart status to inProgress
	if cart.Status != data.CartStatusInProgress {
		cart.Status = data.CartStatusInProgress
	}
	if err := data.DB.Cart.Save(cart); err != nil {
		render.Render(w, r, api.ErrInvalidRequest(err))
		return
	}
	lg.SetEntryField(ctx, "variant_id", cartItem.VariantID)

	render.Status(r, http.StatusNoContent)
	render.Respond(w, r, "")
}
