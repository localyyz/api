package collection

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
)

func CollectionCtx(next http.Handler) http.Handler {
	handler := func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		collectionID, err := strconv.ParseInt(chi.URLParam(r, "collectionID"), 10, 64)
		if err != nil {
			render.Render(w, r, api.ErrBadID)
			return
		}

		collection, err := data.DB.Collection.FindByID(collectionID)
		if err != nil {
			render.Respond(w, r, err)
			return
		}

		ctx = context.WithValue(ctx, "collection", collection)
		lg.SetEntryField(ctx, "collection_id", collection.ID)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
	return http.HandlerFunc(handler)
}

func ListCollection(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var collections []*data.Collection
	err := data.DB.Collection.
		Find().
		OrderBy("ordering").
		All(&collections)
	if err != nil {
		render.Respond(w, r, err)
		return
	}
	presented := presenter.NewCollectionList(ctx, collections)
	if err := render.RenderList(w, r, presented); err != nil {
		render.Respond(w, r, err)
	}
}

func GetCollection(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	collection := ctx.Value("collection").(*data.Collection)
	render.Respond(w, r, collection)
}
