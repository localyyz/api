package presenter

import (
	"context"
	"fmt"

	"bitbucket.org/moodie-app/moodie-api/data"
	"github.com/goware/lg"
	"github.com/pkg/errors"
	"upper.io/db.v3"
)

type Promo struct {
	*data.Promo
	Place *Place      `json:"place,omitempty"`
	Claim *data.Claim `json:"claim,omitempty"`

	//fields that can be viewed
	NumClaimed int64 `json:"numClaimed,omitempty"`

	ctx context.Context
}

func NewPromo(ctx context.Context, promo *data.Promo) *Promo {
	return &Promo{
		Promo: promo,
		ctx:   ctx,
	}
}

func (p *Promo) WithPlace() *Promo {
	user := p.ctx.Value("session.user").(*data.User)

	var place *data.Place
	err := data.DB.Place.
		Find(p.PlaceID).
		Select(
			db.Raw("*"),
			db.Raw(fmt.Sprintf("ST_Distance(geo, st_geographyfromtext('%v'::text)) distance", user.Geo)),
		).
		OrderBy("distance").
		One(&place)
	if err != nil {
		if err != db.ErrNoMoreRows {
			lg.Error(errors.Wrapf(err, "failed to present promo(%v) place", p.ID))
		}
	}

	p.Place = (&Place{Place: place}).WithGeo()
	return p
}

func (p *Promo) WithClaim() *Promo {
	user := p.ctx.Value("session.user").(*data.User)

	var err error
	p.Claim, err = data.DB.Claim.FindOne(
		db.Cond{
			"place_id": p.PlaceID,
			"promo_id": p.ID,
			"user_id":  user.ID,
		},
	)
	if err != nil {
		if err != db.ErrNoMoreRows {
			lg.Error(errors.Wrapf(err, "failed to present promo(%v) place", p.ID))
		}
	}
	return p
}
