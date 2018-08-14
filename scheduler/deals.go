package scheduler

import (
	"context"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/lib/shopify"
	"github.com/pressly/lg"
	db "upper.io/db.v3"
)

const LocalyyzStoreId = 4164
const DealDuration = time.Hour

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

func (h *Handler) SyncDeals() {
	h.wg.Add(1)
	defer h.wg.Done()

	s := time.Now()
	lg.Info("job_sync_deals running...")
	defer func() {
		lg.Infof("job_sync_deals finished in %s", time.Since(s))
	}()

	ctx := context.Background()
	// getting the shopify cred
	cred, err := data.DB.ShopifyCred.FindOne(db.Cond{"place_id": LocalyyzStoreId})
	if err != nil {
		lg.Alert("Sync deals: Failed to get Shopify Credentials")
		return
	}

	// creating the client
	client := shopify.NewClient(nil, cred.AccessToken)
	client.BaseURL, err = url.Parse(cred.ApiURL)
	if err != nil {
		lg.Alert("Sync deals: Failed to instantiate Shopify client")
		return
	}

	startAt := time.Now().UTC().Truncate(time.Hour)
	endAt := time.Now().Add(24 * time.Hour)
	params := &shopify.PriceRuleParam{
		StartsAtMin: &startAt,
		EndsAtMax:   &endAt,
	}

	// fetch the products
	priceRules, resp, err := client.PriceRule.List(ctx, params)
	if err != nil || resp.StatusCode != http.StatusOK {
		lg.Alert("Sync deals: Failed to get price rules from Shopify")
		return
	}

	for _, rule := range priceRules {
		// if not a specific product deal, skip.
		// NOTE, in the future we can expand to all deals
		if len(rule.EntitledProductIds) == 0 {
			continue
		}
		// for now, only handle fixed amount deals. ie ($30 off)
		if rule.ValueType != shopify.PriceRuleValueTypeFixedAmount {
			continue
		}
		// check if deal already exist
		deal, err := data.DB.Deal.FindOne(db.Cond{"external_id": rule.ID})
		if err != nil {
			if err != db.ErrNoMoreRows {
				lg.Alertf("failed to fetch deal with err %+v", err)
				continue
			}

			// for now, parse as dollar amount
			value, err := strconv.ParseFloat(rule.Value, 64)
			if err != nil {
				lg.Alertf("failed to parse deal value with err %+v", err)
				continue
			}

			startAt := rule.StartsAt.UTC()

			deal = &data.Deal{
				ExternalID:      rule.ID,
				Status:          data.DealStatusQueued,
				MerchantID:      LocalyyzStoreId,
				StartAt:         &startAt,
				Code:            rule.Title,
				Value:           value,
				UsageLimit:      int32(rule.UsageLimit),
				OncePerCustomer: rule.OncePerCustomer,
			}
			if e := rule.EndsAt; e != nil {
				endAt := e.UTC()
				deal.EndAt = &endAt
			} else {
				// if ends at is not set, update it to default
				endAt := deal.StartAt.Add(DealDuration)
				deal.EndAt = &endAt
			}

			//save the deal.
			if err := data.DB.Deal.Save(deal); err != nil {
				lg.Alertf("failed to save deal with err %+v", err)
				continue
			}
		}

		// find the products with the external ids
		products, err := data.DB.Product.FindAll(db.Cond{"external_id": rule.EntitledProductIds})
		if err != nil {
			lg.Alertf("failed to fetch deal products with err %+v", err)
			continue
		}
		if len(products) != len(rule.EntitledProductIds) {
			lg.Alertf("failed to fetch deal products with extIDs %+v", rule.EntitledProductIds)
			continue
		}

		batch := data.DB.
			InsertInto("deal_products").
			Columns("deal_id", "product_id").
			Batch(5)

		go func() {
			defer batch.Done()
			for _, p := range products {
				batch.Values(deal.ID, p.ID)
			}
		}()

		if err := batch.Wait(); err != nil {
			lg.Alertf("failed to insert deal products with err %+v", err)
			continue
		}
		lg.Infof("inserted %d products for deal %s", len(products), deal.Code)
	}
}
