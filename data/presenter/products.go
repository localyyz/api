package presenter

import (
	"context"
	"fmt"
	"time"

	"upper.io/db.v3"

	"bitbucket.org/moodie-app/moodie-api/data"
)

type Product struct {
	*data.Product
	Promos  []*Promo `json:"promos"`
	ShopUrl string   `json:"shopUrl"`

	CreateAt  *time.Time `json:"createdAt,omitempty"`
	UpdatedAt *time.Time `json:"updatedAt,omitempty"`
	DeleteAt  *time.Time `json:"deletedAt,omitempty"`

	ctx context.Context
}

func NewProduct(ctx context.Context, product *data.Product) *Product {
	return &Product{
		Product: product,
		Promos:  make([]*Promo, 0),
		ctx:     ctx,
	}
}

func (p *Product) WithShopUrl() *Product {
	track, ok := p.ctx.Value("track").(*data.TrackList)
	if !ok {
		return p
	}
	p.ShopUrl = fmt.Sprintf("%s/products/%s", track.SalesUrl, p.ExternalID)
	return p
}

func (p *Product) WithPromo() *Product {
	var promo *data.Promo
	err := data.DB.Promo.Find(
		db.Cond{
			"product_id": p.ID,
			"status":     data.PromoStatusActive,
		},
	).OrderBy("id").One(&promo)
	if err != nil {
		return p
	}

	p.Promos = []*Promo{NewPromo(p.ctx, promo)}
	return p
}
