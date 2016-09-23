package presenter

import (
	"context"

	"bitbucket.org/moodie-app/moodie-api/data"
	"github.com/goware/lg"
	"github.com/pkg/errors"
	"upper.io/db.v2"
)

type Place struct {
	*data.Place
	Locale *data.Locale `json:"locale"`
	Promo  *data.Promo  `json:"promo"`
	Claim  *data.Claim  `json:"claim"`

	ctx context.Context
}

func NewPlace(ctx context.Context, place *data.Place) *Place {
	return &Place{Place: place, ctx: ctx}
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
	if pl.Distance < data.PromoDistanceLimit {
		pl.Promo = promo
	}

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
