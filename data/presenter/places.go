package presenter

import (
	"context"
	"time"

	"bitbucket.org/moodie-app/moodie-api/data"
	"github.com/goware/geotools"
	"github.com/goware/lg"
	"github.com/pkg/errors"
	"upper.io/db.v3"
)

type Place struct {
	*data.Place
	Locale     *data.Locale `json:"locale"`
	PromoCount uint64       `json:"promoCount"`
	Following  bool         `json:"following"`

	LatLng *geotools.LatLng `json:"coords"`
	ctx    context.Context
}

func NewPlace(ctx context.Context, place *data.Place) *Place {
	p := &Place{
		Place: place,
		ctx:   ctx,
	}
	return p
}

func (pl *Place) WithFollowing() *Place {
	user := pl.ctx.Value("session.user").(*data.User)
	count, _ := data.DB.Following.Find(
		db.Cond{"place_id": pl.ID, "user_id": user.ID},
	).Count()
	pl.Following = count > 0

	return pl
}

func (pl *Place) WithGeo() *Place {
	pl.LatLng = geotools.LatLngFromPoint(pl.Place.Geo)
	return pl
}

// presents given place with locale detail
func (pl *Place) WithLocale() *Place {
	var err error
	if pl.Locale, err = data.DB.Locale.FindByID(pl.LocaleID); err != nil && err != db.ErrNoMoreRows {
		lg.Error(errors.Wrapf(err, "failed to present place(%v) locale", pl.ID))
	}
	return pl
}

// Count number of active promotions for a place
func (pl *Place) WithPromo() *Place {
	query := data.DB.Promo.Find(
		db.Cond{
			"place_id":    pl.ID,
			"start_at <=": time.Now().UTC(),
			"end_at >":    time.Now().UTC(),
			"status":      data.PromoStatusActive,
		},
	).Select(db.Raw("count(distinct product_id) as count"))

	var c struct {
		Count uint64 `db:"count"`
	}
	if err := query.One(&c); err != nil {
		return pl
	}
	pl.PromoCount = c.Count

	return pl
}
