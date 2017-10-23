package shopify

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"
)

type BillingService service

type Billing struct {
	ID              int64  `json:"id"`
	Name            string `json:"name"`
	Price           string `json:"price"`
	ReturnUrl       string `json:"return_url"`
	ConfirmationUrl string `json:"confirmation_url"`
	CappedAmount    int64  `json:"capped_amount,omitempty"`
	Terms           string `json:"terms"`

	Type   BillingType   `json:"type"`
	Status BillingStatus `json:"status"`

	TrialDays int64 `json:"trial_days"`

	Test bool `json:"test"`

	CreatedAt *time.Time `json:"created_at"`
	UpdatedAt *time.Time `json:"updated_at"`

	// formated as "YYYY-mm-dd"
	TrialEndsOn string `json:"trial_ends_on"`
	ActivatedOn string `json:"activated_on"`
	BillingOn   string `json:"billing_on"`
	CancelledOn string `json:"cancelled_on"`
}

type BillingType uint32
type BillingStatus uint32

type RecurringBillingRequest struct {
	Billing *Billing `json:"recurring_application_charge"`
}

const (
	BillingTypeUnknown = iota
	BillingTypeOneTime
	BillingTypeRecurring
	BillingTypeUsage
)

const (
	BillingStatusUnknown = iota
	// pending: The recurring charge is pending.
	BillingStatusPending
	// accepted: The recurring charge has been accepted.
	BillingStatusAccepted
	// active: The recurring charge is activated.
	// This is the only status that actually causes a merchant to be charged.
	// An accepted charge is transitioned to active via the activate endpoint.
	BillingStatusActive
	// declined: The recurring charge has been declined.
	BillingStatusDeclined
	// expired: The recurring charge was not accepted within 2 days of being created.
	BillingStatusExpired
	// frozen: The recurring charge is on hold due to a shop subscription non-payment.
	// The charge will re-activate once subscription payments resume.
	BillingStatusFrozen
	// cancelled: The developer cancelled the charge.
	BillingStatusCancelled
)

var (
	billingTypes = []string{
		"-",
		"one_time",
		"recurring",
		"usage",
	}

	billingStatuses = []string{
		"-",
		"pending",
		"accepted",
		"active",
		"declined",
		"expired",
		"frozen",
		"cancelled",
	}

	billingApiPath = map[BillingType]string{
		BillingTypeRecurring: "recurring_application_charges",
	}

	ErrUnsupportedBillingType = errors.New("unsupported billing type")
)

func (c *BillingService) Create(ctx context.Context, billing *Billing) (*Billing, *http.Response, error) {
	apiPath, found := billingApiPath[billing.Type]
	if !found {
		return nil, nil, ErrUnsupportedBillingType
	}

	req, err := c.client.NewRequest(
		"POST",
		fmt.Sprintf("/admin/%s.json", apiPath),
		&RecurringBillingRequest{billing},
	)
	if err != nil {
		return nil, nil, err
	}

	billingWrapper := &RecurringBillingRequest{billing}
	resp, err := c.client.Do(ctx, req, billingWrapper)
	if err != nil {
		return nil, resp, err
	}

	return billingWrapper.Billing, resp, nil
}

func (c *BillingService) Get(ctx context.Context, billing *Billing) (*Billing, *http.Response, error) {
	apiPath, found := billingApiPath[billing.Type]
	if !found {
		return nil, nil, ErrUnsupportedBillingType
	}

	req, err := c.client.NewRequest(
		"GET",
		fmt.Sprintf("/admin/%s/%d.json", apiPath, billing.ID),
		nil,
	)
	if err != nil {
		return nil, nil, err
	}

	billingWrapper := &RecurringBillingRequest{billing}
	resp, err := c.client.Do(ctx, req, billingWrapper)
	if err != nil {
		return nil, resp, err
	}

	return billingWrapper.Billing, resp, nil
}

func (c *BillingService) Activate(ctx context.Context, billing *Billing) (*Billing, *http.Response, error) {
	apiPath, found := billingApiPath[billing.Type]
	if !found {
		return nil, nil, ErrUnsupportedBillingType
	}

	req, err := c.client.NewRequest(
		"POST",
		fmt.Sprintf("/admin/%s/%d/activate.json", apiPath, billing.ID),
		&RecurringBillingRequest{billing},
	)
	if err != nil {
		return nil, nil, err
	}

	billingWrapper := &RecurringBillingRequest{billing}
	resp, err := c.client.Do(ctx, req, billingWrapper)
	if err != nil {
		return nil, resp, err
	}

	return billingWrapper.Billing, resp, nil
}

func (c *BillingService) Cancel(ctx context.Context, billing *Billing) (*Billing, *http.Response, error) {
	apiPath, found := billingApiPath[billing.Type]
	if !found {
		return nil, nil, ErrUnsupportedBillingType
	}

	req, err := c.client.NewRequest(
		"DELETE",
		fmt.Sprintf("/admin/%s/%d/activate.json", apiPath, billing.ID),
		&RecurringBillingRequest{billing},
	)
	if err != nil {
		return nil, nil, err
	}

	resp, err := c.client.Do(ctx, req, nil)
	if err != nil {
		return nil, resp, err
	}

	return nil, resp, nil
}

func (c *BillingService) Update(ctx context.Context, billing *Billing) (*Billing, *http.Response, error) {
	apiPath, found := billingApiPath[billing.Type]
	if !found {
		return nil, nil, ErrUnsupportedBillingType
	}

	req, err := c.client.NewRequest(
		"PUT",
		fmt.Sprintf("/admin/%s/%d.json", apiPath, billing.ID),
		&RecurringBillingRequest{billing},
	)
	if err != nil {
		return nil, nil, err
	}

	resp, err := c.client.Do(ctx, req, nil)
	if err != nil {
		return nil, resp, err
	}

	return nil, resp, nil
}

// String returns the string value of the status.
func (t BillingType) String() string {
	return billingTypes[t]
}

// MarshalText satisfies TextMarshaler
func (t BillingType) MarshalText() ([]byte, error) {
	return []byte(t.String()), nil
}

// UnmarshalText satisfies TextUnmarshaler
func (t *BillingType) UnmarshalText(text []byte) error {
	enum := string(text)
	for i := 0; i < len(billingTypes); i++ {
		if enum == billingTypes[i] {
			*t = BillingType(i)
			return nil
		}
	}
	return fmt.Errorf("unknown billing type %s", enum)
}

// String returns the string value of the status.
func (s BillingStatus) String() string {
	return billingStatuses[s]
}

// MarshalText satisfies TextMarshaler
func (s BillingStatus) MarshalText() ([]byte, error) {
	return []byte(s.String()), nil
}

// UnmarshalText satisfies TextUnmarshaler
func (s *BillingStatus) UnmarshalText(text []byte) error {
	enum := string(text)
	for i := 0; i < len(billingStatuses); i++ {
		if enum == billingStatuses[i] {
			*s = BillingStatus(i)
			return nil
		}
	}
	return fmt.Errorf("unknown billing status %s", enum)
}
