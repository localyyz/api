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
	ID                int64  `json:"id"`
	Title             string `json:"title"`
	ValueType         string `json:"value_type"`
	Value             string `json:"value"`
	CustomerSelection string `json:"customer_selection"`
	TargetType        string `json:"target_type"`
	TargetSelection   string `json:"target_selection"`
	AllocationMethod  string `json:"allocation_method"`
	OncePerCustomer   bool   `json:"once_per_customer"`
	UsageLimit        int    `json:"usage_limit"`

	EntitledProductIds    []int64 `json:"entitled_product_ids"`
	EntitledVariantIds    []int64 `json:"entitled_variant_ids"`
	EntitledCollectionIds []int64 `json:"entitled_collection_ids"`
	EntitledCountryIds    []int64 `json:"entitled_country_ids"`

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

	StartsAt  time.Time  `json:"starts_at"`
	EndsAt    *time.Time `json:"ends_at"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

type PriceRuleParam struct {
	Limit     int        `json:"limit"`
	Page      int        `json:"page"`
	EndsAtMin *time.Time `json:"ends_at_min"`
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

type PriceRuleValueType uint32

const (
	_ PriceRuleValueType = iota
	PriceRuleValueTypeFixedAmount
	PriceRuleValueTypePercentage
)

var (
	priceRuleValueTypes = []string{"-", "fixed_amount", "percentage"}
)

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

// String returns the string value of the status.
func (s PriceRuleValueType) String() string {
	return priceRuleValueTypes[s]
}

// MarshalText satisfies TextMarshaler
func (s PriceRuleValueType) MarshalText() ([]byte, error) {
	return []byte(s.String()), nil
}

// UnmarshalText satisfies TextUnmarshaler
func (s *PriceRuleValueType) UnmarshalText(text []byte) error {
	enum := string(text)
	for i := 0; i < len(priceRuleValueTypes); i++ {
		if enum == priceRuleValueTypes[i] {
			*s = PriceRuleValueType(i)
			return nil
		}
	}
	return fmt.Errorf("unknown value type %s", enum)
}
