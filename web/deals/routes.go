package deals

import (
	"bitbucket.org/moodie-app/moodie-api/data"
	"github.com/go-chi/chi"
	"github.com/go-chi/render"
	"net/http"
	"upper.io/db.v3"
	"bitbucket.org/moodie-app/moodie-api/data/presenter"
)

func Routes() chi.Router {
	r := chi.NewRouter()

	r.Get("/active", getActiveLightningCollections)
	r.Get("/upcoming", getUpcomingLightningCollections)

	return r
}

/*
	retrieves all the active lightning collections ordered by the earliest it ends
	in the presenter -> calculates percentage complete
	in the presenter -> returns one product from the lightning collection
*/
func getActiveLightningCollections(w http.ResponseWriter, r *http.Request) {
	var collections []*data.Collection

	res := data.DB.Collection.Find(db.Cond{"lightning": true, "status": data.CollectionStatusActive}).OrderBy("time_end ASC")
	err := res.All(&collections)
	if err != nil {
		render.Respond(w, r, err)
		return
	}

	collectionsWithFeaturedProducts, err := presenter.PresentActiveLightningCollection(collections)
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
func getUpcomingLightningCollections(w http.ResponseWriter, r *http.Request) {
	var collections []*data.Collection

	res := data.DB.Collection.Find(db.Cond{"lightning": true, "status": data.CollectionStatusQueued}).OrderBy("time_start ASC")
	err := res.All(&collections)
	if err != nil {
		render.Respond(w, r, err)
		return
	}

	collectionsWithFeaturedProducts, err := presenter.PresentUpcomingLightningCollection(collections)
	if err != nil {
		render.Respond(w, r, err)
	}

	if err := render.RenderList(w, r, collectionsWithFeaturedProducts); err != nil {
		render.Respond(w, r, err)
	}
}


