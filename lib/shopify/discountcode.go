package shopify

import "time"

type DiscountCode struct {
	ID          int64      `json:"id"`
	PriceRuleID int64      `json:"price_rule_id"`
	Code        string     `json:"code"`
	UsageCount  int32      `json:"usage_count"`
	CreatedAt   *time.Time `json:"created_at"`
	UpdatedAt   *time.Time `json:"updated_at"`
}

type AppliedDiscount struct {
	Amount              string `json:"amount"`
	Title               string `json:"title"`
	Description         string `json:"description"`
	Value               string `json:"value"`
	ValueType           string `json:"value_type"`
	Applicable          bool   `json:"applicable"`
	NonApplicableReason string `json:"non_applicable_reason,omitempty"`
}
