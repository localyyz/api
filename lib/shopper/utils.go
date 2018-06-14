package shopper

import (
	"context"
	"strings"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/lib/connect"
	"bitbucket.org/moodie-app/moodie-api/lib/shopify"
	"bitbucket.org/moodie-app/moodie-api/lib/shopify/cardvault"
	"bitbucket.org/moodie-app/moodie-api/lib/stripe"
	"bitbucket.org/moodie-app/moodie-api/web/api"
	"github.com/pressly/lg"
)

const BillingAddressCtxKey = "shopper.billing.address"
const ShippingAddressCtxKey = "shopper.shipping.address"
const EmailCtxKey = "shipper.email"
const RequestIPCtxKey = "shopper.request.ip"
const PaymentCardCtxKey = "shopper.payment.card"

func toShopifyAddress(a *data.CartAddress) *shopify.CustomerAddress {
	return &shopify.CustomerAddress{
		Address1:  a.Address,
		Address2:  a.AddressOpt,
		City:      a.City,
		Country:   a.Country,
		FirstName: a.FirstName,
		LastName:  a.LastName,
		Province:  a.Province,
		Zip:       a.Zip,
	}
}

func NewVaultToken(ctx context.Context, p *Payment) (string, error) {
	expiryYM := strings.Split(p.card.Expiry, "/")
	nameParts := strings.Split(p.card.Name, " ")
	firstName := strings.Join(nameParts[0:len(nameParts)-1], " ")
	lastName := nameParts[len(nameParts)-1]
	cardParam := &cardvault.CreditCard{
		FirstName:         firstName,
		LastName:          lastName,
		Number:            p.card.Number,
		Month:             expiryYM[0],
		Year:              expiryYM[1],
		VerificationValue: p.card.CVC,
	}

	req := &cardvault.PaymentRequest{
		Payment: &cardvault.Payment{
			Amount:      p.checkout.PaymentDue,
			CreditCard:  cardParam,
			UniqueToken: p.uniqueID,
		},
	}
	// network call to shopify's cardvaulting service
	lg.Info("exchanging cardvault id")
	cardVaultID, _, err := cardvault.AddCard(ctx, req)
	if err != nil {
		return "", api.ErrCardVaultProcess(err)
	}
	lg.Info("received cardvault token")
	return cardVaultID, nil
}

func NewStripeToken(ctx context.Context, p *Payment) (string, error) {
	ctx = context.WithValue(ctx, connect.StripeAccountKey, p.checkout.PaymentAccountID)

	expiryYM := strings.Split(p.card.Expiry, "/")
	cardParam := &stripe.CardParams{
		Name:   p.card.Name,
		Number: p.card.Number,
		Month:  expiryYM[0],
		Year:   expiryYM[1],
		CVC:    p.card.CVC,
	}
	if b := p.billingAddress; b != nil {
		cardParam.Address1 = b.Address
		cardParam.Address2 = b.AddressOpt
		cardParam.City = b.City
		cardParam.State = b.Province
		cardParam.Zip = b.Zip
		cardParam.Country = b.Country
	}

	// network call to stripe's token exchange service
	lg.Info("exchanging stripe token")
	stripeToken, err := connect.ST.ExchangeToken(ctx, cardParam)
	if err != nil {
		return "", api.ErrStripeProcess(err)
	}

	// update request ip to one stripe passes back
	p.requestIP = stripeToken.ClientIP

	lg.Info("received stripe token")
	return stripeToken.ID, nil
}
