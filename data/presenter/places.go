package presenter

import (
	"context"
	"net/http"

	"bitbucket.org/moodie-app/moodie-api/data"
	"github.com/goware/geotools"
	"github.com/pressly/chi/render"
	"upper.io/db.v3"
)

type Place struct {
	*data.Place
	Locale       *Locale `json:"locale"`
	ProductCount uint64  `json:"productCount"`
	Following    bool    `json:"following"`

	LatLng *geotools.LatLng `json:"coords"`
	ctx    context.Context
}

func NewPlace(ctx context.Context, place *data.Place) *Place {
	p := &Place{
		Place: place,
		ctx:   ctx,
	}
	p.ProductCount, _ = data.DB.Product.Find(db.Cond{"place_id": p.ID}).Count()

	locale, _ := data.DB.Locale.FindByID(p.LocaleID)
	p.Locale = NewLocale(ctx, locale)
	p.Following = data.DB.Following.IsFollowing(ctx, p.ID)

	return p
}

func NewPlaceList(ctx context.Context, places []*data.Place) []render.Renderer {
	list := []render.Renderer{}
	for _, place := range places {
		list = append(list, NewPlace(ctx, place))
	}
	return list
}

// Place implements render.Renderer interface
func (pl *Place) Render(w http.ResponseWriter, r *http.Request) error {
	if len(pl.Geo.Coordinates) > 1 {
		pl.LatLng = geotools.LatLngFromPoint(pl.Geo)
	}
	return nil
}
