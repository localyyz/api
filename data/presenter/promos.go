package presenter

import (
	"context"
	"fmt"
	"net/http"

	"bitbucket.org/moodie-app/moodie-api/data"
	"github.com/goware/lg"
	"github.com/pkg/errors"
	"github.com/pressly/chi/render"
	"upper.io/db.v3"
)

type Promo struct {
	*data.Promo
	Place   *Place        `json:"place,omitempty"`
	Product *data.Product `json:"product,omitempty"`
	Claim   *data.Claim   `json:"claim,omitempty"`

	//fields that can be viewed
	NumClaimed int64 `json:"numClaimed,omitempty"`

	ctx context.Context
}

func NewPromo(ctx context.Context, promo *data.Promo) *Promo {
	p := &Promo{
		Promo: promo,
		ctx:   ctx,
	}

	if p.Place == nil {
		user := ctx.Value("session.user").(*data.User)

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
		p.Place = NewPlace(ctx, place)
	}

	return p
}

func NewPromoList(ctx context.Context, promos []*data.Promo) []render.Renderer {
	list := []render.Renderer{}
	for _, promo := range promos {
		list = append(list, NewPromo(ctx, promo))
	}
	return list
}

// Promo implements render.Renderer interface
func (p *Promo) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

func (p *Promo) WithProduct() *Promo {
	product, err := data.DB.Product.FindByID(p.ProductID)
	if err != nil {
		return p
	}
	p.Product = product

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
