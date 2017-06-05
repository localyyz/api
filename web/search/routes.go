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
	Places   []*presenter.Place          `json:"places"`
	Products presenter.SearchProductList `json:"products"`
}

func (*omniSearch) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
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
	//user := ctx.Value("session.user").(*data.User)

	//var distPlaces []*data.Place
	s := &omniSearch{}

	places, err := data.DB.Place.MatchName(q)
	if err != nil {
		render.Respond(w, r, err)
		return
	}
	s.Places = make([]*presenter.Place, len(places))
	for i, pl := range places {
		place := presenter.NewPlace(ctx, pl)
		s.Places[i] = place
		//distPlaces = append(distPlaces, pl)
	}

	products, err := data.DB.Product.MatchTags(q)
	if err != nil {
		render.Respond(w, r, err)
		return
	}
	s.Products = presenter.NewSearchProductList(ctx, products)
	//distPlaces = append(distPlaces, pp.Place.Place)
	//user.DistanceToPlaces(distPlaces...)

	render.Render(w, r, s)
}
