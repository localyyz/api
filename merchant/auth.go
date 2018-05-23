package merchant

import (
	"net/http"
	"strconv"
	"time"

	"bitbucket.org/moodie-app/moodie-api/lib/connect"
	"github.com/go-chi/render"
	"github.com/pressly/lg"
)

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
