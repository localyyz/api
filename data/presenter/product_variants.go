package presenter

import (
	"context"
	"net/http"

	"bitbucket.org/moodie-app/moodie-api/data"
	xchange "bitbucket.org/moodie-app/moodie-api/lib/xchanger"
	"github.com/go-chi/render"
)

type ProductVariant struct {
	*data.ProductVariant
	Place   *data.Place   `json:"-"`
	Product *data.Product `json:"-"`
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
	pv := &ProductVariant{
		ProductVariant: variant,
		Price:          variant.Price,
		ctx:            ctx,
	}

	// modify product price if deal is active
	if deal, ok := ctx.Value("deal").(*data.Deal); ok {
		// NOTE: deal value here is negative because the type is fixed amount only for now
		if deal.Featured && deal.ProductListType == data.ProductListTypeAssociated {
			pv.Price += deal.Value
		}
	}

	return pv
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
	if p.Place != nil && p.Place.Currency != "USD" {
		p.Price = xchange.ToUSD(p.Price, p.Place.Currency)
		p.PrevPrice = xchange.ToUSD(p.PrevPrice, p.Place.Currency)
	}
	return nil
}
