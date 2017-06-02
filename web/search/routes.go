package search

import (
	"net/http"
	"strings"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/data/presenter"
	"github.com/pressly/chi"
	"github.com/pressly/chi/render"
)

type omniSearch struct {
	Places   []*presenter.Place   `json:"places"`
	Products []*presenter.Product `json:"products"`
}

func Routes() chi.Router {
	r := chi.NewRouter()

	r.Post("/", OmniSearch)

	return r
}

// OmniSearch catch all search endpoint and returns categorized
// json search results
func OmniSearch(w http.ResponseWriter, r *http.Request) {
	q := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("q")))
	ctx := r.Context()
	user := ctx.Value("session.user").(*data.User)

	var distPlaces []*data.Place
	s := omniSearch{}

	places, err := data.DB.Place.MatchName(q)
	if err != nil {
		render.Render(w, r, api.WrapErr(err))
		return
	}
	s.Places = make([]*presenter.Place, len(places))
	for i, pl := range places {
		place := presenter.NewPlace(ctx, pl)
		s.Places[i] = place
		distPlaces = append(distPlaces, pl)
	}

	products, err := data.DB.Product.MatchTags(q)
	if err != nil {
		render.Render(w, r, api.WrapErr(err))
		return
	}
	s.Products = make([]*presenter.Product, len(products))
	for i, p := range products {
		pp := presenter.NewProduct(ctx, p)
		s.Products[i] = pp
		distPlaces = append(distPlaces, pp.Place.Place)
	}
	user.DistanceToPlaces(distPlaces...)

	render.Respond(w, r, s)
}
