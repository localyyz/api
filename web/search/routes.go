package search

import (
	"fmt"
	"net/http"
	"strings"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/data/presenter"
	"bitbucket.org/moodie-app/moodie-api/lib/ws"
	"github.com/pressly/chi"
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
		ws.Respond(w, http.StatusInternalServerError, err)
		return
	}
	s.Places = make([]*presenter.Place, len(places))
	for i, pl := range places {
		place := presenter.NewPlace(ctx, pl).WithLocale()
		s.Places[i] = place
		distPlaces = append(distPlaces, pl)
	}

	products, err := data.DB.Product.MatchTags(q)
	if err != nil {
		ws.Respond(w, http.StatusInternalServerError, err)
		return
	}
	s.Products = make([]*presenter.Product, len(products))
	for i, p := range products {
		pp := presenter.NewProduct(ctx, p).WithPromo().WithPlace()
		pp.ShopUrl = fmt.Sprintf("%s/products/%s", pp.Place.Website, p.ExternalID)
		s.Products[i] = pp
		distPlaces = append(distPlaces, pp.Place.Place)
	}

	user.DistanceToPlaces(distPlaces...)

	ws.Respond(w, http.StatusOK, s)
}
