package presenter

import (
	"context"
	"fmt"
	"net/http"

	"bitbucket.org/moodie-app/moodie-api/data"

	"math"

	"github.com/go-chi/render"
	"upper.io/db.v3"
)

type Deal struct {
	*data.Deal
	Place *Place `json:"place"`

	ExternalID interface{} `json:"externalId,omitempty"`
	ParentID   interface{} `json:"parentId,omitempty"`
	UserID     interface{} `json:"userId,omitempty"`
	// Legacy + not implemented
	Cap int32 `json:"cap"`

	Title       string `json:"title"`
	Description string `json:"description"`
	Info        string `json:"info"`

	// for putting together into and titles
	products         []*Product `json:"products"`
	preReqProducts   []*Product `json:"prereqProducts"`
	entitledProducts []*Product `json:"entitledProducts"`
}

const DealCtxKey = "presenter.deal"

func (c *Deal) Render(w http.ResponseWriter, r *http.Request) error {

	switch c.Type {
	case data.DealTypePercentageOff:
		c.RenderDealTypePercentage()
	case data.DealTypeAmountOff:
		c.RenderDealTypeAmount()
	case data.DealTypeFreeShipping:
		c.Title = "Free Shipping"
		c.Description = "Free Shipping"
	case data.DealTypeBXGY:
		c.RenderDealTypeBXGY()

	}

	if c.Prerequisite.QuantityRange >= 1 {
		c.Info = fmt.Sprintf("When you buy at least %d products from %s.", c.Prerequisite.QuantityRange, c.Place.Name)
	}
	if c.Prerequisite.SubtotalRange >= 1 {
		c.Info = fmt.Sprintf("%s\nWhen you spend at least $%d on %s products.", c.Info, c.Prerequisite.SubtotalRange/100, c.Place.Name)
	}
	if c.BXGYPrerequisite.AllocationLimit >= 1 {
		c.Info = fmt.Sprintf("%s\nUp to %d use(s).", c.Info, c.BXGYPrerequisite.AllocationLimit)
	}
	if c.Prerequisite.ShippingPriceRange >= 1 {
		c.Info = fmt.Sprintf("%s\nFor shipping fees under $%d.", c.Info, c.Prerequisite.ShippingPriceRange/100)
	}
	if c.OncePerCustomer == true {
		c.Info = fmt.Sprintf("%s\nOne use per customer per order.", c.Info)
	}
	c.Info = fmt.Sprintf("%s\nConditions may apply.", c.Info)

	c.Cap = c.UsageLimit
	return nil
}

func (c *Deal) RenderDealTypePercentage() {
	// store wide discount by percentage
	if c.ProductListType == data.ProductListTypeMerch {
		c.Title = c.Place.Name
		c.Description = fmt.Sprintf("%d%% Off Storewide", int(math.Abs(c.Value)))
		return
	}
	// product(s) discounted by percentage
	c.Title = "Discounted Products"
	c.Description = fmt.Sprintf("%d%% Off Select Products", int(math.Abs(c.Value)))
}

func (c *Deal) RenderDealTypeAmount() {
	// store wide discount by dollar value
	if c.ProductListType == data.ProductListTypeMerch {
		c.Title = c.Place.Name
		c.Description = fmt.Sprintf("$%d Off Storewide", int(math.Abs(c.Value)))
		return
	}
	// product(s) discounted by dollar value
	c.Title = "Discounted Products"
	c.Description = fmt.Sprintf("$%d Off Select Products", int(math.Abs(c.Value)))
}

func (c *Deal) RenderDealTypeBXGY() {
	if len(c.preReqProducts) > 0 && len(c.entitledProducts) > 0 {
		p := c.preReqProducts[0]
		e := c.entitledProducts[0]
		c.Title = "Discounted Products"
		var value string
		if c.Value == 100 {
			value = "FREE"
		} else {
			value = fmt.Sprintf("%d%% OFF", int(math.Abs(c.Value)))
		}
		var entitledName string
		if p.ID == e.ID {
			entitledName = ""
		} else {
			entitledName = e.Title
		}
		c.Description = fmt.Sprintf(
			"Buy %d %s, get %d %s for %s",
			c.BXGYPrerequisite.PrerequisiteQuantityBXGY,
			p.Title,
			c.BXGYPrerequisite.EntitledQuantityBXGY,
			entitledName,
			value)
	}
}

func NewDealList(ctx context.Context, deals []*data.Deal) []render.Renderer {
	list := []render.Renderer{}
	for _, c := range deals {
		//append to the final list to return
		d := NewDeal(ctx, c)
		if d.ProductListType != data.ProductListTypeMerch && len(d.products) == 0 {
			continue
		}
		list = append(list, d)
	}
	return list
}

func NewDeal(ctx context.Context, deal *data.Deal) *Deal {
	presented := &Deal{Deal: deal}

	dps, err := data.DB.DealProduct.FindByDealID(deal.ID)
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
	if err != nil && err != db.ErrNoMoreRows {
		return presented
	}

	// prerequiste products (for bxgy)
	preReqProdIDs := deal.BXGYPrerequisite.PrerequisiteProductIds
	preReqProds, err := data.DB.Product.FindAll(db.Cond{
		"id":     preReqProdIDs,
		"status": data.ProductStatusApproved,
	})

	// entitled products (for bxgy)
	entProdIDs := deal.BXGYPrerequisite.EntitledProductIds
	entProds, err := data.DB.Product.FindAll(db.Cond{
		"id":     entProdIDs,
		"status": data.ProductStatusApproved,
	})

	// fill in the merchant
	place, _ := data.DB.Place.FindByID(deal.MerchantID)
	presented.Place = &Place{Place: place}

	// shove into context for later consumption
	ctx = context.WithValue(ctx, DealCtxKey, presented)
	presented.products = newProductList(ctx, products)
	presented.preReqProducts = newProductList(ctx, preReqProds)
	presented.entitledProducts = newProductList(ctx, entProds)

	return presented
}
