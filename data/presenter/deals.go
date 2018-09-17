package presenter

import (
	"context"
	"fmt"
	"net/http"

	"bitbucket.org/moodie-app/moodie-api/data"
	xchange "bitbucket.org/moodie-app/moodie-api/lib/xchanger"
	"github.com/go-chi/render"
	db "upper.io/db.v3"
)

type Deal struct {
	*data.Deal
	Products []*Product `json:"products"`

	Code       interface{} `json:"code,omitempty"`
	ExternalID interface{} `json:"externalId,omitempty"`
	MerchantID interface{} `json:"merchantId,omitempty"`
	ParentID   interface{} `json:"parentId,omitempty"`
	UserID     interface{} `json:"userId,omitempty"`
	// Legacy + not implemented
	Cap int32 `json:"cap"`

	// pulled from products
	Title       string `json:"title"`
	Description string `json:"description"`
	ImageURL    string `json:"imageUrl,omitempty"`
	ImageWidth  int64  `json:"imageWidth,omitempty"`
	ImageHeight int64  `json:"imageHeight,omitempty"`
}

const DealCtxKey = "presenter.deal"

func (c *Deal) Render(w http.ResponseWriter, r *http.Request) error {
	if len(c.Products) > 0 {
		p := c.Products[0]
		p.Render(w, r)
		// product iterates and renders variants

		v := p.Variants[0]
		c.Title = p.Title

		oPrice := xchange.ToUSD(v.ProductVariant.Price, p.Place.Place.Currency)
		c.Description = fmt.Sprintf(
			"Retail price $%.2f. Deal price $%.2f -> %.f%% OFF!",
			oPrice,
			v.Price,
			100.0-(v.Price*100/oPrice),
		)

		if len(p.Images) > 0 {
			c.ImageURL = p.Images[0].ImageURL
			c.ImageWidth = p.Images[0].Width
			c.ImageHeight = p.Images[0].Height
		}
	}
	c.Cap = c.UsageLimit
	return nil
}

func NewDealList(ctx context.Context, deals []*data.Deal) []render.Renderer {
	list := []render.Renderer{}
	for _, c := range deals {
		//append to the final list to return
		list = append(list, NewDeal(ctx, c))
	}
	return list
}

func NewDeal(ctx context.Context, deal *data.Deal) *Deal {
	presented := &Deal{Deal: deal}

	// look up deal products by the parent deal id
	lookupID := deal.ID
	if deal.ParentID != nil {
		lookupID = *deal.ParentID
	}
	dps, err := data.DB.DealProduct.FindByDealID(lookupID)
	if err != nil {
		return presented
	}
	var productIDs []int64
	for _, p := range dps {
		productIDs = append(productIDs, p.ProductID)
	}
	products, err := data.DB.Product.FindAll(db.Cond{
		"id":     productIDs,
		"status": data.ProductStatusApproved,
	})
	if err != nil {
		return presented
	}

	// shove into context for later consumption
	ctx = context.WithValue(ctx, DealCtxKey, presented)
	presented.Products = newProductList(ctx, products)

	return presented
}
