package presenter

import (
	"context"
	"net/http"

	"bitbucket.org/moodie-app/moodie-api/data"
	"github.com/go-chi/render"
	"github.com/goware/geotools"
	"upper.io/db.v3"
)

type Place struct {
	*data.Place
	Locale       *Locale `json:"locale"`
	ProductCount uint64  `json:"productCount"`
	Following    bool    `json:"following"`

	LatLng   *geotools.LatLng `json:"coords"`
	ImageURL string           `json:"imageUrl"`

	ctx context.Context
}

func NewPlace(ctx context.Context, place *data.Place) *Place {
	p := &Place{
		Place: place,
		ctx:   ctx,
	}
	p.ProductCount, _ = data.DB.Product.Find(db.Cond{"place_id": p.ID}).Count()

	locale, _ := data.DB.Locale.FindByID(p.LocaleID)
	p.Locale = NewLocale(ctx, locale)

	if user, ok := ctx.Value("session.user").(*data.User); ok {
		count, _ := data.DB.Following.Find(
			db.Cond{"place_id": p.ID, "user_id": user.ID},
		).Count()
		p.Following = (count > 0)
	}

	return p
}

func NewPlaceList(ctx context.Context, places []*data.Place) []render.Renderer {
	list := []render.Renderer{}
	for _, place := range places {
		p := NewPlace(ctx, place)
		if p.ProductCount == 0 {
			continue
		}
		list = append(list, p)
	}
	return list
}

// Place implements render.Renderer interface
func (pl *Place) Render(w http.ResponseWriter, r *http.Request) error {
	if len(pl.Geo.Coordinates) > 1 {
		pl.LatLng = geotools.LatLngFromPoint(pl.Geo)
	}
	pl.ImageURL = pl.Place.ImageURL
	if len(pl.ImageURL) == 0 {
		var p *data.Product
		err := data.DB.Product.Find(
			db.Cond{"place_id": pl.ID},
		).OrderBy("-created_at").One(&p)
		if err != nil {
			return err
		}
		pl.ImageURL = p.ImageUrl
	}
	return nil
}
