package presenter

import (
	"context"
	"net/http"

	"bitbucket.org/moodie-app/moodie-api/data"
)

type CartAddress struct {
	*data.CartAddress

	HasError bool   `json:"hasError"`
	Error    string `json:"error"`

	IsShipping bool `json:"isShipping"`
	IsBilling  bool `json:"isBilling"`
}

func NewCartAddress(ctx context.Context, s *data.CartAddress) *CartAddress {
	return &CartAddress{
		CartAddress: s,
	}
}

func (*CartAddress) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}
