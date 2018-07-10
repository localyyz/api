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
	db "upper.io/db.v3"
)

func FeaturedScopeCtx(next http.Handler) http.Handler {
	handler := func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// get any existing scope
		scope, ok := ctx.Value("scope").(db.Cond)
		if !ok {
			scope = db.Cond{}
		}
		scope["collections.featured"] = true

		ctx = context.WithValue(ctx, "scope", scope)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
	return http.HandlerFunc(handler)
}

func FemaleScopeCtx(next http.Handler) http.Handler {
	handler := func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// get any existing scope
		scope, ok := ctx.Value("scope").(db.Cond)
		if !ok {
			scope = db.Cond{}
		}
		scope["collections.gender"] = data.ProductGenderFemale
		scope["collections.featured"] = false

		ctx = context.WithValue(ctx, "scope", scope)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
	return http.HandlerFunc(handler)
}

func MaleScopeCtx(next http.Handler) http.Handler {
	handler := func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// get any existing scope
		scope, ok := ctx.Value("scope").(db.Cond)
		if !ok {
			scope = db.Cond{}
		}
		scope["collections.gender"] = data.ProductGenderMale
		scope["collections.featured"] = false

		ctx = context.WithValue(ctx, "scope", scope)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
	return http.HandlerFunc(handler)
}

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

	genderScope := []int{
		int(data.ProductGenderFemale),
		int(data.ProductGenderMale),
		int(data.ProductGenderUnisex),
	}
	if gender, ok := ctx.Value("session.gender").(data.UserGender); ok {
		genderScope = []int{int(gender)}
	}
	var collections []*data.Collection
	err := data.DB.Collection.
		Find(db.Cond{
			"gender":    genderScope,
			"featured":  true,
			"lightning": false,
		}).
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
