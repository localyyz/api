package approval

import (
	"context"
	"net/http"
	"net/url"
	"strconv"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/lib/connect"
	"bitbucket.org/moodie-app/moodie-api/lib/shopify"
	"bitbucket.org/moodie-app/moodie-api/lib/sync"
	"bitbucket.org/moodie-app/moodie-api/web/api"
	"github.com/flosch/pongo2"
	"github.com/go-chi/chi"
	"github.com/go-chi/render"
	"github.com/pressly/lg"
	db "upper.io/db.v3"
)

var (
	indexTmpl = pongo2.Must(pongo2.FromFile("./merchant/approval/index.html"))
)

func Routes() chi.Router {
	r := chi.NewRouter()

	r.Use(PlaceCtx)

	r.Get("/", Index)

	r.Put("/", Update)
	r.Post("/approve", Approve)
	r.Post("/reject", Reject)

	return r
}

func PlaceCtx(next http.Handler) http.Handler {
	handler := func(w http.ResponseWriter, r *http.Request) {
		placeID, err := strconv.ParseInt(chi.URLParam(r, "placeID"), 10, 64)
		if err != nil {
			render.Render(w, r, api.ErrBadID)
			return
		}

		var place *data.Place
		err = data.DB.Place.Find(db.Cond{
			"id": placeID,
			//"status": data.PlaceStatusWaitApproval,
		}).One(&place)
		if err != nil {
			render.Respond(w, r, err)
			return
		}

		approval, err := data.DB.MerchantApproval.FindByPlaceID(place.ID)
		if err != nil {
			if err != db.ErrNoMoreRows {
				render.Respond(w, r, err)
				return
			}
			approval = &data.MerchantApproval{
				PlaceID: place.ID,
			}
			// Change the status to "reviewing" if it's the first time
			place.Status = data.PlaceStatusReviewing
			data.DB.Place.Save(place)
		}

		ctx := r.Context()
		ctx = context.WithValue(ctx, "place", place)
		ctx = context.WithValue(ctx, "approval", approval)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
	return http.HandlerFunc(handler)
}

func Approve(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	place := ctx.Value("place").(*data.Place)
	approval := ctx.Value("approval").(*data.MerchantApproval)

	place.Status = data.PlaceStatusActive
	place.ApprovedAt = data.GetTimeUTCPointer()
	approval.ApprovedAt = place.ApprovedAt

	data.DB.Place.Save(place)
	data.DB.Place.Save(approval)

	go func() {
		creds, err := data.DB.ShopifyCred.FindOne(
			db.Cond{
				"place_id": place.ID,
				"status":   data.ShopifyCredStatusActive,
			},
		)
		if err != nil {
			lg.Alertf("merchant approval: credential invalid %+v", err)
			return
		}
		// Run this in a separate go func so it doesn't block
		// returning to slack

		// check if the merchant has already published item to us. and
		// if they have. pull and create the products on our side
		ctx := context.WithValue(r.Context(), "sync.place", place)
		cl := shopify.NewClient(nil, creds.AccessToken)
		cl.BaseURL, _ = url.Parse(creds.ApiURL)

		if count, _, _ := cl.ProductList.Count(ctx); count > 0 {
			page := 1
			for {
				productList, _, _ := cl.ProductList.Get(
					ctx,
					&shopify.ProductListParam{Limit: 50, Page: page},
				)
				if len(productList) == 0 {
					break
				}
				ctx = context.WithValue(ctx, "sync.list", productList)
				sync.ShopifyProductListingsCreate(ctx)
				page += 1
			}
		}
	}()

	// register the webhooks
	connect.SH.RegisterWebhooks(r.Context(), place)

	render.Respond(w, r, approval)
}

func Reject(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	place := ctx.Value("place").(*data.Place)
	approval := ctx.Value("approval").(*data.MerchantApproval)

	place.Status = data.PlaceStatusInActive
	approval.RejectedAt = data.GetTimeUTCPointer()

	data.DB.Place.Save(place)
	data.DB.MerchantApproval.Save(approval)

	render.Respond(w, r, approval)
}

func Index(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	place := ctx.Value("place").(*data.Place)
	approval := ctx.Value("approval").(*data.MerchantApproval)

	pageContext := pongo2.Context{
		"place":    place,
		"approval": approval,
		"status":   place.Status.String(),
		"plan":     place.Plan,
	}
	t, _ := indexTmpl.Execute(pageContext)
	render.HTML(w, r, t)
}

type merchantApprovalRequest struct {
	*data.MerchantApproval

	Gender data.PlaceGender `json:"gender"`

	ID         interface{} `json:"id"`
	PlaceID    interface{} `json:"placeID"`
	CreatedAt  interface{} `json:"createdAt"`
	UpdatedAt  interface{} `json:"updatedAt"`
	ApprovedAt interface{} `json:"approvedAt"`
	RejectedAt interface{} `json:"rejectedAt"`
}

func (m *merchantApprovalRequest) Bind(r *http.Request) error {
	return nil
}

func Update(w http.ResponseWriter, r *http.Request) {
	var payload merchantApprovalRequest
	if err := render.Bind(r, &payload); err != nil {
		render.Respond(w, r, api.ErrInvalidRequest(err))
		return
	}

	ctx := r.Context()
	approval := ctx.Value("approval").(*data.MerchantApproval)

	if payload.Category != 0 {
		approval.Category = payload.Category
	}
	if payload.PriceRange != 0 {
		approval.PriceRange = payload.PriceRange
	}
	if payload.Collection != 0 {
		approval.Collection = payload.Collection
	}
	if payload.Gender != 0 {
		place := ctx.Value("place").(*data.Place)
		place.Gender = payload.Gender
		if err := data.DB.Place.Save(place); err != nil {
			render.Respond(w, r, api.ErrInvalidRequest(err))
			return
		}
	}
	if payload.RejectionReason != 0 {
		approval.RejectionReason = payload.RejectionReason
	}

	if err := data.DB.MerchantApproval.Save(approval); err != nil {
		render.Respond(w, r, api.ErrInvalidRequest(err))
		return
	}
	render.Respond(w, r, approval)
}
