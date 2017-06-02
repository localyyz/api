package user

import (
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
		render.Render(w, r, api.WrapErr(err))
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
		render.Render(w, r, api.WrapErr(err))
		return
	}

	// promo -> products
	var productIDs []int64
	promoMap := map[int64][]*presenter.Promo{}
	for _, p := range promos {
		productIDs = append(productIDs, p.ProductID)
		promoMap[p.ProductID] = append(promoMap[p.ProductID], presenter.NewPromo(ctx, p))
	}

	products, err := data.DB.Product.FindAll(db.Cond{"id": productIDs})
	if err != nil {
		render.Render(w, r, api.WrapErr(err))
		return
	}

	res := make([]*presenter.Product, len(products))
	for i, p := range products {
		res[i] = presenter.NewProduct(ctx, p)
		promo := promoMap[p.ID]
		res[i].Promos = promo
		res[i].UserClaimStatus = claimsMap[promo[0].ID].Status
	}

	render.Respond(w, r, res)
}
