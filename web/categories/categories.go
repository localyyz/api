package categories

/****

import (
	"context"
	"net/http"
	"strconv"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/data/presenter"
	"bitbucket.org/moodie-app/moodie-api/lib/ws"
	"bitbucket.org/moodie-app/moodie-api/web/api"
	"github.com/pressly/chi"
)

func CategoryCtx(next http.Handler) http.Handler {
	handler := func(w http.ResponseWriter, r *http.Request) {
		rawCategory, err := strconv.Atoi(chi.URLParam(r, "category"))
		if err != nil {
			ws.Respond(w, http.StatusBadRequest, api.ErrBadID)
			return
		}
		category := data.Category(rawCategory)
		ctx := r.Context()
		ctx = context.WithValue(ctx, "category", category)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
	return http.HandlerFunc(handler)
}

// Return list of categories
func ListCategories(w http.ResponseWriter, r *http.Request) {
	// TODO: fake it til we make it
	type category struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}
	categories := make([]category, len(data.Categories))
	for i, cat := range data.Categories {
		categories[i] = category{ID: i, Name: cat}
	}
	ws.Respond(w, http.StatusOK, categories)
}

func GetCategory(w http.ResponseWriter, r *http.Request) {
	// TODO: fake it til you make it
	ctx := r.Context()
	category := ctx.Value("category").(data.Category)

	result := struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}{
		ID:   int(category),
		Name: category.String(),
	}
	ws.Respond(w, http.StatusOK, result)
}

// ListPlaces returns places and promos based on category
func ListPlaces(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	category := ctx.Value("category").(data.Category)

	places, err := data.DB.Place.FindByCategory(category)
	if err != nil {
		ws.Respond(w, http.StatusInternalServerError, err)
		return
	}

	presented := make([]*presenter.Place, len(places))
	for i, pl := range places {
		presented[i] = presenter.NewPlace(ctx, pl).WithLocale().WithGeo().WithPromo()
	}

	ws.Respond(w, http.StatusOK, presented)
}
***/
