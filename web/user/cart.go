package user

import (
	"context"
	"net/http"

	"github.com/pressly/chi/render"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/data/presenter"
	"bitbucket.org/moodie-app/moodie-api/web/api"
	db "upper.io/db.v3"
)

// TODO: shopping list concept
// claim -> product -> place is too complicated
//
// The architecture should be:
// - products can be added to shopping carts
//    at "checkout" pick the promotion if available
// - multiple shopping carts? would that become "collections"?
func GetCart(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	currentUser := ctx.Value("session.user").(*data.User)
	claims, err := data.DB.Claim.FindAll(
		db.Cond{
			"user_id": currentUser.ID,
			"status": []data.ClaimStatus{
				data.ClaimStatusActive,
				data.ClaimStatusCompleted,
			},
		},
	)
	if err != nil {
		render.Respond(w, r, err)
		return
	}

	if len(claims) == 0 {
		render.Render(w, r, api.EmptyListResp)
		return
	}

	// claim -> promotions
	promoIDs := make([]int64, len(claims))
	claimsMap := make(map[int64]*data.Claim, len(claims))
	for i, c := range claims {
		promoIDs[i] = c.PromoID
		claimsMap[c.PromoID] = c
	}

	promos, err := data.DB.Promo.FindAll(db.Cond{"id": promoIDs})
	if err != nil {
		render.Respond(w, r, err)
		return
	}

	// promo -> products
	var productIDs []int64
	promosMap := make(map[int64]*data.Promo)
	for _, p := range promos {
		if _, found := promosMap[p.ProductID]; !found {
			productIDs = append(productIDs, p.ProductID)
			promosMap[p.ProductID] = p
		}
	}

	products, err := data.DB.Product.FindAll(db.Cond{"id": productIDs})
	if err != nil {
		render.Respond(w, r, err)
		return
	}

	ctx = context.WithValue(ctx, "claims", claimsMap)
	ctx = context.WithValue(ctx, "promos", promosMap)
	presented := presenter.NewCartProductList(ctx, products)
	render.Render(w, r, presented)
}
