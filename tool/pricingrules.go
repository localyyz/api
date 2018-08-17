package tool

import (
	"net/http"
	"time"

	"bitbucket.org/moodie-app/moodie-api/lib/shopify"
	"github.com/go-chi/render"
	"github.com/pressly/lg"
)

func ListPriceRule(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	client := ctx.Value("shopify.client").(*shopify.Client)

	params := &shopify.PriceRuleParam{
		Page:  1,
		Limit: 200,
	}

	rules := []*shopify.PriceRule{}

	for {
		priceRules, _, _ := client.PriceRule.List(ctx, params)
		if len(priceRules) == 0 {
			break
		}
		for _, p := range priceRules {
			if p.CustomerSelection != "all" {
				continue
			}
			if p.EndsAt != nil && p.EndsAt.Before(time.Now()) {
				continue
			}
			if p.TargetType != "line_item" {
				continue
			}
			if p.TargetSelection != "all" {
				continue
			}
			rules = append(rules, p)
		}
		lg.Warnf("parsed page %d", params.Page)
		params.Page += 1

	}

	render.Respond(w, r, rules)
}
