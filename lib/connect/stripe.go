package connect

import (
	"context"
	"net/http"

	"golang.org/x/oauth2"

	"bitbucket.org/moodie-app/moodie-api/lib/stripe"
)

type Stripe struct {
	c    *http.Client
	auth string
}

const StripeAccountKey = "stripe.account"

var (
	ST *Stripe
)

func SetupStripe(conf Config) {
	// initiate authorization token source transport
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: conf.AppSecret},
	)
	tc := oauth2.NewClient(ctx, ts)

	ST = &Stripe{
		c:    tc,
		auth: conf.AppSecret,
	}
}

func (s *Stripe) ExchangeToken(ctx context.Context, card *stripe.CardParams) (*stripe.Token, error) {
	stripeAccountID, _ := ctx.Value(StripeAccountKey).(string)
	// TODO: context -> is there a better way?
	// i guess doesn't matter, this should be done on the frontend
	c := stripe.NewClient(stripeAccountID, s.c)
	tok, _, err := c.Token.CreateCard(ctx, card)
	return tok, err
}
