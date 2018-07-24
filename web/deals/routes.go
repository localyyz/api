package deals

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/data/presenter"
	"bitbucket.org/moodie-app/moodie-api/web/api"
	"github.com/go-chi/chi"
	"github.com/go-chi/render"
	"github.com/pressly/lg"
	"upper.io/db.v3"
)

func Routes() chi.Router {
	r := chi.NewRouter()

	r.Get("/upcoming", ListQueuedDeals)
	r.Get("/history", ListInactiveDeals)

	r.Post("/activate", ActivateDeal)
	r.Route("/active", func(r chi.Router) {
		r.Get("/", ListActiveDeals)
		r.Route("/{dealID}", func(r chi.Router) {
			r.Use(DealCtx)
			r.Get("/", GetDeal)
		})
	})

	return r
}

/*
	parses the dealID from the request url and fetches the deal to put in context
*/
func DealCtx(next http.Handler) http.Handler {
	handler := func(w http.ResponseWriter, r *http.Request) {
		dealID, err := strconv.ParseInt(chi.URLParam(r, "dealID"), 10, 64)
		if err != nil {
			render.Render(w, r, api.ErrBadID)
			return
		}

		deal, err := data.DB.Collection.FindOne(
			db.Cond{
				"id":     dealID,
				"status": data.CollectionStatusActive,
			},
		)
		if err != nil {
			render.Respond(w, r, err)
			return
		}
		ctx := r.Context()
		ctx = context.WithValue(ctx, "deal", deal)
		lg.SetEntryField(ctx, "deal_id", deal.ID)

		next.ServeHTTP(w, r.WithContext(ctx))
	}
	return http.HandlerFunc(handler)
}

type ActivateRequest struct {
	DealID int64 `json:"dealId,required"`
	//Token    string     `json:"token,required"`
	StartAt  *time.Time `json:"startAt,omitempty"`
	Duration int64      `json:"duration,omitempty"`
}

func (a *ActivateRequest) Bind(r *http.Request) error {
	if a.StartAt == nil {
		a.StartAt = data.GetTimeUTCPointer()
	}
	if a.Duration == 0 {
		a.Duration = 1
	}
	// TODO... validate token
	return nil
}

func ActivateDeal(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := ctx.Value("session.user").(*data.User)

	var payload ActivateRequest
	if err := render.Bind(r, &payload); err != nil {
		render.Respond(w, r, api.ErrInvalidRequest(err))
		return
	}

	// validate that user has not activated this deal before
	exists, _ := data.DB.UserDeal.Find(db.Cond{
		"user_id": user.ID,
		"status":  db.NotEq(data.CollectionStatusActive),
		"deal_id": payload.DealID,
	}).Exists()
	if exists {
		// return that the deal has already expired
		render.Respond(w, r, api.ErrExpiredDeal)
		return
	}

	// validate that the deal id must be a lightning deal
	// it can be any deal
	deal, err := data.DB.Collection.FindOne(db.Cond{
		"id":        payload.DealID,
		"lightning": true,
	})
	if err != nil {
		// return that the deal has already expired
		render.Respond(w, r, api.ErrInvalidRequest(err))
		return
	}

	// insert an active user deal
	userDeal := data.UserDeal{
		UserID:  user.ID,
		DealID:  deal.ID,
		Status:  data.CollectionStatusActive, // TODO... activate later?
		StartAt: *payload.StartAt,
		EndAt:   payload.StartAt.Add(time.Duration(payload.Duration) * time.Hour),
	}

	if err := data.DB.UserDeal.Create(userDeal); err != nil {
		lg.Warn(err, userDeal)
		// return that the deal has already expired
		render.Respond(w, r, api.ErrInvalidRequest(err))
		return
	}

	// activate the deal
	deal.Status = data.CollectionStatusActive
	deal.StartAt = &userDeal.StartAt
	deal.EndAt = &userDeal.EndAt

	render.Status(r, http.StatusCreated)
	render.Render(w, r, presenter.NewDeal(ctx, deal))
}

/*
	retrieves all the active lightning collections ordered by the earliest it ends
	in the presenter -> returns the products associated with it
*/
func ListActiveDeals(w http.ResponseWriter, r *http.Request) {
	var collections []*data.Collection

	// only fetch featured deal globally
	err := data.DB.Collection.Find(
		db.Cond{
			"lightning": true,
			"featured":  true,
			"status":    data.CollectionStatusActive,
		},
	).OrderBy("end_at ASC").All(&collections)
	if err != nil {
		render.Respond(w, r, err)
		return
	}

	ctx := r.Context()
	if user, ok := ctx.Value("session.user").(*data.User); ok {
		// combine with any user activated lightning deals
		userDeals, _ := data.DB.UserDeal.FindAll(db.Cond{
			"user_id": user.ID,
			"status":  data.CollectionStatusActive,
		})

		if len(userDeals) > 0 {
			userDealsSet := map[int64]*data.UserDeal{}
			dealIDs := make([]int64, len(userDeals))
			for i, d := range userDeals {
				dealIDs[i] = d.DealID
				userDealsSet[d.DealID] = d
			}

			// fetch the user deals. for now let's assume
			// there's no overlap between featured and user deals
			deals, _ := data.DB.Collection.FindAll(db.Cond{
				"id":        dealIDs,
				"lightning": true,
			})

			for _, d := range deals {
				dd := userDealsSet[d.ID]
				d.StartAt = &dd.StartAt
				d.EndAt = &dd.EndAt
			}

			// prepend the user deals
			collections = append(deals, collections...)
		}
	}

	presented := presenter.NewDealList(ctx, collections)
	if err := render.RenderList(w, r, presented); err != nil {
		render.Respond(w, r, err)
	}
}

/*
	retrieves all the upcoming lightning collections ordered by the earliest it starts
	in the presenter -> does not return any products
*/
func ListQueuedDeals(w http.ResponseWriter, r *http.Request) {
	var collections []*data.Collection

	res := data.DB.Collection.Find(
		db.Cond{
			"lightning": true,
			"status":    data.CollectionStatusQueued,
		},
	).OrderBy("start_at ASC")
	err := res.All(&collections)
	if err != nil {
		render.Respond(w, r, err)
		return
	}

	if err := render.RenderList(w, r, presenter.NewDealList(r.Context(), collections)); err != nil {
		render.Respond(w, r, err)
	}
}

/*
	retrieves all the inactive lightning collections ordered by the earliest it ended
	in the presenter -> returns the products associated with it
*/
func ListInactiveDeals(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	cursor := ctx.Value("cursor").(*api.Page)

	var collections []*data.Collection

	res := data.DB.Collection.Find(
		db.Cond{
			"lightning": true,
			"status":    data.CollectionStatusInactive,
		},
	).OrderBy("end_at DESC")

	paginate := cursor.UpdateQueryUpper(res)
	if err := paginate.All(&collections); err != nil {
		render.Respond(w, r, err)
		return
	}

	if err := render.RenderList(w, r, presenter.NewDealList(r.Context(), collections)); err != nil {
		render.Respond(w, r, err)
	}
}

func GetDeal(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	deal := ctx.Value("deal").(*data.Collection)
	presented := presenter.NewDeal(ctx, deal)
	if err := render.Render(w, r, presented); err != nil {
		render.Respond(w, r, err)
	}
}
