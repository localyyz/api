package search

import (
	"net/http"
	"strings"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/data/presenter"
	"github.com/pressly/chi/render"
)

type omniSearch struct {
	Places   []*presenter.Place          `json:"places"`
	Products presenter.SearchProductList `json:"products"`
}

func (*omniSearch) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

// OmniSearch catch all search endpoint and returns categorized
// json search results
func OmniSearch(w http.ResponseWriter, r *http.Request) {
	q := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("q")))
	ctx := r.Context()

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
	}

	products, err := data.DB.Product.Fuzzy(q)
	if err != nil {
		render.Respond(w, r, err)
		return
	}
	s.Products = presenter.NewSearchProductList(ctx, products)

	render.Render(w, r, s)
}
