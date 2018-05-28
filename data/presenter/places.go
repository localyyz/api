package presenter

import (
	"context"
	"net/http"

	"bitbucket.org/moodie-app/moodie-api/data"
	"github.com/go-chi/render"
	"upper.io/db.v3"
)

type Place struct {
	*data.Place
	ProductCount uint64            `json:"productCount"`
	Products     []render.Renderer `json:"products"`
	Following    bool              `json:"following"`

	IsFeatured bool `json:"isFeatured"`

	Billing     interface{} `json:"billing,omitempty"`
	TOSAgreedAt interface{} `json:"tosAgreedAt,omitempty"`
	ApprovedAt  interface{} `json:"approvedAt,omitempty"`

	ctx context.Context
}

func NewPlace(ctx context.Context, place *data.Place) *Place {
	p := &Place{
		Place: place,
		ctx:   ctx,
	}
	p.ProductCount, _ = data.DB.Product.Find(db.Cond{"place_id": p.ID}).Count()

	if withPreview, ok := ctx.Value("with.preview").(bool); withPreview && ok {
		cond := db.Cond{"place_id": p.ID}
		if gender, ok := ctx.Value("session.gender").(data.UserGender); ok {
			cond["gender"] = gender
		}
		var products []*data.Product
		data.DB.Product.Find(cond).Limit(4).All(&products)
		p.Products = NewProductList(ctx, products)
	}

	return p
}

func NewPlaceList(ctx context.Context, places []*data.Place) []render.Renderer {
	list := []render.Renderer{}
	for _, place := range places {
		p := NewPlace(ctx, place)
		if p.ProductCount == 0 {
			continue
		}
		list = append(list, p)
	}
	return list
}

// Place implements render.Renderer interface
func (pl *Place) Render(w http.ResponseWriter, r *http.Request) error {
	if pl.Weight >= data.PlaceFeatureWeightCutoff {
		pl.IsFeatured = true
	}
	return nil
}
