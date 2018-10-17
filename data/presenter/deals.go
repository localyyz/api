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
	Products         []*Product `json:"products"`
	PreReqProducts   []*Product `json:"prereqProducts"`
	EntitledProducts []*Product `json:"entitledProducts"`
	ProductList      []*Product `json:"ProductList"`
	MerchantProducts []*Product `json:"MerchantProducts"`

	ExternalID interface{} `json:"externalId,omitempty"`
	ParentID   interface{} `json:"parentId,omitempty"`
	UserID     interface{} `json:"userId,omitempty"`
	// Legacy + not implemented
	Cap int32 `json:"cap"`

	Title          string `json:"title"`
	Code           string `json:"code"`
	Description    string `json:"description"`
	Timed          bool   `json:"timed"`
	Info           string `json:"info"`
	Merchant       string `json:"merchant"`
	preReqQuantity int    `json:"preReqQuan,omitempty"`
	entQuantity    int    `json:"entQuan,omitempty"`
}

const DealCtxKey = "presenter.deal"

func (c *Deal) Render(w http.ResponseWriter, r *http.Request) error {

	if c.Type == data.DealTypePercentageOff {
		c.RenderDealTypePercentage(w, r)
	} else if c.Type == data.DealTypeAmountOff {
		c.RenderDealTypeAmount(w, r)
	} else if c.Type == data.DealTypeFreeShipping {
		c.RenderDealTypeFreeShipping(w, r)
	} else if c.Type == data.DealTypeBXGY {
		c.RenderDealTypeBXGY(w, r)
	} else if c.Prerequisite.QuantityRange >= 1 {
		c.Info = fmt.Sprintf("When you buy at least %d products from %s. ",
			c.Prerequisite.QuantityRange,
			c.Merchant)
	}

	if c.Prerequisite.SubtotalRange >= 1 {
		c.Info = fmt.Sprintf(c.Info+
			"When you spend at least $%d on %s products. ",
			c.Prerequisite.SubtotalRange/100,
			c.Merchant)
	}
	if c.BXGYPrerequisite.AllocationLimit >= 1 {
		c.Info = fmt.Sprintf(c.Info+
			"Up to %d use(s). ",
			c.BXGYPrerequisite.AllocationLimit)
	}
	if c.Prerequisite.ShippingPriceRange >= 1 {
		c.Info = fmt.Sprintf(c.Info+
			"For shipping fees under $%d. ",
			c.Prerequisite.ShippingPriceRange/100)
	}
	if c.OncePerCustomer == true {
		c.Info = fmt.Sprintf(c.Info + "One use per customer per order. ")
	}
	c.Cap = c.UsageLimit
	return nil
}

func (c *Deal) RenderDealTypePercentage(w http.ResponseWriter, r *http.Request) error {
	if c.ProductListType == data.ProductListTypeMerch {
		c.Title = c.Merchant
		c.Description = fmt.Sprintf(
			"%d%% Off Storewide",
			int(math.Abs(c.Value)))
		c.ProductList = c.MerchantProducts
	} else {
		if len(c.Products) > 1 {
			p := c.Products[0]
			p.Render(w, r)
			c.Title = "Discounted Products"
			c.Description = fmt.Sprintf(
				"%d%% Off Select Products",
				int(math.Abs(c.Value)))
			c.ProductList = c.Products
		} else if len(c.Products) == 1 {
			p := c.Products[0]
			p.Render(w, r)
			c.Title = fmt.Sprintf(
				"%s",
				p.Title)
			c.Description = fmt.Sprintf(
				"%d%% Off %s",
				int(math.Abs(c.Value)),
				p.Title)
			c.ProductList = c.Products
		} else {
			// something is wrong here. ie. associated product is not active
			// NOTE/TODO: what happens when a deal with associated product
			// is sold out / deleted. need scheduler to find and deactivate deals
		}
	}
	return nil
}

func (c *Deal) RenderDealTypeAmount(w http.ResponseWriter, r *http.Request) error {
	if c.ProductListType == data.ProductListTypeMerch {
		c.Title = c.Merchant
		c.Description = fmt.Sprintf(
			"$%d Off Storewide",
			int(math.Abs(c.Value)))
		c.ProductList = c.MerchantProducts
	} else {
		if len(c.Products) > 1 {
			p := c.Products[0]
			p.Render(w, r)
			c.Title = "Discounted Products"
			c.Description = fmt.Sprintf(
				"$%d Off Select Products",
				int(math.Abs(c.Value)))
			c.ProductList = c.Products
		} else if len(c.Products) == 1 {
			p := c.Products[0]
			p.Render(w, r)
			c.Title = fmt.Sprintf(
				"%s",
				p.Title)
			c.Description = fmt.Sprintf(
				"$%d Off %s",
				int(math.Abs(c.Value)),
				p.Title)
			c.ProductList = c.Products
		} else {
		}
	}
	return nil
}

func (c *Deal) RenderDealTypeFreeShipping(w http.ResponseWriter, r *http.Request) error {
	c.Title = "Free Shipping"
	c.Description = "Free Shipping"
	c.ProductList = c.MerchantProducts

	return nil
}

func (c *Deal) RenderDealTypeBXGY(w http.ResponseWriter, r *http.Request) error {
	if len(c.PreReqProducts) > 0 && len(c.EntitledProducts) > 0 {
		p := c.PreReqProducts[0]
		e := c.EntitledProducts[0]
		p.Render(w, r)
		e.Render(w, r)
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
			c.preReqQuantity,
			p.Title,
			c.entQuantity,
			entitledName,
			value)
	}
	c.ProductList = append(c.EntitledProducts, c.PreReqProducts...)
	return nil
}

func NewDealList(ctx context.Context, deals []*data.Deal) []render.Renderer {
	list := []render.Renderer{}
	for _, c := range deals {
		//append to the final list to return
		d := NewDeal(ctx, c)
		if d.ProductListType != data.ProductListTypeMerch && len(d.Products) == 0 {
			continue
		}
		list = append(list, d)
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
	if err != nil && err != db.ErrNoMoreRows {
		return presented
	}

	if deal.Timed == true {
		presented.Timed = true
	}

	preReqProdIDs := deal.BXGYPrerequisite.PrerequisiteProductIds

	preReqProds, err := data.DB.Product.FindAll(db.Cond{
		"id":     preReqProdIDs,
		"status": data.ProductStatusApproved,
	})

	entProdIDs := deal.BXGYPrerequisite.EntitledProductIds

	entProds, err := data.DB.Product.FindAll(db.Cond{
		"id":     entProdIDs,
		"status": data.ProductStatusApproved,
	})

	presented.Code = deal.Code

	merch, _ := data.DB.Place.FindByID(deal.MerchantID)
	presented.Merchant = merch.Name

	// shove into context for later consumption
	ctx = context.WithValue(ctx, DealCtxKey, presented)
	presented.Products = newProductList(ctx, products)
	presented.PreReqProducts = newProductList(ctx, preReqProds)
	presented.entQuantity = deal.BXGYPrerequisite.EntitledQuantityBXGY
	presented.EntitledProducts = newProductList(ctx, entProds)
	presented.preReqQuantity = deal.BXGYPrerequisite.PrerequisiteQuantityBXGY

	return presented
}
