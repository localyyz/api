package connect

import (
	"context"

	"golang.org/x/oauth2"

	"bitbucket.org/moodie-app/moodie-api/lib/stripe"
)

type Stripe struct {
	c    *stripe.Client
	auth string
}

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
		c:    stripe.NewClient(tc),
		auth: conf.AppSecret,
	}
}

func (s *Stripe) ExchangeToken(ctx context.Context, account string, card *stripe.CardParams) (*stripe.Token, error) {
	// copy client to set the account
	c := *(s.c)

	c.StripeAccount = account
	tok, _, err := c.Token.CreateCard(ctx, card)

	return tok, err
}
