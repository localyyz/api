package presenter

import (
	"context"
	"net/http"

	"bitbucket.org/moodie-app/moodie-api/data"
	"github.com/go-chi/render"
	"github.com/pkg/errors"
	"github.com/pressly/lg"
	"upper.io/db.v3"
)

type ProductVariant struct {
	*data.ProductVariant
	Place   *Place        `json:"place,omitempty"`
	Product *data.Product `json:"product,omitempty"`

	// Hide fields
	ProductID interface{} `json:"productId,omitempty"`
	PlaceID   interface{} `json:"placeId,omitempty"`
	UserID    interface{} `json:"userId,omitempty"`
	Status    interface{} `json:"status,omitempty"`
	CreatedAt interface{} `json:"createdAt,omitempty"`
	UpdatedAt interface{} `json:"updatedAt,omitempty"`
	DeletedAt interface{} `json:"deletedAt,omitempty"`

	ctx context.Context
}

func NewProductVariant(ctx context.Context, variant *data.ProductVariant) *ProductVariant {
	p := &ProductVariant{
		ProductVariant: variant,
		ctx:            ctx,
	}

	if p.Place == nil {

		var place *data.Place
		err := data.DB.Place.
			Find(p.PlaceID).
			Select(db.Raw("*")).
			OrderBy("distance").
			One(&place)
		if err != nil {
			if err != db.ErrNoMoreRows {
				lg.Error(errors.Wrapf(err, "failed to present variant(%v) place", p.ID))
			}
		}
		p.Place = NewPlace(ctx, place)
	}

	return p
}

func NewProductVariantList(ctx context.Context, variants []*data.ProductVariant) []render.Renderer {
	list := []render.Renderer{}
	for _, variant := range variants {
		list = append(list, NewProductVariant(ctx, variant))
	}
	return list
}

// ProductVariant implements render.Renderer interface
func (p *ProductVariant) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}
