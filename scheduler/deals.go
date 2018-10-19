package scheduler

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/lib/connect"
	"bitbucket.org/moodie-app/moodie-api/lib/shopify"
	"github.com/pressly/lg"
	"upper.io/db.v3"
)

const LocalyyzStoreId = 4164

func (h *Handler) ScheduleDeals() {
	h.wg.Add(1)
	defer h.wg.Done()

	s := time.Now()
	lg.Info("job_schedule_deals running...")
	defer func() {
		lg.Infof("job_schedule_deals finished in %s", time.Since(s))
	}()

	// expire deals
	h.DB.Exec(`UPDATE deals SET status = 2 WHERE NOW() at time zone 'utc' > end_at and status = 3`)

	// activate deals
	h.DB.Exec(`UPDATE deals SET status = 3 WHERE NOW() at time zone 'utc' > start_at and status = 1`)
}

func (h *Handler) CreateDealOfTheDay() {
	h.wg.Add(1)
	defer h.wg.Done()

	s := time.Now()
	lg.Info("job_create_deal_of_day running...")
	defer func() {
		lg.Infof("job_create_deal_of_day finished in %s", time.Since(s))
	}()

	client, err := connect.GetShopifyClient(2260)
	client.Debug = true
	if err != nil {
		lg.Warnf("Get Price Rules: Failed to instantiate Shopify client for merchant %s", LocalyyzStoreId)
		return
	}

	products, err := data.DB.Product.FindAll(
		db.Cond{
			"place_id": 2260,
			"status": data.ProductStatusApproved,
		})
	if err != nil {
		lg.Warn("Failed to get Localyyz products")
	}

	dealProduct := chooseDealProduct(products)

	lg.Print(dealProduct.ExternalID)

	var discount string

	if dealProduct.Price >= 100 {
		discount = "-70"
	} else {
		discount = "-40"
	}


	start := time.Now().Add(8 * time.Hour).UTC()
	end := start.Add(1*time.Hour)

	priceRule := &shopify.PriceRule{
		Title: fmt.Sprintf("DOTD-%s", time.Now().Format("02-Jan-2006")),
		TargetType: shopify.PriceRuleTargetTypeLineItem,
		TargetSelection: shopify.PriceRuleTargetSelectionAll,
		ValueType: shopify.PriceRuleValueTypeFixedAmount,
		Value: discount,
		AllocationMethod: "each",
		CustomerSelection: "all",
		EntitledProductIds: []int64 {*dealProduct.ExternalID},
		StartsAt: start,
		EndsAt: &end,
		AllocationLimit: 1,
		OncePerCustomer: true,
		//UsageLimit: stock available
	}

	DealMetrics()

	lg.Print(priceRule)

	client.PriceRule.CreatePriceRule(context.Background(), priceRule)

}

func chooseDealProduct(productList []*data.Product) (*data.Product) {
	return productList[40]
}

func DealMetrics() {

}

func (h *Handler) SyncDiscountCodes() {
	h.wg.Add(1)
	defer h.wg.Done()

	s := time.Now()
	lg.Info("job_get_price_rules running...")
	defer func() {
		lg.Infof("job_get_price_rules finished in %s", time.Since(s))
	}()

	// get all active merchants which may have discounts
	places, err := data.DB.Place.FindAll(
		db.Cond{
			"status": data.PlaceStatusActive,
		},
	)
	if err != nil {
		lg.Warn("Get Price Rules: Failed to get merchant list")
		return
	}

	var wg sync.WaitGroup
	for _, place := range places {
		wg.Add(1)

		p := place
		go parseDeals(context.Background(), p, &wg)
	}

	lg.Info("finishing up scheduler job")
	wg.Wait()
}

