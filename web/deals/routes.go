package deals

import (
	"bitbucket.org/moodie-app/moodie-api/data"
	"github.com/go-chi/chi"
	"github.com/go-chi/render"
	"net/http"
	"upper.io/db.v3"
	"bitbucket.org/moodie-app/moodie-api/data/presenter"
	"strconv"
	"context"
)

func Routes() chi.Router {
	r := chi.NewRouter()

	r.Get("/active", ListActiveDeals)
	r.With(DealCtx).Get("/active/{dealID}", ListSpecificActiveDeal)
	r.Get("/upcoming", ListQueuedDeals)

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

	res := data.DB.Collection.Find(db.Cond{"lightning": true, "status": data.CollectionStatusActive}).OrderBy("time_end ASC")
	err := res.All(&collections)
	if err != nil {
		render.Respond(w, r, err)
		return
	}

	ctx := r.Context()
	collectionsWithFeaturedProducts, err := presenter.PresentLightningCollection(ctx, collections, true)
	if err != nil {
		render.Respond(w, r, err)
	}

	if err := render.RenderList(w, r, collectionsWithFeaturedProducts); err != nil {
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

	res := data.DB.Collection.Find(db.Cond{"lightning": true, "status": data.CollectionStatusQueued}).OrderBy("time_start ASC")
	err := res.All(&collections)
	if err != nil {
		render.Respond(w, r, err)
		return
	}

	ctx := r.Context()
	collectionsWithFeaturedProducts, err := presenter.PresentLightningCollection(ctx, collections, false)
	if err != nil {
		render.Respond(w, r, err)
	}

	if err := render.RenderList(w, r, collectionsWithFeaturedProducts); err != nil {
		render.Respond(w, r, err)
	}
}

func ListSpecificActiveDeal(w http.ResponseWriter, r *http.Request){
	ctx := r.Context()
	dealID := ctx.Value("dealID")

	var collections []*data.Collection
	res := data.DB.Collection.Find(db.Cond{"lightning": true, "status": data.CollectionStatusActive, "id": dealID}).OrderBy("time_end ASC")
	err := res.All(&collections)
	if err != nil {
		render.Respond(w, r, err)
		return
	}

	collectionsWithFeaturedProducts, err := presenter.PresentLightningCollection(ctx, collections, true)
	if err != nil {
		render.Respond(w, r, err)
	}

	if err := render.RenderList(w, r, collectionsWithFeaturedProducts); err != nil {
		render.Respond(w, r, err)
	}

}
