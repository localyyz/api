package presenter

import (
	"context"
	"fmt"
	"net/http"
	"strings"

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
	Products         []*Product `json:"products"`
	PreReqProducts   []*Product `json:"prereqProducts"`
	EntitledProducts []*Product `json:"entitledProducts"`
}

const DealCtxKey = "presenter.deal"
const (
	DealInfoQuantityRange  = "When you buy at least %d products."
	DealInfoSubtotalRange  = "When you spend at least $%d."
	DealInfoShippingRange  = "For shipping fees under $%d."
	DealInfoBXGY           = "Up to %d use(s)."
	DealInfoUsePerCustomer = "One use per customer per order."
	DealInfoCondition      = "(other conditions may apply.)"
)

func (c *Deal) Render(w http.ResponseWriter, r *http.Request) error {
	switch c.Type {
	case data.DealTypePercentageOff, data.DealTypeAmountOff:
		c.renderDealTypeXOff()
	case data.DealTypeFreeShipping:
		c.Title = "Free Shipping"
		c.Description = "Free Shipping"
	case data.DealTypeBXGY:
		c.renderDealTypeBXGY()
	}

	infos := []string{}
	if x := c.Prerequisite.QuantityRange; x >= 1 {
		infos = append(infos, fmt.Sprintf(DealInfoQuantityRange, x))
	}
	if x := c.Prerequisite.SubtotalRange / 100; x >= 1 {
		infos = append(infos, fmt.Sprintf(DealInfoSubtotalRange, x))
	}
	if x := c.BXGYPrerequisite.AllocationLimit; x >= 1 {
		infos = append(infos, fmt.Sprintf(DealInfoBXGY, x))
	}
	if x := c.Prerequisite.ShippingPriceRange / 100; x >= 1 {
		infos = append(infos, fmt.Sprintf(DealInfoShippingRange, x))
	}
	if c.OncePerCustomer {
		infos = append(infos, DealInfoUsePerCustomer)
	}
	c.Info = strings.Join(append(infos, DealInfoCondition), "\n")

	c.Cap = c.UsageLimit
	return nil
}

func (c *Deal) renderDealTypeXOff() {
	amount := int(math.Abs(c.Value))
	val := ""
	if c.Type == data.DealTypePercentageOff {
		val = fmt.Sprintf("%d%%", amount)
	} else {
		val = fmt.Sprintf("$%d", amount)
	}

	if c.ProductListType == data.ProductListTypeMerch {
		// store wide discount
		c.Title = c.Place.Name
		c.Description = fmt.Sprintf("%s Off Storewide", val)
	} else {
		// product(s) discounted
		c.Title = "Discounted Products"
		c.Description = fmt.Sprintf("%s Off Select Products", val)
	}
}

func (c *Deal) renderDealTypeBXGY() {
	if len(c.PreReqProducts) > 0 && len(c.EntitledProducts) > 0 {
		p := c.PreReqProducts[0]
		e := c.EntitledProducts[0]
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
		if d.ProductListType != data.ProductListTypeMerch && len(d.Products) == 0 {
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
	cond := db.Cond{
		"place_id": deal.MerchantID,
		"status":   data.ProductStatusApproved,
	}
	if len(productIDs) != 0 {
		cond["id"] = productIDs
	}
	var products []*data.Product
	data.DB.Product.Find(cond).OrderBy("-id", "-score").Limit(5).All(&products)
	presented.Products = newProductList(ctx, products)

	// prerequiste products (for bxgy)
	if IDs := deal.BXGYPrerequisite.PrerequisiteProductIds; len(IDs) > 0 {
		products, _ := data.DB.Product.FindAll(db.Cond{
			"id":     IDs,
			"status": data.ProductStatusApproved,
		})
		presented.PreReqProducts = newProductList(ctx, products)
	}

	// entitled products (for bxgy)
	if IDs := deal.BXGYPrerequisite.EntitledProductIds; len(IDs) > 0 {
		products, _ := data.DB.Product.FindAll(db.Cond{
			"id":     IDs,
			"status": data.ProductStatusApproved,
		})
		presented.EntitledProducts = newProductList(ctx, products)
	}

	// fill in the merchant
	place, _ := data.DB.Place.FindByID(deal.MerchantID)
	presented.Place = &Place{Place: place}

	// shove into context for later consumption
	ctx = context.WithValue(ctx, DealCtxKey, presented)

	return presented
}