func parseDeals(ctx context.Context, place *data.Place, wg *sync.WaitGroup) {
	defer func() {
		lg.Infof("finished for merchant(%s)", place.Name)
		wg.Done()
	}()

	count, _ := data.DB.Product.Find(db.Cond{
		"place_id": place.ID,
		"status":   data.ProductStatusApproved,
	}).Count()

	if count == 0 {
		// skip the entire process. because we have no products
		return
	}

	page := 1
	createdAt := time.Now().Add(24 * time.Hour).UTC()

	client, err := connect.GetShopifyClient(place.ID)
	if err != nil {
		lg.Warnf("Get Price Rules: Failed to instantiate Shopify client for merchant %s", place.ID)
		return
	}

	for {
		// find all price rules created in the past 24 hours.
		params := &shopify.PriceRuleParam{
			Limit:        50,
			CreatedAtMin: &createdAt,
			Page:         page,
		}

		// fetch price rule list
		priceRules, resp, err := client.PriceRule.List(ctx, params)
		if err != nil || resp.StatusCode != http.StatusOK {
			lg.Warnf("Get Price Rules: Failed to get price rules from Shopify for merchant %d", place.ID)
			return
		}

		// if nothing to parse. skip.
		if len(priceRules) == 0 {
			break
		}

		//increment pages until all valid rules are added to the list
		page++

		for _, rule := range priceRules {
			// only want deals that are applicable to all users
			if rule.CustomerSelection != "all" {
				continue
			}

			// only want deals that does not limit usage
			if rule.UsageLimit != 0 {
				continue
			}

			// parse value amount of deal
			value, err := strconv.ParseFloat(rule.Value, 64)
			if err != nil {
				lg.Warnf("failed to parse deal value with err %+v", err)
				continue
			}

			// skip deals that have no discount value (0% Off)
			if value == 0 {
				continue
			}

			var dealType data.DealType

			// check conditions for deal to be BXGY
			if rule.TargetType == "line_item" &&
				rule.TargetSelection == "entitled" &&
				rule.AllocationMethod == "each" &&
				rule.PrerequisiteQuantityRatio.Quantity != 0 &&
				rule.PrerequisiteQuantityRatio.EntitledQuantity != 0 {

				// for now, ignore BXGY deals that are for collections or other product list types.
				// only include deals that have prereq and entitled product ids
				dealType = data.DealTypeBXGY
				if len(rule.PrerequisiteProductIDs) > 0 &&
					len(rule.EntitledProductIds) > 0 {
					// great. let's continue
				} else {
					// skip. for now skip any BYGX deals that are not prereq
					// with specific products
					continue
				}
				// if target type is shipping line, deal is free shipping
			} else if rule.TargetType == shopify.PriceRuleTargetTypeShippingLine {
				dealType = data.DealTypeFreeShipping

				// if rule value type is fixed amount, deal is a dollar amount off
			} else if rule.ValueType == shopify.PriceRuleValueTypeFixedAmount {
				dealType = data.DealTypeAmountOff

				// if rule value type is a percentage, deal is a percentage off
			} else if rule.ValueType == shopify.PriceRuleValueTypePercentage {
				dealType = data.DealTypePercentageOff
			} else {
				continue
			}

			startAt := rule.StartsAt.UTC()
			deal := &data.Deal{
				ExternalID:      rule.ID,
				Status:          data.DealStatusQueued,
				MerchantID:      place.ID,
				StartAt:         &startAt,
				Value:           value,
				Type:            dealType,
				Code:            rule.Title,
				UsageLimit:      int32(rule.UsageLimit),
				OncePerCustomer: rule.OncePerCustomer,
				Prerequisite: data.DealPrerequisite{
					// setting the prerequisite quantity range of the rule to the deal database
					QuantityRange: rule.PrerequisiteQuantityRange.Gte,
				},
				BXGYPrerequisite: data.BXGYDealPrerequisite{},
			}
			// set deal status based on start time
			if startAt.Before(time.Now()) {
				deal.Status = data.DealStatusActive
			}
			// if deal has an end time, set Timed to true
			if e := rule.EndsAt; e != nil {
				endAt := e.UTC()
				if endAt.Before(time.Now()) {
					continue
				}
				deal.EndAt = &endAt
				deal.Timed = true
			}

			deal.Featured = false

			// if the deal matches the value and type of another deal from that merchant, skip
			// A similar deal has:
			//  - equal or betteer value
			//  - same type of deal (value off, pct off etc)
			//  - status: either queued or active
			similarOrBetter, _ := data.DB.Deal.Find(db.Cond{
				"value":       db.Gte(deal.Value),
				"type":        deal.Type,
				"merchant_id": deal.MerchantID,
				"status":      []data.DealStatus{data.DealStatusQueued, data.DealStatusActive},
			}).Exists()
			if similarOrBetter {
				continue
			}

			// if rule id already exists in database, update the deal with the new parameters
			if d, _ := data.DB.Deal.FindOne(db.Cond{"external_id": rule.ID}); d != nil {
				deal.ID = d.ID
			}

			// setting the shipping range of the rule to the deal database, in cents
			ruleShippingRange, _ := strconv.ParseFloat(rule.PrerequisiteShippingPriceRange.Lte, 64)
			ruleShippingRangeInCents := ruleShippingRange * 100
			deal.Prerequisite.ShippingPriceRange = int(ruleShippingRangeInCents)

			// setting the subtotal range of the rule to the deal database, in cents
			ruleSubtotalRange, _ := strconv.ParseFloat(rule.PrerequisiteSubtotalRange.Gte, 64)
			ruleSubtotalRangeInCents := ruleSubtotalRange * 100
			deal.Prerequisite.SubtotalRange = int(ruleSubtotalRangeInCents)

			// setting the BXGY prerequisites from the rule to the deal database
			if dealType == data.DealTypeBXGY {
				deal.BXGYPrerequisite.AllocationLimit = rule.AllocationLimit
				deal.BXGYPrerequisite.EntitledProductIds = rule.EntitledProductIds
				deal.BXGYPrerequisite.PrerequisiteProductIds = rule.PrerequisiteProductIDs
				deal.BXGYPrerequisite.EntitledQuantityBXGY = rule.PrerequisiteQuantityRatio.EntitledQuantity
				deal.BXGYPrerequisite.PrerequisiteQuantityBXGY = rule.PrerequisiteQuantityRatio.Quantity
				deal.ProductListType = data.ProductListTypeBXGY
			}

			// tile of the price rule is not necessary the actual discount code.
			// fetch the list of discount code and iterate untile we get on
			dealCodes, _, _ := client.PriceRule.ListDiscountCodes(ctx, rule.ID)
			for i := range dealCodes {
				if dealCodes[i].Code != "" && dealCodes[i].Code != deal.Code {
					deal.Code = dealCodes[i].Code
					break
				}
			}

			// if entitled product id is empty -> we are targeting entire
			// merchant collection
			// TODO: is that true? what about entitled collections...?
			deal.ProductListType = data.ProductListTypeMerch
			if len(rule.EntitledProductIds) > 0 {
				deal.ProductListType = data.ProductListTypeAssociated
			}

			if deal.ProductListType == data.ProductListTypeUnKnown {
				continue
			}

			// NOTE:
			// if: the deal is $$ off (amount off)
			// and:
			// - no product associated (store wide deal)
			// - no prereq. (ie. min subtotal spend, min quantity buy)
			// then: mark status as "pending for approval"
			if deal.ProductListType == data.ProductListTypeMerch &&
				deal.Type == data.DealTypeAmountOff &&
				deal.Prerequisite.SubtotalRange == 0 &&
				deal.Prerequisite.QuantityRange == 0 {
				deal.Status = data.DealStatusPending
			}

			// save the deal.
			if err := data.DB.Deal.Save(deal); err != nil {
				lg.Warnf("failed to save deal with err %+v", err)
			}

			if len(rule.EntitledProductIds) > 0 {
				// find the products with the external ids
				products, err := data.DB.Product.FindAll(db.Cond{"external_id": rule.EntitledProductIds})
				if err != nil {
					lg.Warnf("failed to fetch deal products with err %+v", err)
					continue
				}
				if len(products) != len(rule.EntitledProductIds) {
					lg.Warnf("failed to fetch deal products with extIDs %+v", rule.EntitledProductIds)
					continue
				}

				batch := data.DB.
					InsertInto("deal_products").
					Columns("deal_id", "product_id").
					Amend(func(queryIn string) (queryOut string) {
						queryOut = fmt.Sprintf("%s ON CONFLICT DO NOTHING", queryIn)
						return
					}).Batch(5)

				go func() {
					defer batch.Done()
					for _, p := range products {
						batch.Values(deal.ID, p.ID)
					}
				}()

				if err := batch.Wait(); err != nil {
					lg.Warnf("failed to insert deal products with err %+v", err)
					continue
				}
				lg.Infof("inserted %d products for deal %s", len(products), deal.Code)
			}
		}

	}
}
