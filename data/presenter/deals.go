package presenter

import (
	"context"
	"net/http"

	"bitbucket.org/moodie-app/moodie-api/data"
	"github.com/go-chi/render"
	db "upper.io/db.v3"
)

type Deal struct {
	*data.Deal
	Products []*Product `json:"products"`
	// TODO: image + sizing etc
}

const DealCtxKey = "presenter.deal"

func (c *Deal) Render(w http.ResponseWriter, r *http.Request) error {
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

	if deal.Status != data.DealStatusInactive {
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
		ctx = context.WithValue(ctx, DealCtxKey, deal)
		presented.Products = newProductList(ctx, products)
	}
	return presented
}
