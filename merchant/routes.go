package merchant

import (
	"net/http"
	"net/url"
	"strings"
	"time"

	db "upper.io/db.v3"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/lib/connect"
	"bitbucket.org/moodie-app/moodie-api/lib/shopify"
	"bitbucket.org/moodie-app/moodie-api/lib/token"
	"bitbucket.org/moodie-app/moodie-api/merchant/approval"
	"github.com/flosch/pongo2"
	_ "github.com/flosch/pongo2-addons"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/jwtauth"
	"github.com/go-chi/render"
)

type Handler struct {
	DB          *data.Database
	SH          *connect.Shopify
	SL          *connect.Slack
	ApiURL      string
	Debug       bool
	Environment string
}

const (
	SignatureTimeout = 30 * time.Second
)

var (
	indexTmpl *pongo2.Template
)

func New(h *Handler) chi.Router {
	r := chi.NewRouter()

	if h.Environment == "development" {
		indexTmpl = pongo2.Must(pongo2.FromFile("./merchant/index.html"))
	} else {
		indexTmpl = pongo2.Must(pongo2.FromFile("/merchant/index.html"))
	}

	// initialize approval
	// TODO: move this to TOOL
	approval.Init(h.Environment)

	r.Use(middleware.RealIP)
	r.Use(middleware.NoCache)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.WithValue("shopify.client", h.SH))
	r.Use(middleware.WithValue("shopify.appid", h.SH.ClientID()))
	r.Use(middleware.WithValue("shopify.appname", h.SH.AppName()))
	r.Use(middleware.WithValue("slack.client", h.SL))
	r.Use(middleware.WithValue("api.url", h.ApiURL))
	r.Use(middleware.WithValue("debug", h.Debug))

	// Shopify auth routes
	r.Group(func(r chi.Router) {
		if !h.Debug {
			r.Use(VerifySignature)
		}
		r.Use(ShopifyShopCtx)
		r.Get("/", Index)
	})

	// Jwt auth routes
	r.Group(func(r chi.Router) {
		r.Use(token.Verify())
		r.Use(SessionCtx)
		r.Use(MustValidateSessionCtx)

		r.Post("/tos", AcceptTOS)
		r.Mount("/plan", planRoutes())
	})

	// TODO: move this to the TOOL
	r.Mount("/approval", approval.Routes())

	return r
}

func Index(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	place := ctx.Value("place").(*data.Place)
	clientID := ctx.Value("shopify.appid").(string)

	pageContext := pongo2.Context{
		"place":    place,
		"name":     strings.Replace(url.QueryEscape(place.Name), "+", "%20", -1),
		"status":   place.Status.String(),
		"clientID": clientID,
	}
	pageContext["productCount"], _ = data.DB.Product.Find(db.Cond{"place_id": place.ID}).Count()

	if strings.HasPrefix(r.UserAgent(), "Shopify Mobile") {
		pageContext["isMobile"] = true
	}

	// fetch the number of places waiting for approval
	pageContext["approvalWait"], _ = data.DB.Place.Find(db.Cond{"status": data.PlaceStatusWaitApproval}).Count()

	if place.PlanEnabled {
		billing, _ := data.DB.PlaceBilling.FindOne(
			db.Cond{
				"place_id": place.ID,
			},
		)
		if billing != nil {
			type planWrapper struct {
				Type      string
				Status    string
				StartedOn string
				ExpiresOn string
			}
			if billing.Status == data.BillingStatusPending {
				// fetch the remote updated billing status
				client := ctx.Value("shopify.client").(*shopify.Client)
				shopBilling, _, _ := client.Billing.Get(ctx, &shopify.Billing{ID: billing.ExternalID})
				if shopBilling != nil {
					if shopBilling.Status == shopify.BillingStatusPending {
						pageContext["confirmationUrl"] = shopBilling.ConfirmationUrl
						pageContext["shouldRedirect"] = 1
					}
				}
			}
			plan, _ := data.DB.BillingPlan.FindByID(billing.PlanID)
			w := planWrapper{
				Type:   plan.PlanType.String(),
				Status: billing.Status.String(),
			}
			if billing.AcceptedAt != nil {
				w.StartedOn = billing.AcceptedAt.Format("January 2, 2006")
			}
			pageContext["plan"] = w
		}
	}

	// inject a token into the cookie.
	token, _ := token.Encode(jwtauth.Claims{"place_id": place.ID})
	http.SetCookie(w, &http.Cookie{
		Name:     "jwt",
		Value:    token.Raw,
		HttpOnly: false,
	})
	t, _ := indexTmpl.Execute(pageContext)
	render.HTML(w, r, t)
}
