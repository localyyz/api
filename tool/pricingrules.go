package tool

import (
	"net/http"

	"bitbucket.org/moodie-app/moodie-api/lib/shopify"
	"github.com/go-chi/render"
)

func ListPriceRule(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	client := ctx.Value("shopify.client").(*shopify.Client)

	priceRules, _, _ := client.PriceRule.List(ctx)
	render.Respond(w, r, priceRules)
}
