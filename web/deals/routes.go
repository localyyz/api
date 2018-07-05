package deals

import (
	"context"
	"net/http"
	"strconv"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/data/presenter"
	"github.com/go-chi/chi"
	"github.com/go-chi/render"
	"upper.io/db.v3"
)

func Routes() chi.Router {
	r := chi.NewRouter()

	r.Get("/active", ListActiveDeals)
	r.Get("/upcoming", ListQueuedDeals)
	r.With(DealCtx).Get("/active/{dealID}", GetDeal)

	return r
}

func DealCtx(next http.Handler) http.Handler {
	handler := func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		dealIDRaw := chi.URLParam(r, "dealID")
		dealID, _ := strconv.Atoi(dealIDRaw)
		ctx = context.WithValue(ctx, "dealID", dealID)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
	return http.HandlerFunc(handler)
}

/*
	retrieves all the active lightning collections ordered by the earliest it ends
	in the presenter -> calculates percentage complete
	in the presenter -> returns one product from the lightning collection
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
	in the presenter -> calculates percentage complete
	in the presenter -> returns one product from the lightning collection
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

func GetDeal(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	dealID := ctx.Value("dealID")

	var collection data.Collection
	err := data.DB.Collection.Find(db.Cond{"lightning": true, "status": data.CollectionStatusActive, "id": dealID}).OrderBy("end_at ASC").One(&collection)
	if err != nil {
		render.Respond(w, r, err)
		return
	}

	if err := render.Render(w, r, presenter.NewLightningCollection(r.Context(), &collection)); err != nil {
		render.Respond(w, r, err)
	}
}
