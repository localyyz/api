package presenter

import (
	"context"
	"net/http"
	"time"

	"bitbucket.org/moodie-app/moodie-api/data"
	"github.com/go-chi/render"
	"upper.io/db.v3"
)

type Place struct {
	*data.Place
	ProductCount uint64            `json:"productCount"`
	Products     []render.Renderer `json:"products"`
	IsFavourite  bool              `json:"isFavourite"`

	IsFeatured bool   `json:"isFeatured"`
	Currency   string `json:"currency"`

	Billing     interface{} `json:"billing,omitempty"`
	TOSAgreedAt interface{} `json:"tosAgreedAt,omitempty"`
	ApprovedAt  interface{} `json:"approvedAt,omitempty"`

	ctx context.Context
}

func NewPlace(ctx context.Context, place *data.Place) *Place {
	p := &Place{
		Place:    place,
		Currency: place.Currency,
		ctx:      ctx,
	}
	p.ProductCount, _ = data.DB.Product.Find(db.Cond{
		"place_id": p.ID,
		"status":   data.ProductStatusApproved,
	}).Count()
	if user, _ := ctx.Value("session.user").(*data.User); user != nil {
		p.IsFavourite, _ = data.DB.FavouritePlace.Find(db.Cond{"place_id": p.ID, "user_id": user.ID}).Exists()
	}

	if withPreview, ok := ctx.Value("with.preview").(bool); withPreview && ok {
		cond := db.Cond{"place_id": p.ID}
		if gender, ok := ctx.Value("session.gender").(data.UserGender); ok {
			cond["gender"] = gender
		}
		var products []*data.Product
		data.DB.Product.Find(cond).Limit(4).All(&products)
		p.Products = NewProductList(ctx, products)
	}

	return p
}

func NewPlaceList(ctx context.Context, places []*data.Place) []render.Renderer {
	list := []render.Renderer{}
	for _, place := range places {
		p := NewPlace(ctx, place)
		if p.ProductCount == 0 {
			continue
		}
		list = append(list, p)
	}
	return list
}

// Place implements render.Renderer interface
func (pl *Place) Render(w http.ResponseWriter, r *http.Request) error {
	if pl.Weight >= data.PlaceFeatureWeightCutoff {
		pl.IsFeatured = true
	}
	return nil
}

// PlaceApproval is used to combine useful fields
// to sent to Zapier and then synced to a google spread sheet
type PlaceApproval struct {
	ID        int64            `json:"id"`
	Name      string           `json:"name"`
	Status    data.PlaceStatus `json:"status"`
	Plan      string           `json:"shopify_plan"`
	Currency  string           `json:"currency"`
	Website   string           `json:"website"`
	CreatedAt time.Time        `json:"createdAt"`

	PlanType      data.BillingPlanType `json:"subscription"`
	BillingStatus data.BillingStatus   `json:"billing_status"`

	Gender      *data.Gender     `json:"gender"`
	StyleFemale *data.PlaceStyle `json:"style_female"`
	StyleMale   *data.PlaceStyle `json:"style_male"`
	Pricing     string           `json:"pricing"`
}

func NewPlaceApproval(place *data.Place) *PlaceApproval {
	presented := &PlaceApproval{
		ID:        place.ID,
		Name:      place.Name,
		Status:    place.Status,
		Plan:      place.Plan,
		Currency:  place.Currency,
		Website:   place.Website,
		CreatedAt: *place.CreatedAt,
	}

	if b, _ := data.DB.PlaceBilling.FindByPlaceID(place.ID); b != nil {
		if plan, _ := data.DB.BillingPlan.FindByID(b.PlanID); plan != nil {
			presented.PlanType = plan.PlanType
			presented.BillingStatus = b.Status
		}
	}

	if pm, _ := data.DB.PlaceMeta.FindByPlaceID(place.ID); pm != nil {
		presented.Gender = pm.Gender
		presented.StyleMale = pm.StyleMale
		presented.StyleFemale = pm.StyleFemale
		presented.Pricing = pm.Pricing
	}

	return presented
}
