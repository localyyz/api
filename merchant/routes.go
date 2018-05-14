package merchant

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	db "upper.io/db.v3"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/lib/connect"
	"bitbucket.org/moodie-app/moodie-api/lib/shopify"
	"bitbucket.org/moodie-app/moodie-api/lib/slack"
	"bitbucket.org/moodie-app/moodie-api/lib/token"
	"bitbucket.org/moodie-app/moodie-api/merchant/approval"
	"bitbucket.org/moodie-app/moodie-api/merchant/plan"
	"github.com/flosch/pongo2"
	_ "github.com/flosch/pongo2-addons"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/jwtauth"
	"github.com/go-chi/render"
	"github.com/pressly/lg"
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
		r.Post("/tos", AcceptTOS)
		r.Mount("/plan", plan.Routes())
	})

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
	if place.Status == data.PlaceStatusWaitApproval {
		pageContext["approvalTs"] = place.TOSAgreedAt.Add(1 * 24 * time.Hour)
	}
	pageContext["productCount"], _ = data.DB.Product.Find(db.Cond{"place_id": place.ID}).Count()

	if strings.HasPrefix(r.UserAgent(), "Shopify Mobile") {
		pageContext["isMobile"] = true
	}

	if place.PlanEnabled {
		// if we have a pending plan, fetch and redirect
		pendingBilling, _ := data.DB.PlaceBilling.FindOne(
			db.Cond{
				"place_id": place.ID,
				"status":   data.BillingStatusPending,
			},
		)
		if pendingBilling != nil {
			// fetch the remote updated billing status
			client := ctx.Value("shopify.client").(*shopify.Client)
			shopBilling, _, _ := client.Billing.Get(ctx, &shopify.Billing{ID: pendingBilling.ExternalID})
			if shopBilling != nil {
				if shopBilling.Status == shopify.BillingStatusPending {
					pageContext["confirmationUrl"] = shopBilling.ConfirmationUrl
					pageContext["shouldRedirect"] = 1
				}
			}
		}

		// find active billing
		// if we have a pending plan, fetch and redirect
		activeBilling, _ := data.DB.PlaceBilling.FindOne(
			db.Cond{
				"place_id": place.ID,
				"status":   data.BillingStatusActive,
			},
		)
		if activeBilling != nil {
			type planWrapper struct {
				Type      string
				Status    string
				StartedOn string
				ExpiresOn string
			}
			plan, _ := data.DB.BillingPlan.FindByID(activeBilling.PlanID)
			pageContext["plan"] = planWrapper{
				Type:      plan.PlanType.String(),
				Status:    activeBilling.Status.String(),
				StartedOn: activeBilling.AcceptedAt.Format("January 2, 2006"),
			}
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

func AcceptTOS(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	place := ctx.Value("place").(*data.Place)

	// must be in "Wait Agreement" status to accept
	if place.Status != data.PlaceStatusWaitAgreement {
		return
	}

	// proceed to next status
	place.Status += 1
	place.TOSAgreedAt = data.GetTimeUTCPointer()
	place.TOSIP = r.RemoteAddr
	if err := data.DB.Place.Save(place); err != nil {
		// error has occured. respond
		render.Status(r, http.StatusInternalServerError)
		render.Respond(w, r, err)
		return
	}

	// notify slack with button for approving the merchant
	sl := ctx.Value("slack.client").(*connect.Slack)
	sl.Notify(
		"store",
		fmt.Sprintf("<%s|%s> (id: %v) just accepted the TOS!", place.Website, place.Name, place.ID),
		&slack.Attachment{
			Title:      "start review process:",
			TitleLink:  fmt.Sprintf("https://merchant.localyyz.com/approval/%d", place.ID),
			Fallback:   "You are unable to approve / reject the store.",
			CallbackID: fmt.Sprintf("placeid%d", place.ID),
			Color:      "0195ff",
		},
	)

	render.Respond(w, r, "success")
	return
}

func SessionCtx(next http.Handler) http.Handler {
	handler := func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		token, claims, err := jwtauth.FromContext(ctx)
		if token == nil || err != nil {
			lg.Infof("invalid merchant ctx token is nil (%s): %+v", r.URL.Path, err)
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		rawPlaceID, ok := claims["place_id"].(json.Number)
		if !ok {
			lg.Error("invalid session merchant, no place_id found")
			return
		}

		placeID, err := rawPlaceID.Int64()
		if err != nil {
			lg.Errorf("invalid session merchant: %+v", err)
			return
		}

		// find a logged in user with the given id
		place, err := data.DB.Place.FindOne(
			db.Cond{"id": placeID},
		)
		if err != nil {
			lg.Errorf("invalid session merchant: %+v", err)
			return
		}

		ctx = context.WithValue(ctx, "place", place)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
	return http.HandlerFunc(handler)
}

func VerifySignature(next http.Handler) http.Handler {
	handler := func(w http.ResponseWriter, r *http.Request) {
		// verify the signature
		ctx := r.Context()
		sh := ctx.Value("shopify.client").(*connect.Shopify)

		q := r.URL.Query()
		sig := []byte(q.Get("hmac"))
		if len(sig) == 0 {
			lg.Warn("verify: missing hmac")
			w.WriteHeader(http.StatusUnauthorized)
			render.Respond(w, r, "unauthorized")
			return
		}

		// verify timestamp
		ts, err := strconv.ParseInt(q.Get("timestamp"), 10, 64)
		if err != nil {
			lg.Warn("verify: missing timestamp")
			w.WriteHeader(http.StatusUnauthorized)
			render.Respond(w, r, "unauthorized")
			return
		}
		tm := time.Unix(ts, 0)
		if time.Now().Before(tm) {
			// is tm in the future?
			lg.Warn("verify: timestamp before current time")
			w.WriteHeader(http.StatusUnauthorized)
			render.Respond(w, r, "unauthorized")
			return
		}
		//if time.Since(tm) > SignatureTimeout {
		//// is tm outside of the timeout (30s)
		//lg.Warn("verify: timestamp timed out (30s)")
		//render.Respond(w, r, "unauthorized")
		//return
		//}

		// remove the hmac key
		q.Del("hmac")

		if !sh.VerifySignature(sig, q.Encode()) {
			lg.Warn("verify: hmac mismatch")
			w.WriteHeader(http.StatusUnauthorized)
			render.Respond(w, r, "unauthorized")
			return
		}
		next.ServeHTTP(w, r.WithContext(ctx))
	}
	return http.HandlerFunc(handler)
}
