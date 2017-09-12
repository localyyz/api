package stripe

import (
	"context"
	"net/http"
)

type TokenService service

// TokenType is the list of allowed values for a token's type.
// Allowed values are "card", "bank_account".
type TokenType string

// TokenParams is the set of parameters that can be used when creating a token.
// For more details see https://stripe.com/docs/api#create_card_token and https://stripe.com/docs/api#create_bank_account_token.
type TokenParams struct {
	Params
	Card *CardParams
	//Bank     *BankAccountParams
	//PII      *PIIParams
	Customer string
	// Email is an undocumented parameter used by Stripe Checkout
	// It may be removed from the API without notice.
	Email string
}

// Token is the resource representing a Stripe token.
// For more details see https://stripe.com/docs/api#tokens.
type Token struct {
	ID      string    `json:"id"`
	Live    bool      `json:"livemode"`
	Created int64     `json:"created"`
	Type    TokenType `json:"type"`
	Used    bool      `json:"used"`
	//Bank     *BankAccount `json:"bank_account"`
	Card     *Card  `json:"card"`
	ClientIP string `json:"client_ip"`
	// Email is an undocumented field but included for all tokens created
	// with Stripe Checkout.
	Email string `json:"email"`
}

func (t *TokenService) CreateCard(ctx context.Context, card *CardParams) (*Token, *http.Response, error) {
	body := &RequestValues{}
	card.AppendDetails(body, true)

	req, err := t.client.NewRequest("POST", "tokens", body)
	if err != nil {
		return nil, nil, err
	}

	token := new(Token)
	resp, err := t.client.Do(ctx, req, token)
	if err != nil {
		return nil, resp, err
	}

	return token, resp, nil
}
