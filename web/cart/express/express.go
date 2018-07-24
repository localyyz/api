package express

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/data/presenter"
	"bitbucket.org/moodie-app/moodie-api/lib/connect"
	"bitbucket.org/moodie-app/moodie-api/lib/events"
	"bitbucket.org/moodie-app/moodie-api/lib/shopify"
	"bitbucket.org/moodie-app/moodie-api/web/api"
	"github.com/go-chi/render"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/pressly/lg"
	"upper.io/db.v3"
)

type CartItemRequest struct {
	ProductID int64  `json:"productId"`
	Color     string `json:"color"`
	Size      string `json:"size"`
	Quantity  uint32 `json:"quantity"`
}

func (*CartItemRequest) Bind(r *http.Request) error {
	return nil
}

type expressCheckoutResponse struct {
	*presenter.Cart
	Rates []*data.CartShippingMethod `json:"rates"`
}

func GetCart(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	cart := ctx.Value("cart").(*data.Cart)
	render.Render(w, r, presenter.NewCart(ctx, cart))
}

func CreateCartItem(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	cart := ctx.Value("cart").(*data.Cart)
	user := ctx.Value("session.user").(*data.User)

	var payload CartItemRequest
	if err := render.Bind(r, &payload); err != nil {
		render.Render(w, r, api.ErrInvalidRequest(err))
		return
	}

	/*checking if the user is attempting to add an item from an expired deal*/
	var deal *data.Collection

	// TODO: make this better
	err := data.DB.Select("c.*").
		From("collections as c").
		LeftJoin("collection_products as cp").On("c.id = cp.collection_id").
		Where(db.Cond{
			"c.lightning":   true,
			"cp.product_id": payload.ProductID,
		}).
		One(&deal)
	if err != nil && err != db.ErrNoMoreRows {
		render.Respond(w, r, err)
		return
	}

	//product is part of a lightning collection
	if deal != nil {
		// check if the user has already  purchased
		userPurchased, _ := user.GetTotalCheckout(payload.ProductID)
		if userPurchased > 0 { //user has a limit of one purchase per deal
			render.Render(w, r, api.ErrMultiplePurchase)
			return
		}

		//the cron might not be in sync therefore we need to check percentage completion as well
		totalCheckouts, _ := deal.GetCheckoutCount()
		if int64(totalCheckouts) == deal.Cap {
			render.Render(w, r, api.ErrLightningOutOfStock)
			return
		}

		// check if user deal exists + active
		userDeal, _ := data.DB.UserDeal.FindOne(db.Cond{
			"deal_id": deal.ID,
			"user_id": user.ID,
			"status":  data.CollectionStatusActive,
		})

		// expire if:
		// - the deal is expired and no user deal
		// - the deal is expired and the user deal has expired
		if (deal.Status != data.CollectionStatusActive && userDeal == nil) ||
			(userDeal != nil && userDeal.Status != data.CollectionStatusActive) {
			render.Render(w, r, api.ErrExpiredDeal)
			return
		}
	}

	cond := db.Cond{
		"product_id": payload.ProductID,
		"limits >=":  1,
	}
	if len(payload.Color) > 0 {
		cond[db.Raw("lower(etc->>'color')")] = payload.Color
	}
	if len(payload.Size) > 0 {
		cond[db.Raw("lower(etc->>'size')")] = payload.Size
	}

	// fetch the variant from given payload (product id, color and size)
	variant, err := data.DB.ProductVariant.FindOne(cond)
	if err != nil {
		if err == db.ErrNoMoreRows {
			render.Render(w, r, api.ErrOutOfStockAdd(err))
			return
		}
		render.Render(w, r, api.ErrInvalidRequest(err))
		return
	}

	toSave := &data.CartItem{
		CartID:    cart.ID,
		ProductID: payload.ProductID,
		VariantID: variant.ID,
		PlaceID:   variant.PlaceID,
		Quantity:  uint32(payload.Quantity)}

	// find the shopify cred for the merchant and start the checkout process
	cred, err := data.DB.ShopifyCred.FindByPlaceID(variant.PlaceID)
	if err != nil {
		render.Respond(w, r, err)
		return
	}

	shopifyCheckout := &shopify.Checkout{
		LineItems: []*shopify.LineItem{{VariantID: variant.OfferID, Quantity: 1}},
	}
	if discount, _ := data.DB.PlaceDiscount.FindByPlaceID(toSave.PlaceID); discount != nil {
		shopifyCheckout.DiscountCode = discount.Code
	}

	client := shopify.NewClient(nil, cred.AccessToken)
	client.BaseURL, _ = url.Parse(cred.ApiURL)
	//client.Debug = true
	checkout, _, err := client.Checkout.Create(
		ctx,
		shopifyCheckout,
	)
	if err != nil || checkout == nil {
		lg.Alertf("failed to create express checkout(%d) with err: %+v", cart.ID, err)
		render.Respond(w, r, err)
		return
	}
	cart.Etc = data.CartEtc{
		ShopifyData: map[int64]*data.CartShopifyData{
			variant.PlaceID: &data.CartShopifyData{
				Token:      checkout.Token,
				CustomerID: checkout.CustomerID,
				Name:       checkout.Name,
				ShopifyPaymentAccountID: checkout.ShopifyPaymentAccountID,
				PaymentURL:              checkout.PaymentURL,
				WebURL:                  checkout.WebURL,
				WebProcessingURL:        checkout.WebProcessingURL,
				TotalTax:                atoi(checkout.TotalTax),
				TotalPrice:              atoi(checkout.TotalPrice),
				PaymentDue:              checkout.PaymentDue,
				Discount:                checkout.AppliedDiscount,
				Currency:                checkout.Currency,
			},
		},
	}
	cart.Status = data.CartStatusCheckout
	data.DB.Cart.Save(cart)

	// save the cart item
	if err := data.DB.CartItem.Save(toSave); err != nil {
		lg.Alertf("failed to save express cart item(%d) on cart(%d) with err: %+v", variant.ID, cart.ID, err)
		render.Respond(w, r, err)
		return
	}
	lg.SetEntryField(ctx, "variant_id", variant.ID)

	render.Render(w, r, presenter.NewCartItem(ctx, toSave))
}

