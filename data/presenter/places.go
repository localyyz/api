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
	Locale     *Locale `json:"locale"`
	PromoCount uint64  `json:"promoCount"`
	Following  bool    `json:"following"`

	LatLng *geotools.LatLng `json:"coords"`
	ctx    context.Context
}

func NewPlace(ctx context.Context, place *data.Place) *Place {
	p := &Place{
		Place: place,
		ctx:   ctx,
	}

	{ // promotion count
		query := data.DB.Promo.Find(
			db.Cond{
				"place_id": p.ID,
				"status":   data.PromoStatusActive,
			},
		).Select(db.Raw("count(distinct product_id) as count"))

		var c struct {
			Count uint64 `db:"count"`
		}
		if err := query.One(&c); err != nil {
			return p
		}
		p.PromoCount = c.Count
	}

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
