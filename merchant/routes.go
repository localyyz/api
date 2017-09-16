package merchant

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"bitbucket.org/moodie-app/moodie-api/data"
	"github.com/pressly/chi"
	"github.com/pressly/chi/middleware"
	"github.com/pressly/chi/render"
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

	r.Use(middleware.WithValue(ShopifySecretKey, h.ShopifySecret))
	r.Use(VerifySignature)

	r.Get("/", Index)

	return r
}

func Index(w http.ResponseWriter, r *http.Request) {
	b, _ := ioutil.ReadFile("./merchant/index.html")
	render.HTML(w, r, string(b))
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
