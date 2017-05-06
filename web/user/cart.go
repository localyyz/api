package user

import (
	"net/http"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/data/presenter"
	"bitbucket.org/moodie-app/moodie-api/lib/ws"
	db "upper.io/db.v3"
)

// TODO: shopping list concept
func GetCart(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	currentUser := ctx.Value("session.user").(*data.User)
	claims, err := data.DB.Claim.FindAll(
		db.Cond{
			"user_id": currentUser.ID,
			"status": []data.ClaimStatus{
				data.ClaimStatusActive,
				data.ClaimStatusSaved,
			},
		},
	)
	if err != nil {
		ws.Respond(w, http.StatusInternalServerError, err)
		return
	}

	if len(claims) == 0 {
		ws.Respond(w, http.StatusOK, []struct{}{})
		return
	}

	// claim -> promotions
	promoIDs := make([]int64, len(claims))
	for i, c := range claims {
		promoIDs[i] = c.PromoID
	}

	promos, err := data.DB.Promo.FindAll(db.Cond{"id": promoIDs})
	if err != nil {
		ws.Respond(w, http.StatusInternalServerError, err)
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
		ws.Respond(w, http.StatusInternalServerError, err)
		return
	}

	res := make([]*presenter.Product, len(products))
	for i, p := range products {
		res[i] = presenter.NewProduct(ctx, p).WithPlace().WithShopUrl()
		res[i].Promos = promoMap[p.ID]
	}

	ws.Respond(w, http.StatusOK, res)
}
