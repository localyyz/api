package presenter

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/pressly/chi/render"

	"upper.io/db.v3"

	"bitbucket.org/moodie-app/moodie-api/data"
)

type Product struct {
	*data.Product
	Promos []*Promo `json:"promos"`
	Place  *Place   `json:"place"`

	ShopUrl string `json:"shopUrl"`
	//session user's claim status
	UserClaimStatus data.ClaimStatus `json:"userClaimStatus,omitempty"`

	CreateAt  *time.Time `json:"createdAt,omitempty"`
	UpdatedAt *time.Time `json:"updatedAt,omitempty"`
	DeleteAt  *time.Time `json:"deletedAt,omitempty"`

	ctx context.Context
}

type SearchProductList []*Product

func (l SearchProductList) Render(w http.ResponseWriter, r *http.Request) error {
	for _, v := range l {
		if err := v.Render(w, r); err != nil {
			return err
		}
	}
	return nil
}

func NewSearchProductList(ctx context.Context, products []*data.Product) SearchProductList {
	list := SearchProductList{}
	for _, product := range products {
		list = append(list, NewProduct(ctx, product))
	}
	return list
}

func NewProductList(ctx context.Context, products []*data.Product) []render.Renderer {
	list := []render.Renderer{}
	for _, product := range products {
		list = append(list, NewProduct(ctx, product))
	}
	return list
}

func NewProduct(ctx context.Context, product *data.Product) *Product {
	p := &Product{
		Product: product,
		ctx:     ctx,
	}

	place, ok := p.ctx.Value("place").(*data.Place)
	if !ok {
		place, _ = data.DB.Place.FindByID(p.PlaceID)
	}
	p.Place = &Place{Place: place}

	var promo *data.Promo
	data.DB.Promo.Find(
		db.Cond{
			"product_id": p.ID,
			"status":     data.PromoStatusActive,
		},
	).OrderBy("id").One(&promo)
	p.Promos = []*Promo{{Promo: promo}}

	return p
}

func (p *Product) Render(w http.ResponseWriter, r *http.Request) error {
	var u *url.URL
	if p.Place.ShopifyID != "" {
		u = &url.URL{
			Host: fmt.Sprintf("%s.myshopify.com", p.Place.ShopifyID),
		}
	} else if p.Place.Website != "" {
		u, _ = url.Parse(p.Place.Website)
	}

	u.Scheme = "https"
	u.Path = fmt.Sprintf("products/%s", p.ExternalID)

	p.ShopUrl = u.String()

	return nil
}
