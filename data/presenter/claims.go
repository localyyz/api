package presenter

import (
	"context"
	"net/http"

	"bitbucket.org/moodie-app/moodie-api/data"
)

type Claim struct {
	*data.Claim
	ctx context.Context
}

func NewClaim(ctx context.Context, claim *data.Claim) *Claim {
	return &Claim{Claim: claim, ctx: ctx}
}

func (c *Claim) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}
