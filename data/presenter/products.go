package presenter

import (
	"context"
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

func (p *Product) WithPromo() *Product {
	promos, err := data.DB.Promo.FindAll(
		db.Cond{
			"product_id": p.ID,
			"status":     data.PromoStatusActive,
		},
	)
	if err != nil {
		return p
	}

	for _, pr := range promos {
		p.Promos = append(p.Promos, NewPromo(p.ctx, pr))
		break // TODO: just return 1 for now
	}

	return p
}
