package scheduler

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"bitbucket.org/moodie-app/moodie-api/data"
	"github.com/go-chi/chi"
	"github.com/robfig/cron"
)

type Handler struct {
	DB          *data.Database
	Environment string

	wg   sync.WaitGroup
	cron *cron.Cron
}

func New(db *data.Database) *Handler {
	return &Handler{
		DB:   db,
		cron: cron.New(),
	}
}

func (h *Handler) Start() {
	duration := time.Second
	if h.Environment != "production" {
		duration = 1 * time.Minute
	}

	h.cron.AddFunc(fmt.Sprintf("@every %s", duration), h.ScheduleDeals)
	h.cron.AddFunc(fmt.Sprintf("@every %s", duration), h.ScheduleWelcomeEmail)
	h.cron.Start()
}

func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`⏰`))
	})

	return r
}
