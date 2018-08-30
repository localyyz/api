package approval

import (
	"context"
	"net/http"
	"strconv"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/lib/connect"
	"bitbucket.org/moodie-app/moodie-api/web/api"
	"github.com/flosch/pongo2"
	"github.com/go-chi/chi"
	"github.com/go-chi/render"
	"github.com/pressly/lg"
	db "upper.io/db.v3"
)

var (
	indexTmpl *pongo2.Template
	listTmpl  *pongo2.Template
)

func Init(env string) {
	filePath := ""
	if env == "development" {
		filePath = "."
	}
	indexTmpl = pongo2.Must(pongo2.FromFile(filePath + "/merchant/approval.html"))
	listTmpl = pongo2.Must(pongo2.FromFile(filePath + "/merchant/approvallist.html"))
}

func Routes() chi.Router {
	r := chi.NewRouter()

	r.Get("/", List)
	r.Route("/{placeID}", func(r chi.Router) {
		r.Use(PlaceCtx)

		r.Get("/", Index)
		r.Put("/", Update)
		r.Post("/approve", Approve)
		r.Post("/reject", Reject)
	})

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
		}).One(&place)
		if err != nil {
			render.Respond(w, r, err)
			return
		}

		ctx := r.Context()
		approval, err := data.DB.MerchantApproval.FindByPlaceID(place.ID)
		if err != nil {
			if err != db.ErrNoMoreRows {
				render.Respond(w, r, err)
				return
			}
			approval = &data.MerchantApproval{
				PlaceID: place.ID,
			}
		}
		ctx = context.WithValue(ctx, "approval", approval)

		if place.Status == data.PlaceStatusWaitApproval {
			// Change the status to "reviewing" if waiting for approval
			place.Status = data.PlaceStatusReviewing
			if err := data.DB.Place.Save(place); err != nil {
				lg.Warn(err)
			}
		}

		ctx = context.WithValue(ctx, "place", place)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
	return http.HandlerFunc(handler)
}

func List(w http.ResponseWriter, r *http.Request) {
	var waitApprovals []*data.Place
	data.DB.Place.Find(
		db.Cond{"status": data.PlaceStatusWaitApproval},
	).OrderBy("-id").All(&waitApprovals)

	var reviewing []*data.Place
	data.DB.Place.Find(
		db.Cond{"status": data.PlaceStatusReviewing},
	).OrderBy("-id").Limit(10).All(&reviewing)

	var recentApproved []*data.Place
	data.DB.Place.Find(
		db.Cond{"status": data.PlaceStatusActive},
	).OrderBy("-id").Limit(10).All(&recentApproved)

	pageContext := pongo2.Context{
		"wait":      waitApprovals,
		"reviewing": reviewing,
		"approved":  recentApproved,
	}
	t, err := listTmpl.Execute(pageContext)
	if err != nil {
		lg.Warn(err)
	}
	render.HTML(w, r, t)
}

func Approve(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	place := ctx.Value("place").(*data.Place)
	approval := ctx.Value("approval").(*data.MerchantApproval)

	approval.ApprovedAt = place.ApprovedAt
	data.DB.MerchantApproval.Save(approval)

	place.Status = data.PlaceStatusActive
	place.ApprovedAt = data.GetTimeUTCPointer()

	data.DB.Place.Save(place)

	go func() {
		// TODO: this should be some job some where initialized
		// by some signal
		connect.SH.RegisterWebhooks(context.Background(), place)
		connect.SH.RegisterReturnPolicy(context.Background(), place)
		connect.SH.RegisterShippingPolicy(context.Background(), place)
	}()

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

	Collection      int32 `json:"collection"`
	Category        int32 `json:"category"`
	PriceRange      int32 `json:"priceRange"`
	RejectionReason int32 `json:"rejectionReason"`

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
		approval.Category = data.MerchantApprovalCategory(payload.Category)
	}
	if payload.PriceRange != 0 {
		approval.PriceRange = data.MerchantApprovalPriceRange(payload.PriceRange)
	}
	if payload.Collection != 0 {
		approval.Collection = data.MerchantApprovalCollection(payload.Collection)
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
		approval.RejectionReason = data.MerchantApprovalRejection(payload.RejectionReason)
	}

	if err := data.DB.MerchantApproval.Save(approval); err != nil {
		render.Respond(w, r, api.ErrInvalidRequest(err))
		return
	}
	render.Respond(w, r, approval)
}
