package syncer

import (
	"net/http"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/syncer/shopify"
	"bitbucket.org/moodie-app/moodie-api/web/api"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/pressly/lg"
	db "upper.io/db.v3"
)

type Handler struct {
	DB    *data.Database
	Debug bool
}

func New(DB *data.Database, Debug bool) *Handler {
	if places, _ := DB.Place.FindAll(db.Cond{"status": data.PlaceStatusActive}); places != nil {
		shopify.SetupShopCache(places...)
	}
	if categories, _ := DB.Category.FindAll(nil); categories != nil {
		shopify.SetupCategoryCache(categories...)
	}
	if blacklist, _ := DB.Blacklist.FindAll(nil); blacklist != nil {
		shopify.SetupCategoryBlacklistCache(blacklist...)
	}
	return &Handler{DB, Debug}
}

func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()

	r.Use(middleware.RealIP)
	r.Use(middleware.NoCache)
	r.Use(middleware.RequestID)
	if h.Debug {
		r.Use(middleware.Logger)
		r.Use(lg.PrintPanics)
	} else {
		r.Use(NewStructuredLogger())
	}

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`😎`))
	})
	r.Mount("/shopify", shopify.Routes())
	r.Mount("/webhooks/shopify", shopify.Routes())

	return r
}

func NewStructuredLogger() func(next http.Handler) http.Handler {
	return api.RequestLogger(lg.DefaultLogger)
}
