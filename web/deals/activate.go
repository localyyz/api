package deals

import (
	"fmt"
	"net/http"
	"time"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/data/presenter"
	"bitbucket.org/moodie-app/moodie-api/lib/connect"
	"bitbucket.org/moodie-app/moodie-api/lib/shopify"
	"bitbucket.org/moodie-app/moodie-api/web/api"
	"github.com/go-chi/render"
	"github.com/pkg/errors"
)

type ActivateRequest struct {
	DealID int64 `json:"dealId,required"`
	//Token    string     `json:"token,required"`
	dealID   int64
	StartAt  *time.Time `json:"startAt,omitempty"`
	EndAt    *time.Time `json:"endAt,omitempty"`
	Duration int64      `json:"duration,omitempty"`
}

func (a *ActivateRequest) Bind(r *http.Request) error {
	if a.StartAt == nil {
		a.StartAt = data.GetTimeUTCPointer()
	}
	if a.Duration == 0 {
		a.Duration = 1
	}
	if a.EndAt == nil {
		duration := time.Duration(a.Duration) * time.Hour
		endAt := a.StartAt.Add(duration)
		a.EndAt = &endAt
	}
	// TODO... validate token
	return nil
}

func ActivateDeal(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var payload ActivateRequest
	if err := render.Bind(r, &payload); err != nil {
		render.Respond(w, r, api.ErrInvalidRequest(err))
		return
	}

	// look up the parent deal
	parentDeal, err := data.DB.Deal.FindByID(payload.DealID)
	if err != nil {
		render.Respond(w, r, errors.Wrap(err, "get parent deal"))
		return
	}

	// if parent deal is already active. skip it
	if parentDeal.Status == data.DealStatusActive {
		render.Respond(w, r, api.ErrDealActive)
		return
	}

	client, err := connect.GetShopifyClient(parentDeal.MerchantID)
	if err != nil {
		render.Respond(w, r, errors.Wrap(err, "shopify client"))
		return
	}

	user := ctx.Value("session.user").(*data.User)
	d := &shopify.DiscountCode{
		PriceRuleID: parentDeal.ExternalID,
		Code:        fmt.Sprintf("%s_USER_%s", parentDeal.Code, user.InviteCode),
	}
	if _, _, err := client.DiscountCode.Create(ctx, d); err != nil {
		render.Respond(w, r, api.ErrInvalidRequest(err))
		return
	}

	userDeal := &data.Deal{
		ExternalID: d.ID,
		UserID:     &(user.ID),
		ParentID:   &(parentDeal.ID),
		MerchantID: parentDeal.MerchantID,
		Status:     data.DealStatusActive,
		Code:       d.Code,
		StartAt:    payload.StartAt,
		EndAt:      payload.EndAt,
	}
	if err := data.DB.Deal.Save(userDeal); err != nil {
		// return that the deal has already expired
		render.Respond(w, r, errors.Wrap(err, "create deal"))
		return
	}

	render.Status(r, http.StatusCreated)
	render.Render(w, r, presenter.NewDeal(ctx, userDeal))
}
