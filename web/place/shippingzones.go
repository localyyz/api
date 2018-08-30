package place

import (
	"net/http"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/data/presenter"
	"github.com/go-chi/render"
)

func ListShippingZone(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	place := ctx.Value("place").(*data.Place)
	zones, err := data.DB.ShippingZone.FindByPlaceID(place.ID)
	if err != nil {
		render.Respond(w, r, err)
		return
	}

	render.RenderList(w, r, presenter.NewShippingZoneList(ctx, zones))
}

func SearchShippingZone(w http.ResponseWriter, r *http.Request) {
}
