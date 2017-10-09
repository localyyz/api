package web

import (
	"net/http"

	db "upper.io/db.v3"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/lib/connect"
	"bitbucket.org/moodie-app/moodie-api/lib/pusher"
	"bitbucket.org/moodie-app/moodie-api/lib/token"
	"bitbucket.org/moodie-app/moodie-api/web/api"
	"bitbucket.org/moodie-app/moodie-api/web/auth"
	"bitbucket.org/moodie-app/moodie-api/web/cart"
	"bitbucket.org/moodie-app/moodie-api/web/locale"
	"bitbucket.org/moodie-app/moodie-api/web/place"
	"bitbucket.org/moodie-app/moodie-api/web/product"
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
		r.Get("/signup", auth.GetSignupPage)
		r.Post("/signup", auth.EmailSignup)
		r.Post("/register", auth.RegisterSignup)
		r.Get("/leaderboard", leaderBoard)

		r.Get("/connect", shopify.Connect)
		r.Get("/oauth/shopify/callback", connect.SH.OAuthCb)
		r.With(shopify.ShopifyStoreWhCtx).
			Post("/webhooks/shopify", shopify.WebhookHandler)

		r.Post("/echo", echoPush)
	})

	// Authed Routes
	r.Group(func(r chi.Router) {
		r.Use(token.Verify())
		r.Use(session.SessionCtx)
		r.Use(session.UserRefresh)

		r.Mount("/session", session.Routes())
		r.Mount("/users", user.Routes())
		r.Mount("/places", place.Routes())
		r.Mount("/locales", locale.Routes())
		r.Mount("/products", product.Routes())
		r.Mount("/carts", cart.Routes())
		r.Post("/search", search.OmniSearch)
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

func leaderBoard(w http.ResponseWriter, r *http.Request) {
	query := data.DB.Select(
		db.Raw("v.etc->>'firstName' as name"),
		"v.avatar_url",
		db.Raw("count(*) count")).
		From("users u").
		LeftJoin("users v").
		On("v.id = (u.etc->>'invitedBy')::bigint").
		Where(
			db.And(
				db.Raw("u.etc->>'invitedBy' is not null"),
				db.Cond{"v.id !=": 0},
			),
		).
		GroupBy("v.id").
		OrderBy("count desc").
		Limit(10)

	type leader struct {
		FirstName string `db:"name" json:"name"`
		AvatarURL string `db:"avatar_url" json:"avatarUrl"`
		Count     int64  `db:"count" json:"count"`
	}
	var leaders []*leader
	if err := query.All(&leaders); err != nil {
		render.Respond(w, r, err)
		return
	}

	render.Respond(w, r, leaders)
}
