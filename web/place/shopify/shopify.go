package shopify

import (
	"context"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	set "gopkg.in/fatih/set.v0"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/lib/connect"
	"bitbucket.org/moodie-app/moodie-api/lib/shopify"
	"bitbucket.org/moodie-app/moodie-api/lib/ws"
	db "upper.io/db.v3"
)

func CredCtx(next http.Handler) http.Handler {
	handler := func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		place := ctx.Value("place").(*data.Place)

		creds, err := data.DB.ShopifyCred.FindByPlaceID(place.ID)
		if err != nil {
			ws.Respond(w, http.StatusInternalServerError, err)
			return
		}

		ctx = context.WithValue(ctx, "creds", creds)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
	return http.HandlerFunc(handler)
}

func ClientCtx(next http.Handler) http.Handler {
	handler := func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		creds := ctx.Value("creds").(*data.ShopifyCred)
		//authClient := connect.SH.ClientFromCred(r)
		api := shopify.NewClient(nil, creds.AccessToken)

		api.BaseURL, _ = url.Parse(creds.ApiURL)

		ctx = context.WithValue(ctx, "api", api)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
	return http.HandlerFunc(handler)
}

func Connect(w http.ResponseWriter, r *http.Request) {
	place := r.Context().Value("place").(*data.Place)
	count, err := data.DB.ShopifyCred.Find(db.Cond{"place_id": place.ID}).Count()
	if err != nil {
		ws.Respond(w, http.StatusInternalServerError, err)
		return
	}
	if count > 0 {
		ws.Respond(w, http.StatusConflict, "shopify store already connected")
		return
	}

	url := connect.SH.AuthCodeURL(r)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func tagSplit(r rune) bool {
	return r == ',' || r == ' '
}

func SyncProduct(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	place := ctx.Value("place").(*data.Place)

	api := ctx.Value("api").(*shopify.Client)
	productList, _, err := api.Product.List(ctx)
	if err != nil {
		ws.Respond(w, http.StatusBadRequest, err)
		return
	}

	// initial sync up
	for _, p := range productList {
		product := &data.Product{
			PlaceID:     place.ID,
			ExternalID:  p.Handle,
			Title:       p.Title,
			Description: p.BodyHTML,
			ImageUrl:    p.Image.Src,
		}

		tt := strings.FieldsFunc(p.Tags, tagSplit)
		tagSet := set.New()
		for _, t := range tt {
			tagSet.Add(strings.ToLower(strings.TrimSpace(t)))
		}
		tagSet.Add(strings.ToLower(p.ProductType))
		tagSet.Add(strings.ToLower(p.Vendor))

		product.Tags = set.StringSlice(tagSet)

		if err := data.DB.Product.Save(product); err != nil {
			ws.Respond(w, http.StatusInternalServerError, err)
			return
		}

		for _, v := range p.Variants {
			now := time.Now().UTC()
			start := now.Add(1 * time.Minute)
			end := now.Add(30 * 24 * time.Hour)
			price, _ := strconv.ParseFloat(v.Price, 64)
			promo := &data.Promo{
				PlaceID:     place.ID,
				ProductID:   product.ID,
				Type:        data.PromoTypePrice,
				OfferID:     v.ID,
				Status:      data.PromoStatusActive,
				Description: v.Title,
				UserID:      0, // admin
				Etc: data.PromoEtc{
					Price: price,
					Sku:   v.Sku,
				},
				StartAt: &start,
				EndAt:   &end, // 1 month
			}

			if err := data.DB.Promo.Save(promo); err != nil {
				ws.Respond(w, http.StatusInternalServerError, err)
				return
			}
		}
	}

	ws.Respond(w, http.StatusOK, productList)
}
