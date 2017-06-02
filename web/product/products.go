package product

import (
	"context"
	"net/http"
	"net/url"
	"strconv"

	db "upper.io/db.v3"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/data/presenter"
	"bitbucket.org/moodie-app/moodie-api/web/api"
	"github.com/goware/lg"
	"github.com/pressly/chi"
	"github.com/pressly/chi/render"
)

type claimRequest struct {
	ProductUrl string `json:"url"`
}

func (*claimRequest) Bind(r *http.Request) error {
	return nil
}

func ProductCtx(next http.Handler) http.Handler {
	handler := func(w http.ResponseWriter, r *http.Request) {
		productID, err := strconv.ParseInt(chi.URLParam(r, "productID"), 10, 64)
		if err != nil {
			render.Render(w, r, api.ErrBadID)
			return
		}

		product, err := data.DB.Product.FindByID(productID)
		if err != nil {
			render.Render(w, r, api.WrapErr(err))
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

	payload := &claimRequest{}
	if err := render.Bind(r, payload); err != nil {
		render.Render(w, r, api.ErrInvalidRequest(err))
		return
	}

	// find the promotion we're claiming
	u, err := url.Parse(payload.ProductUrl)
	if err != nil {
		render.Render(w, r, api.WrapErr(err))
		return
	}

	cond := db.Cond{
		"product_id": product.ID,
		"status":     data.PromoStatusActive,
	}
	if variant := u.Query().Get("variant"); len(variant) > 0 {
		cond["offer_id"], _ = strconv.ParseInt(variant, 10, 64)
	}
	promo, err := data.DB.Promo.FindOne(cond)
	if err != nil {
		if err == db.ErrNoMoreRows {
			lg.Warnf("no promo found with %+v", cond)
		}
		render.Render(w, r, api.WrapErr(err))
		return
	}

	newClaim := &data.Claim{
		PromoID: promo.ID,
		UserID:  user.ID,
		PlaceID: promo.PlaceID,
		Status:  data.ClaimStatusActive,
	}

	// check if claim already exists
	count, err := data.DB.Claim.Find(db.Cond{
		"promo_id": promo.ID,
		"user_id":  user.ID,
		"status":   data.ClaimStatusActive,
	}).Count()
	if err != nil {
		render.Render(w, r, api.WrapErr(err))
		return
	}

	presented := presenter.NewClaim(ctx, newClaim)
	if count > 0 {
		render.Render(w, r, presented)
		return
	}

	if err := data.DB.Claim.Save(newClaim); err != nil {
		render.Render(w, r, api.WrapErr(err))
		return
	}

	render.Status(r, http.StatusCreated)
	render.Render(w, r, presented)
}
