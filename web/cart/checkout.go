package cart

import (
	"net/http"
	"net/url"

	"github.com/pressly/chi/render"
	db "upper.io/db.v3"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/data/presenter"
	"bitbucket.org/moodie-app/moodie-api/lib/shopify"
)

func Checkout(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := ctx.Value("session.user").(*data.User)
	cart := ctx.Value("cart").(*data.Cart)

	//if cart.Status == data.CartStatusComplete {
	//return
	//}

	// find cart items
	cartItems, err := data.DB.CartItem.FindByCartID(cart.ID)
	if err != nil {
		render.Respond(w, r, err)
		return
	}

	// shopify cart request objects
	checkoutMap := make(map[int64]*shopify.Checkout)
	variantMap := make(map[int64]*data.ProductVariant)
	productMap := make(map[int64]*data.Product)
	placeMap := make(map[int64]*data.Place)
	credMap := make(map[int64]*data.ShopifyCred)

	{ // figure out which place each cartItem belongs to
		var productIDs []int64
		var variantIDs []int64
		for _, item := range cartItems {
			productIDs = append(productIDs, item.ProductID)
			variantIDs = append(variantIDs, item.VariantID)
		}

		products, err := data.DB.Product.FindAll(db.Cond{"id": productIDs})
		if err != nil {
			render.Respond(w, r, err)
			return
		}

		variants, err := data.DB.ProductVariant.FindAll(db.Cond{"id": variantIDs})
		if err != nil {
			render.Respond(w, r, err)
			return
		}

		for _, v := range variants {
			variantMap[v.ID] = v
		}

		var placeIDs []int64
		for _, p := range products {
			placeIDs = append(placeIDs, p.PlaceID)
			productMap[p.ID] = p
		}

		places, err := data.DB.Place.FindAll(db.Cond{"id": placeIDs})
		if err != nil {
			render.Respond(w, r, err)
			return
		}

		// transform address
		address := &shopify.CustomerAddress{
			Address1:  cart.Etc.ShippingAddress.Address,
			Address2:  cart.Etc.ShippingAddress.AddressOpt,
			City:      cart.Etc.ShippingAddress.City,
			Country:   cart.Etc.ShippingAddress.Country,
			FirstName: cart.Etc.ShippingAddress.FirstName,
			LastName:  cart.Etc.ShippingAddress.LastName,
			Province:  cart.Etc.ShippingAddress.Province,
			Zip:       cart.Etc.ShippingAddress.Zip,
		}
		// TODO: for now, billing address is the same as shippingaddress

		for _, p := range places {
			placeMap[p.ID] = p
			checkoutMap[p.ID] = &shopify.Checkout{
				Email:           user.Email,
				ShippingAddress: address,
				BillingAddress:  address,
			}
		}

		creds, err := data.DB.ShopifyCred.FindAll(db.Cond{"place_id": placeIDs})
		if err != nil {
			render.Respond(w, r, err)
			return
		}

		for _, c := range creds {
			credMap[c.PlaceID] = c
		}
	}

	for _, item := range cartItems {
		product := productMap[item.ProductID]
		place := placeMap[product.PlaceID]
		variant := variantMap[item.VariantID]

		checkoutMap[place.ID].LineItems = append(
			checkoutMap[place.ID].LineItems,
			&shopify.LineItem{
				VariantID: variant.OfferID,
				Quantity:  int64(item.Quantity),
			},
		)
	}

	cart.Etc.ShopifyData = make(map[int64]*data.CartShopifyData)
	for placeID, checkout := range checkoutMap {
		cred := credMap[placeID]

		api := shopify.NewClient(nil, cred.AccessToken)
		api.BaseURL, _ = url.Parse(cred.ApiURL)

		cc, _, err := api.Checkout.Create(ctx, &shopify.CheckoutRequest{checkout})
		if err != nil {
			render.Respond(w, r, err)
			return
		}

		cart.Etc.ShopifyData[placeID] = &data.CartShopifyData{
			Token:            cc.Token,
			CustomerID:       cc.CustomerID,
			Name:             cc.Name,
			PaymentAccountID: cc.PaymentAccountID,
			WebURL:           cc.WebURL,
			WebProcessingURL: cc.WebProcessingURL,
			SubtotalPrice:    atoi(cc.SubtotalPrice),
			TotalPrice:       atoi(cc.TotalPrice),
			TotalTax:         atoi(cc.TotalTax),
			PaymentDue:       cc.PaymentDue,
		}
	}
	cart.Status = data.CartStatusProcessing
	if err := data.DB.Cart.Save(cart); err != nil {
		render.Respond(w, r, err)
		return
	}

	render.Render(w, r, presenter.NewCart(ctx, cart))
}
