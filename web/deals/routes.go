package deals

import (
	"context"
	"net/http"
	"strconv"

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

	r.Get("/active", ListActiveDeals)
	r.Get("/upcoming", ListQueuedDeals)
	r.Get("/history", ListInactiveDeals)
	r.With(DealCtx).Get("/active/{dealID}", GetDeal)

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

/*
	retrieves all the active lightning collections ordered by the earliest it ends
	in the presenter -> returns the products associated with it
*/
func ListActiveDeals(w http.ResponseWriter, r *http.Request) {
	var collections []*data.Collection

	res := data.DB.Collection.Find(
		db.Cond{
			"lightning": true,
			"status":    data.CollectionStatusActive,
		},
	).OrderBy("end_at ASC")
	err := res.All(&collections)
	if err != nil {
		render.Respond(w, r, err)
		return
	}

	if err := render.RenderList(w, r, presenter.NewLightningCollectionList(r.Context(), collections)); err != nil {
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

	if err := render.RenderList(w, r, presenter.NewLightningCollectionList(r.Context(), collections)); err != nil {
		render.Respond(w, r, err)
	}
}

/*
	retrieves all the inactive lightning collections ordered by the earliest it ended
	in the presenter -> returns the products associated with it
*/
func ListInactiveDeals(w http.ResponseWriter, r *http.Request){
	var collections []*data.Collection

	res := data.DB.Collection.Find(
		db.Cond{
			"lightning": true,
			"status": data.CollectionStatusInactive,
		},
	).OrderBy("end_at DESC")

	err := res.All(&collections)
	if err != nil {
		render.Respond(w, r, err)
		return
	}

	if err := render.RenderList(w, r, presenter.NewLightningCollectionList(r.Context(), collections)); err != nil {
		render.Respond(w, r, err)
	}
}

func GetDeal(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	deal := ctx.Value("deal").(*data.Collection)
	presented := presenter.NewLightningCollection(ctx, deal)
	if err := render.Render(w, r, presented); err != nil {
		render.Respond(w, r, err)
	}
}
