package web

import (
	"net/http"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/lib/connect"
	"bitbucket.org/moodie-app/moodie-api/lib/pusher"
	"bitbucket.org/moodie-app/moodie-api/lib/token"
	"bitbucket.org/moodie-app/moodie-api/web/api"
	"bitbucket.org/moodie-app/moodie-api/web/auth"
	"bitbucket.org/moodie-app/moodie-api/web/cart"
	"bitbucket.org/moodie-app/moodie-api/web/cart/express"
	"bitbucket.org/moodie-app/moodie-api/web/category"
	"bitbucket.org/moodie-app/moodie-api/web/collection"
	"bitbucket.org/moodie-app/moodie-api/web/designer"
	"bitbucket.org/moodie-app/moodie-api/web/ping"
	"bitbucket.org/moodie-app/moodie-api/web/place"
	"bitbucket.org/moodie-app/moodie-api/web/product"
	"bitbucket.org/moodie-app/moodie-api/web/search"
	"bitbucket.org/moodie-app/moodie-api/web/session"
	"bitbucket.org/moodie-app/moodie-api/web/shopify"
	"bitbucket.org/moodie-app/moodie-api/web/user"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
	"github.com/pressly/lg"
)

type Handler struct {
	DB    *data.Database
	Debug bool
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
	if h.Debug {
		r.Use(middleware.Logger)
	} else {
		r.Use(NewStructuredLogger())
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

		// push notification
		r.Post("/echo", echoPush)

		// public api routes
		r.Mount("/search", search.Routes())
		r.Mount("/collections", collection.Routes())
		r.Mount("/categories", category.Routes())
		r.Mount("/designers", designer.Routes())
		r.Mount("/places", place.Routes())
		r.Mount("/products", product.Routes())
		r.Mount("/ping", ping.Routes())
	})

	// Semi-authed route. User can be "shadow"
	r.Group(func(r chi.Router) {
		r.Use(auth.DeviceCtx)
		r.Mount("/carts/express", express.Routes())
	})

	// Authed Routes
	r.Group(func(r chi.Router) {
		r.Use(auth.SessionCtx)

		r.Mount("/session", session.Routes())
		r.Mount("/users", user.Routes())
		r.Mount("/carts", cart.Routes())
	})

	return r
}

func NewStructuredLogger() func(next http.Handler) http.Handler {
	return api.RequestLogger(lg.DefaultLogger)
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
