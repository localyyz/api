package merchant

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	db "upper.io/db.v3"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/lib/connect"
	"bitbucket.org/moodie-app/moodie-api/lib/token"
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/flosch/pongo2"
	_ "github.com/flosch/pongo2-addons"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
	"github.com/goware/jwtauth"
	"github.com/pressly/lg"
)

type Handler struct {
	DB     *data.Database
	SH     *connect.Shopify
	ApiURL string
	Debug  bool
}

const (
	SignatureTimeout = 30 * time.Second
)

func New(h *Handler) chi.Router {
	r := chi.NewRouter()

	r.Use(middleware.RealIP)
	r.Use(middleware.NoCache)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.WithValue("shopify.client", h.SH))
	r.Use(middleware.WithValue("api.url", h.ApiURL))

	// Shopify auth routes
	r.Group(func(r chi.Router) {
		if !h.Debug {
			r.Use(VerifySignature)
		}
		r.Use(ShopifyShopCtx)
		r.Use(ShopifyChargeCtx)
		r.Get("/", Index)
	})

	// Jwt auth routes
	r.Group(func(r chi.Router) {
		r.Use(token.Verify())
		r.Use(SessionCtx)
		r.Post("/tos", AcceptTOS)
	})

	return r
}

func Index(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	place := ctx.Value("place").(*data.Place)

	pageContext := pongo2.Context{
		"place":   place,
		"billing": place.Billing,
		"name":    strings.Replace(url.QueryEscape(place.Name), "+", "%20", -1),
		"status":  place.Status.String(),
	}
	if place.Status == data.PlaceStatusWaitApproval {
		pageContext["approvalTs"] = place.TOSAgreedAt.Add(1 * 24 * time.Hour)
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

	render.Respond(w, r, "success")
	return
}

func SessionCtx(next http.Handler) http.Handler {
	handler := func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		tok, _ := ctx.Value("jwt").(*jwt.Token)
		if tok == nil {
			return
		}

		token, err := token.Decode(tok.Raw)
		if err != nil {
			lg.Errorf("invalid session token: %+v", err)
			return
		}

		rawPlaceID, ok := token.Claims["place_id"].(json.Number)
		if !ok {
			lg.Error("invalid session token, no user_id found")
			return
		}

		placeID, err := rawPlaceID.Int64()
		if err != nil {
			lg.Errorf("invalid session token: %+v", err)
			return
		}

		// find a logged in user with the given id
		place, err := data.DB.Place.FindOne(
			db.Cond{"id": placeID},
		)
		if err != nil {
			lg.Errorf("invalid session user: %+v", err)
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
			render.Respond(w, r, ErrUnauthorized)
			return
		}

		// verify timestamp
		ts, err := strconv.ParseInt(q.Get("timestamp"), 10, 64)
		if err != nil {
			lg.Warn("verify: missing timestamp")
			render.Respond(w, r, ErrUnauthorized)
			return
		}
		tm := time.Unix(ts, 0)
		if time.Now().Before(tm) {
			// is tm in the future?
			lg.Warn("verify: timestamp before current time")
			render.Respond(w, r, ErrUnauthorized)
			return
		}
		if time.Since(tm) > SignatureTimeout {
			// is tm outside of the timeout (30s)
			lg.Warn("verify: timestamp timed out (30s)")
			render.Respond(w, r, ErrUnauthorized)
			return
		}

		// remove the hmac key
		q.Del("hmac")

		if !sh.VerifySignature(sig, q.Encode()) {
			lg.Warn("verify: hmac mismatch")
			render.Respond(w, r, ErrUnauthorized)
			return
		}
		next.ServeHTTP(w, r.WithContext(ctx))
	}
	return http.HandlerFunc(handler)
}