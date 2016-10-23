package presenter

import (
	"context"

	"bitbucket.org/moodie-app/moodie-api/data"
	"github.com/goware/geotools"
	"github.com/goware/lg"
	"github.com/pkg/errors"
	"upper.io/db.v2"
)

type Place struct {
	*data.Place
	Locale *data.Locale `json:"locale"`
	Claim  *data.Claim  `json:"claim"`
	Promo  *Promo       `json:"promo"`

	LatLng *geotools.LatLng `json:"coords"`
	ctx    context.Context
}

func NewPlace(ctx context.Context, place *data.Place) *Place {
	return &Place{Place: place, Promo: &Promo{}, Claim: &data.Claim{}, ctx: ctx}
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

// present the place with promotion
func (pl *Place) WithPromo() *Place {
	user := pl.ctx.Value("session.user").(*data.User)

	promo, err := data.DB.Promo.FindByPlaceID(pl.ID)
	if err != nil {
		if err != db.ErrNoMoreRows {
			lg.Error(errors.Wrapf(err, "failed to present place(%v) promo", pl.ID))
		}
		return pl
	}
	pl.Promo = &Promo{}
	//if pl.Distance < data.PromoDistanceLimit {
	// TODO: for now, everything is viewable
	pl.Promo.Promo = promo
	//}

	nc, err := data.DB.Claim.Find(db.Cond{"promo_id": promo.ID}).Count()
	if err != nil {
		return pl
	}
	pl.Promo.NumClaimed = int64(nc)

	claim, err := data.DB.Claim.FindOne(db.Cond{"user_id": user.ID, "promo_id": promo.ID})
	if err != nil {
		if err != db.ErrNoMoreRows {
			lg.Error(errors.Wrapf(err, "failed to present user(%v) claim on promo(%v)", user.ID, promo.ID))
			return pl
		}
	}
	pl.Claim = claim

	return pl
}
