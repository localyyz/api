package presenter

import (
	"context"
	"net/http"

	"bitbucket.org/moodie-app/moodie-api/data"
	"github.com/go-chi/render"
)

type ProductVariant struct {
	*data.ProductVariant
	Place   *Place        `json:"place,omitempty"`
	Product *data.Product `json:"product,omitempty"`
	ImageID int64         `json:"imageId,omitempty"`
	Price   float64       `json:"price,omitempty"`

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
		Price:          variant.Price,
		ctx:            ctx,
	}

	// modify product price if deal is active
	if deal, ok := ctx.Value(DealCtxKey).(*data.Deal); ok {
		// NOTE: deal value here is negative because the type is fixed amount only for now
		p.Price += deal.Value
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
