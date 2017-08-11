package cardvault

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
)

type CreditCard struct {
	Number            string `json:"number"`
	FirstName         string `json:"first_name"`
	LastName          string `json:"last_name"`
	Month             string `json:"month"`
	Year              string `json:"year"`
	VerificationValue string `json:"verification_value"`
}

const cardVaultURL = "https://elb.deposit.shopifycs.com/sessions"

func AddCard(ctx context.Context, card *CreditCard) (string, *http.Response, error) {
	var buf io.ReadWriter
	if card != nil {
		buf = new(bytes.Buffer)
		err := json.NewEncoder(buf).Encode(card)
		if err != nil {
			return "", nil, err
		}
	}

	req, err := http.NewRequest("POST", cardVaultURL, buf)
	if err != nil {
		return "", nil, err
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", nil, err
	}

	var cardVaultResponse struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(res.Body).Decode(&cardVaultResponse); err != nil {
		return "", res, err
	}

	return cardVaultResponse.ID, res, nil
}
