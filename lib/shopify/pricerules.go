package shopify

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

type PriceRuleService service

type PriceRule struct {
	ID                int64                    `json:"id"`
	Title             string                   `json:"title"`
	ValueType         PriceRuleValueType       `json:"value_type"`
	Value             string                   `json:"value"`
	CustomerSelection string                   `json:"customer_selection"`
	TargetType        PriceRuleTargetType      `json:"target_type"`
	TargetSelection   PriceRuleTargetSelection `json:"target_selection"`
	AllocationMethod  string                   `json:"allocation_method"`
	OncePerCustomer   bool                     `json:"once_per_customer"`
	UsageLimit        int                      `json:"usage_limit"`

	EntitledProductIds    []int64 `json:"entitled_product_ids"`
	EntitledVariantIds    []int64 `json:"entitled_variant_ids"`
	EntitledCollectionIds []int64 `json:"entitled_collection_ids"`
	EntitledCountryIds    []int64 `json:"entitled_country_ids"`

	// Prefreq for BUY X GET Y type deals
	PrerequisiteSavedSearchIds []int64 `json:"prerequisite_saved_search_ids"`
	PrerequisiteCustomerIds    []int64 `json:"prerequisite_customer_ids"`
	PrerequisiteSubtotalRange  struct {
		Gte string `json:"greater_than_or_equal_to"`
	} `json:"prerequisite_subtotal_range"`
	PrerequisiteShippingPriceRange struct {
		Lte string `json:"less_than_or_equal_to"`
	} `json:"prerequisite_shipping_price_range"`
	PrerequisiteQuantityRange struct {
		Gte string `json:"greater_than_or_equal_to"`
	} `json:"prerequisite_quantity_range"`

	PrerequisiteQuantityRatio struct {
		Quantity         int `json:"prerequisite_quantity"`
		EntitledQuantity int `json:"entitled_quantity"`
	} `json:"prerequisite_to_entitlement_quantity_ratio"`

	PrerequisiteProductIDs    []int64 `json:"prerequisite_product_ids"`
	PrerequisiteVariantIDs    []int64 `json:"prerequisite_variant_ids"`
	PrerequisiteCollectionIDs []int64 `json:"prerequisite_collection_ids"`

	StartsAt  time.Time  `json:"starts_at"`
	EndsAt    *time.Time `json:"ends_at"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

type PriceRuleTargetSelection string
type PriceRuleTargetType string
type PriceRuleValueType string

const (
	PriceRuleTargetSelectionAll      PriceRuleTargetSelection = "all"
	PriceRuleTargetSelectionEntitled                          = "entitled"

	PriceRuleTargetTypeLineItem     PriceRuleTargetType = "line_item"     // The price rule applies to the cart's line items
	PriceRuleTargetTypeShippingLine                     = "shipping_line" // The price rule applies to the cart's shipping lines

	PriceRuleValueTypeFixedAmount PriceRuleValueType = "fixed_amount"
	PriceRuleValueTypePercentage                     = "percentage"
)

type PriceRuleParam struct {
	Limit int `json:"limit"`
	Page  int `json:"page"`
	// Show rule starting AFTER date
	StartsAtMin *time.Time `json:"starts_at_min"`
	// Show rule starting BEFORE date
	StartsAtMax *time.Time `json:"starts_at_max"`
	// Show rule ending AFTER date
	EndsAtMin *time.Time `json:"ends_at_min"`
	// Show rule ending BEFORE date
	EndsAtMax *time.Time `json:"ends_at_max"`
	SinceID   int64      `json:"since_id"`
	TimesUsed int        `json:"times_used"`
}

func (p *PriceRuleParam) EncodeQuery() string {
	if p == nil {
		return ""
	}
	// for now just allow handle
	// TODO: support all params
	v := url.Values{}

	if p.Limit > 0 {
		v.Add("limit", fmt.Sprintf("%d", p.Limit))
	}
	if p.Page > 0 {
		v.Add("page", fmt.Sprintf("%d", p.Page))
	}
	if p.EndsAtMin != nil {
		v.Add("ends_at_min", p.EndsAtMin.Format(timeFormat))
	}
	return v.Encode()
}

func (p *PriceRuleService) List(ctx context.Context, params *PriceRuleParam) ([]*PriceRule, *http.Response, error) {
	req, err := p.client.NewRequest("GET", "/admin/price_rules.json", nil)
	if err != nil {
		return nil, nil, err
	}
	req.URL.RawQuery = params.EncodeQuery()

	var priceRuleWrapper struct {
		PriceRules []*PriceRule `json:"price_rules"`
	}
	resp, err := p.client.Do(ctx, req, &priceRuleWrapper)
	if err != nil {
		return nil, resp, err
	}

	return priceRuleWrapper.PriceRules, resp, nil
}

func (p *PriceRuleService) Get(ctx context.Context, ID int64) (*PriceRule, *http.Response, error) {
	req, err := p.client.NewRequest("GET", fmt.Sprintf("/admin/price_rules/%d.json", ID), nil)
	if err != nil {
		return nil, nil, err
	}

	var priceRuleWrapper struct {
		PriceRule *PriceRule `json:"price_rule"`
	}
	resp, err := p.client.Do(ctx, req, &priceRuleWrapper)
	if err != nil {
		return nil, resp, err
	}

	return priceRuleWrapper.PriceRule, resp, nil
}
