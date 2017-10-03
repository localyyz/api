package merchant

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	db "upper.io/db.v3"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/lib/token"
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/flosch/pongo2"
	_ "github.com/flosch/pongo2-addons"
	"github.com/goware/jwtauth"
	"github.com/pressly/chi"
	"github.com/pressly/chi/middleware"
	"github.com/pressly/chi/render"
	"github.com/pressly/lg"
)

type Handler struct {
	DB            *data.Database
	Debug         bool
	ShopifySecret string
}

const (
	ShopifySecretKey = "shopify.secret"
	SignatureTimeout = 30 * time.Second
)

func New(h *Handler) chi.Router {
	r := chi.NewRouter()

	r.Use(middleware.NoCache)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Shopify auth routes
	r.Group(func(r chi.Router) {
		if !h.Debug {
			r.Use(middleware.WithValue(ShopifySecretKey, h.ShopifySecret))
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
	})

	return r
}

func Index(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	place := ctx.Value("place").(*data.Place)

	pageContext := pongo2.Context{
		"place":  place,
		"name":   strings.Replace(url.QueryEscape(place.Name), "+", "%20", -1),
		"status": place.Status.String(),
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

func ShopifyShopCtx(next http.Handler) http.Handler {
	handler := func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		q := r.URL.Query()
		shopDomain := q.Get("shop")

		if len(shopDomain) == 0 {
			render.Status(r, http.StatusNotFound)
			render.Respond(w, r, "")
			return
		}

		// TODO: Use a tld lib
		parts := strings.Split(shopDomain, ".")
		shopID := parts[0]

		place, err := data.DB.Place.FindByShopifyID(shopID)
		if err != nil {
			render.Respond(w, r, err)
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
		secret := ctx.Value(ShopifySecretKey).(string)

		q := r.URL.Query()
		sig := []byte(q.Get("hmac"))
		if len(sig) == 0 {
			render.Respond(w, r, ErrUnauthorized)
			return
		}

		// verify timestamp
		ts, err := strconv.ParseInt(q.Get("timestamp"), 10, 64)
		if err != nil {
			render.Respond(w, r, ErrUnauthorized)
			return
		}
		tm := time.Unix(ts, 0)
		if time.Now().Before(tm) {
			// is tm in the future?
			render.Respond(w, r, ErrUnauthorized)
			return
		}
		if time.Since(tm) > SignatureTimeout {
			// is tm outside of the timeout (30s)
			render.Respond(w, r, ErrUnauthorized)
			return
		}

		// remove the hmac key
		q.Del("hmac")

		mac := hmac.New(sha256.New, []byte(secret))
		// query unescape
		uu, _ := url.QueryUnescape(q.Encode())
		mac.Write([]byte(uu))

		src := mac.Sum(nil)
		// hex encode
		expectedSig := make([]byte, hex.EncodedLen(len(src)))
		hex.Encode(expectedSig, src)

		if !hmac.Equal(sig, expectedSig) {
			render.Respond(w, r, ErrUnauthorized)
			return
		}
		next.ServeHTTP(w, r.WithContext(ctx))
	}
	return http.HandlerFunc(handler)
}