type CartPaymentRequest struct {
	BillingAddress      *data.CartAddress `json:"billingAddress"`
	ExpressPaymentToken string            `json:"expressPaymentToken"`
	Email               string            `json:"email,omitempty"`
}

func (p *CartPaymentRequest) Bind(r *http.Request) error {
	if p.BillingAddress == nil {
		lg.Warn("billing address")
		return errors.New("invalid billing address")
	}
	if len(p.ExpressPaymentToken) == 0 {
		return errors.New("invalid token")
	}
	if len(p.Email) == 0 {
		user := r.Context().Value("session.user").(*data.User)
		p.Email = user.Email
	}
	if len(p.Email) == 0 {
		return errors.New("email must be provided")
	}
	return nil
}

func CreatePayment(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	cart := ctx.Value("cart").(*data.Cart)
	client := ctx.Value("shopify.client").(*shopify.Client)

	lg.Infof("express cart(%d) start payment", cart.ID)
	var payload CartPaymentRequest
	if err := render.Bind(r, &payload); err != nil {
		render.Render(w, r, api.ErrInvalidRequest(err))
		return
	}

	var (
		placeID     int64
		shopifyData *data.CartShopifyData
	)
	// for express cart, there's only 1 item for 1 merchant
	for pID, sh := range cart.Etc.ShopifyData {
		placeID = pID
		shopifyData = sh
	}

	cc, _, err := client.Checkout.Update(
		ctx,
		&shopify.Checkout{
			Token: shopifyData.Token,
			BillingAddress: &shopify.CustomerAddress{
				Address1:  payload.BillingAddress.Address,
				Address2:  payload.BillingAddress.AddressOpt,
				City:      payload.BillingAddress.City,
				Country:   payload.BillingAddress.Country,
				FirstName: payload.BillingAddress.FirstName,
				LastName:  payload.BillingAddress.LastName,
				Province:  payload.BillingAddress.Province,
				Zip:       payload.BillingAddress.Zip,
			},
			Email: payload.Email,
		},
	)
	if err != nil {
		lg.Alert(errors.Wrapf(err, "failed to pay cart(%d). shopify(%v)", cart.ID, placeID))
		return
	}

	u, _ := uuid.NewUUID()
	payment := &shopify.Payment{
		Amount:      cc.PaymentDue,
		UniqueToken: u.String(),
		PaymentToken: &shopify.PaymentToken{
			PaymentData: payload.ExpressPaymentToken,
			Type:        shopify.StripeVaultToken,
		},
		RequestDetails: &shopify.RequestDetail{
			IPAddress: r.RemoteAddr,
		},
	}

	// 4. send payment to shopify
	p, _, err := client.Checkout.Payment(ctx, shopifyData.Token, payment)
	if err != nil {
		lg.Alertf("payment fail: cart(%d) place(%d) with err %+v", cart.ID, placeID, err)
		render.Respond(w, r, err)
		// TODO: do we return here?
		return
	}

	// check payment transaction
	if p.Transaction == nil {
		// something failed. Try again
		lg.Alertf("cart(%d) failed with empty transaction response", cart.ID)
		render.Respond(w, r, errors.New("payment failed, please try again"))
		return
	}

	if p.Transaction.Status != shopify.TransactionStatusSuccess {
		// something failed. Try again
		lg.Alertf("cart(%d) failed with transaction status %v", cart.ID, p.Transaction.Status)
		render.Respond(w, r, api.ErrCardVaultProcess(fmt.Errorf(p.Transaction.Message)))
		return
	}

	lg.Alertf("express cart(%d) was just paid!", cart.ID)

	// 5. save shopify payment id
	shopifyData.PaymentID = p.ID
	shopifyData.PaymentDue = cc.PaymentDue
	shopifyData.TotalTax = atoi(cc.TotalTax)
	shopifyData.TotalPrice = atoi(cc.TotalPrice)

	// TODO: sync user email and name
	cart.Etc.BillingAddress = payload.BillingAddress
	cart.Etc.BillingAddress.Email = payload.Email

	// mark checkout as has payed
	cart.Status = data.CartStatusPaymentSuccess
	if err := data.DB.Cart.Save(cart); err != nil {
		lg.Alertf("express cart (%d) payment save failed with %+v", cart.ID, err)
	}
	// TODO: create a customer on stripe after the first
	// tokenization so we can send stripe customer id moving forward

	// upgrade user to a full user
	user := ctx.Value("session.user").(*data.User)
	newUser := &data.User{
		ID:          user.ID,
		Username:    user.Email,
		Email:       user.Email,
		DeviceToken: &(user.Username),
		Name:        fmt.Sprintf("%s %s", payload.BillingAddress.FirstName, payload.BillingAddress.LastName),
		Network:     "shadow",
		LastLogInAt: data.GetTimeUTCPointer(),
		LoggedIn:    true,
		Etc:         data.UserEtc{},
	}
	// TODO: how does the user login here??
	data.DB.User.Save(newUser)

	// fetch the product from the checkout

	go func() {
		cartItem, err := data.DB.CartItem.FindOne(db.Cond{"cart_id": cart.ID})
		if err != nil {
			return
		}
		product, err := data.DB.Product.FindByID(cartItem.ProductID)
		if err != nil {
			return
		}
		connect.NATS.Emit(events.EvProductPurchased, &presenter.ProductEvent{
			Product: product,
			BuyerID: user.ID,
		})
	}()

	render.Render(w, r, presenter.NewCart(ctx, cart))
}

func DeleteCart(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	cart := ctx.Value("cart").(*data.Cart)

	if itemCount, _ := data.DB.CartItem.Find(
		db.Cond{"cart_id": cart.ID},
	).Count(); itemCount == 0 {
		render.Status(r, http.StatusNoContent)
		render.Respond(w, r, "")
	}

	cart.DeletedAt = data.GetTimeUTCPointer()
	if err := data.DB.Cart.Delete(cart); err != nil {
		render.Respond(w, r, err)
		return
	}

	render.Status(r, http.StatusNoContent)
	render.Respond(w, r, "")
}

func atoi(s string) int64 {
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		lg.Errorf("failed to parse %s to float", s)
		return 0
	}
	return int64(f * 100.0)
}
