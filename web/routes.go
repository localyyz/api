package web

import (
	"net/http"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/lib/connect"
	"bitbucket.org/moodie-app/moodie-api/lib/pusher"
	"bitbucket.org/moodie-app/moodie-api/lib/token"
	"bitbucket.org/moodie-app/moodie-api/web/api"
	"bitbucket.org/moodie-app/moodie-api/web/auth"
	"bitbucket.org/moodie-app/moodie-api/web/locale"
	"bitbucket.org/moodie-app/moodie-api/web/place"
	"bitbucket.org/moodie-app/moodie-api/web/product"
	"bitbucket.org/moodie-app/moodie-api/web/promo"
	"bitbucket.org/moodie-app/moodie-api/web/search"
	"bitbucket.org/moodie-app/moodie-api/web/session"
	"bitbucket.org/moodie-app/moodie-api/web/shopify"
	"bitbucket.org/moodie-app/moodie-api/web/user"

	"github.com/pressly/chi"
	"github.com/pressly/chi/middleware"
	"github.com/pressly/chi/render"
)

type Handler struct {
	DB    *data.Database
	Debug bool
}

func New(h *Handler) chi.Router {
	r := chi.NewRouter()

	r.Use(middleware.NoCache)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			if h.Debug {
				w.Header().Set("X-Internal-Debug", "1")
			}
			next.ServeHTTP(w, r)
		}
		return http.HandlerFunc(fn)
	})

	// Public Routes
	r.Group(func(r chi.Router) {
		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`( ͡° ͜ʖ ͡°)`))
		})

		r.Post("/login", auth.EmailLogin)
		r.Post("/login/facebook", auth.FacebookLogin)
		r.Post("/signup", auth.EmailSignup)

		r.Get("/connect/:shopID", shopify.Connect)
		r.Get("/oauth/shopify/callback", connect.SH.OAuthCb)
		r.Post("/webhooks/shopify", shopify.WebhookHandler)

		r.Post("/echo", echoPush)
	})

	// Authed Routes
	r.Group(func(r chi.Router) {
		r.Use(token.Verify())
		r.Use(session.SessionCtx)

		r.Mount("/session", session.Routes())
		r.Mount("/users", user.Routes())
		r.Mount("/places", place.Routes())
		r.Mount("/promos", promo.Routes())
		r.Mount("/locales", locale.Routes())
		r.Mount("/products", product.Routes())
		r.Mount("/search", search.Routes())
	})

	return r
}

type pushRequest struct {
	DeviceToken string `json:"deviceToken,required"`
	Payload     string `json:"payload"`
}

func (*pushRequest) Bind(r *http.Request) error {
	return nil
}

// test function: echo push to apns
func echoPush(w http.ResponseWriter, r *http.Request) {
	payload := &pushRequest{}
	if err := render.Bind(r, payload); err != nil {
		render.Render(w, r, api.ErrInvalidRequest(err))
		return
	}

	b := []byte(payload.Payload)
	t := payload.DeviceToken
	if err := pusher.Push(t, b); err != nil {
		render.Respond(w, r, err)
	}
}
