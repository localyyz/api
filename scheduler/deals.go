package scheduler

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"sync"
	"time"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/lib/connect"
	"bitbucket.org/moodie-app/moodie-api/lib/onesignal"
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

	client, err := connect.GetShopifyClient(LocalyyzStoreId)
	client.Debug = true
	if err != nil {
		lg.Warnf("Get Price Rules: Failed to instantiate Shopify client for merchant %s", LocalyyzStoreId)
		return
	}

	productFemale, err := data.DB.Product.FindOne(
		db.Cond{
			"place_id":         LocalyyzStoreId,
			"status":           data.ProductStatusApproved,
			"gender":           data.ProductGenderFemale,
			"price":            db.Gte(100),
			db.Raw("random()"): db.Lt(0.5),
		})
	if err != nil {
		lg.Alert("Failed to get female product for automated deal of the day")
		return
	}

	productMale, err := data.DB.Product.FindOne(
		db.Cond{
			"place_id":         LocalyyzStoreId,
			"status":           data.ProductStatusApproved,
			"gender":           data.ProductGenderMale,
			"price":            db.Gte(100),
			db.Raw("random()"): db.Lt(0.5),
		})
	if err != nil {
		lg.Alert("Failed to get male product for automated deal of the day")
		return
	}

	dealProducts := []*data.Product{productFemale, productMale}

	var hour = []time.Duration{16 * time.Hour, 19 * time.Hour, 22 * time.Hour}
	i := rand.Intn(len(hour))

	//setting the start and end time for the deals
	start := time.Now().UTC().Truncate(24 * time.Hour).Add(24 * time.Hour).Add(hour[i])
	end := start.Add(1 * time.Hour)

	var toSend []data.Notification

	for _, product := range dealProducts {

		discount := "-70"

		//creating the price rule for the deal
		priceRule := &shopify.PriceRule{
			Title:              fmt.Sprintf("DOTD-%s-%s", time.Now().Add(24*time.Hour).Format("02-Jan-2006"), product.Gender),
			TargetType:         shopify.PriceRuleTargetTypeLineItem,
			TargetSelection:    shopify.PriceRuleTargetSelectionEntitled,
			ValueType:          shopify.PriceRuleValueTypeFixedAmount,
			Value:              discount,
			AllocationMethod:   "each",
			CustomerSelection:  "all",
			EntitledProductIds: []int64{*product.ExternalID},
			StartsAt:           start,
			EndsAt:             &end,
			AllocationLimit:    1,
			OncePerCustomer:    true,
			//UsageLimit: stock available
		}

		_, _, err = client.PriceRule.CreatePriceRule(context.Background(), priceRule)
		if err != nil {
			lg.Warn("Failed to create price rule")
			continue
		}

		discountCode := &shopify.DiscountCode{
			Code:        priceRule.Title,
			PriceRuleID: priceRule.ID,
		}

		_, _, err = client.DiscountCode.Create(context.Background(), discountCode)
		if err != nil {
			lg.Warn("Failed to create discount code")
			continue
		}

		ntf := data.Notification{
			ProductID: product.ID,
			Heading:   "⚡️ Deals of the Day! ⚡️",
			Content:   "Hurry in now to save $70 on great products 🤩. Deals end in one hour!",
		}

		toSend = append(toSend, ntf)

	}

	// if toSend has at least 1 notification, only create 1 push
	for i := range toSend {
		createPush(toSend[i], start)
		break
	}
}

func createPush(ntf data.Notification, startTime time.Time) {

	req := onesignal.NotificationRequest{
		Headings:         map[string]string{"en": ntf.Heading},
		Contents:         map[string]string{"en": ntf.Content},
		IncludedSegments: []string{"Subscribed Users"},
		SendAfter:        startTime.String(),
		Data:             map[string]string{"destination": "deal"},
	}

	resp, _, err := connect.ON.Notifications.Create(&req)
	if err != nil {
		lg.Warnf("failed to schedule notification: %v", err)
		return
	}

	ntf.ExternalID = resp.ID
	if err := data.DB.Notification.Save(&ntf); err != nil {
		lg.Warnf("failed to save notification to db: %v", err)
		return
	}

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
	createdAt := time.Now().Add(-24 * time.Hour).UTC()

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
			if len(rule.EntitledProductIds) == 0 {
				similarOrBetter, _ := data.DB.Deal.Find(db.Cond{
					"value":       db.Gte(deal.Value),
					"type":        deal.Type,
					"merchant_id": deal.MerchantID,
					"status":      []data.DealStatus{data.DealStatusQueued, data.DealStatusActive},
				}).Exists()
				if similarOrBetter {
					lg.Debugf("skipping similar deals: %d", deal.ExternalID)
					continue
				}
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

			// set deals of the day from localyyz store as featured and attach the product image
			if place.ID == LocalyyzStoreId && deal.ProductListType == data.ProductListTypeAssociated && len(rule.EntitledProductIds) > 0 {
				deal.Featured = true

				product, err := data.DB.Product.FindByExternalID(rule.EntitledProductIds[0])
				if err != nil {
					lg.Warnf("failed to fetch deal product for deal of the day with err %+v", err)
					continue
				}

				var image *data.ProductImage
				err = data.DB.ProductImage.Find(db.Cond{"product_id": product.ID}).OrderBy("ordering").One(&image)
				if err != nil {
					lg.Warnf("failed to fetch deal product image with err %+v", err)
					continue
				} else {
					deal.ImageURL = image.ImageURL
				}
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
