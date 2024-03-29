package web

import (
	"net/http"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/lib/connect"
	"bitbucket.org/moodie-app/moodie-api/lib/token"
	"bitbucket.org/moodie-app/moodie-api/web/api"
	"bitbucket.org/moodie-app/moodie-api/web/auth"
	"bitbucket.org/moodie-app/moodie-api/web/cart"
	"bitbucket.org/moodie-app/moodie-api/web/category"
	"bitbucket.org/moodie-app/moodie-api/web/collection"
	"bitbucket.org/moodie-app/moodie-api/web/deals"
	"bitbucket.org/moodie-app/moodie-api/web/ping"
	"bitbucket.org/moodie-app/moodie-api/web/place"
	"bitbucket.org/moodie-app/moodie-api/web/product"
	"bitbucket.org/moodie-app/moodie-api/web/search"
	"bitbucket.org/moodie-app/moodie-api/web/session"
	"bitbucket.org/moodie-app/moodie-api/web/shopify"
	"bitbucket.org/moodie-app/moodie-api/web/user"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/pressly/lg"
)

type Handler struct {
	DB     *data.Database
	Debug  bool
	Silent bool
}

func New(DB *data.Database) *Handler {
	if DB == nil {
		return &Handler{}
	}
	return &Handler{DB: DB}
}

func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()

	r.Use(middleware.RealIP)
	r.Use(middleware.NoCache)
	r.Use(middleware.RequestID)

	if !h.Silent {
		if h.Debug {
			r.Use(middleware.Logger)
		} else {
			r.Use(NewStructuredLogger())
		}
	}
	r.Use(func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			if h.Debug {
				w.Header().Set("X-Internal-Debug", "1")
			}
			next.ServeHTTP(w, r)
		}
		return http.HandlerFunc(fn)
	})

	r.Use(token.Verify())
	r.Use(session.SessionCtx)
	r.Use(session.UserRefresh)
	r.Use(session.DeviceCtx)
	r.Use(api.PaginateCtx)

	if h.Debug {
		r.Use(lg.PrintPanics)
	}

	// Public Routes
	r.Group(func(r chi.Router) {
		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`( ͡° ͜ʖ ͡°)`))
		})

		// sign-up and login related
		r.Post("/login", auth.EmailLogin)
		r.Post("/login/facebook", auth.FacebookLogin)
		r.Post("/signup", auth.EmailSignup)

		// shopify related endpoints
		r.Get("/connect", shopify.Connect)
		r.Get("/oauth/shopify/callback", connect.SH.OAuthCb)

		// public api routes
		r.Mount("/search", search.Routes())
		r.Mount("/collections", collection.Routes())
		r.Mount("/categories", category.Routes())
		r.Mount("/places", place.Routes())
		r.Mount("/products", product.Routes())
		r.Mount("/deals", deals.Routes())
		r.Post("/ping", ping.LogDeviceData)
	})

	// Semi-authed route. User can be "shadow"
	r.Group(func(r chi.Router) {
		r.Use(auth.DeviceCtx)
		r.Mount("/carts", cart.Routes())
		r.Mount("/users", user.Routes())
	})

	// Authed Routes
	r.Group(func(r chi.Router) {
		r.Use(auth.SessionCtx)
		r.Mount("/session", session.Routes())

		r.Put("/places/internal", place.UpdateInternal)
	})

	return r
}

func NewStructuredLogger() func(next http.Handler) http.Handler {
	return api.RequestLogger(lg.DefaultLogger)
}
