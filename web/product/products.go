package product

import (
	"context"
	"net/http"
	"net/url"
	"strconv"

	db "upper.io/db.v3"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/lib/ws"
	"bitbucket.org/moodie-app/moodie-api/web/api"
	"github.com/pressly/chi"
)

func ProductCtx(next http.Handler) http.Handler {
	handler := func(w http.ResponseWriter, r *http.Request) {
		productID, err := strconv.ParseInt(chi.URLParam(r, "productID"), 10, 64)
		if err != nil {
			ws.Respond(w, http.StatusBadRequest, api.ErrBadID)
			return
		}

		product, err := data.DB.Product.FindByID(productID)
		if err != nil {
			ws.Respond(w, http.StatusInternalServerError, err)
			return
		}
		ctx := r.Context()
		ctx = context.WithValue(ctx, "product", product)
		next.ServeHTTP(w, r.WithContext(ctx))
	}

	return http.HandlerFunc(handler)
}

func ClaimProduct(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	product := ctx.Value("product").(*data.Product)
	user := ctx.Value("session.user").(*data.User)

	var payload struct {
		ProductUrl string `json:"url"`
	}
	if err := ws.Bind(r.Body, &payload); err != nil {
		ws.Respond(w, http.StatusBadRequest, err)
		return
	}

	// find the promotion we're claiming
	u, err := url.Parse(payload.ProductUrl)
	if err != nil {
		ws.Respond(w, http.StatusBadRequest, err)
		return
	}

	cond := db.Cond{"product_id": product.ID}
	if variant := u.Query().Get("variant"); len(variant) > 0 {
		cond["offer_id"], _ = strconv.ParseInt(variant, 10, 64)
	}
	promo, err := data.DB.Promo.FindOne(cond)
	if err != nil {
		ws.Respond(w, http.StatusInternalServerError, err)
		return
	}

	newClaim := &data.Claim{
		PromoID: promo.ID,
		UserID:  user.ID,
		PlaceID: promo.PlaceID,
		Status:  data.ClaimStatusActive,
	}
	if err := data.DB.Claim.Save(newClaim); err != nil {
		ws.Respond(w, http.StatusInternalServerError, err)
		return
	}

	ws.Respond(w, http.StatusCreated, newClaim)
}
