package scheduler

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"bitbucket.org/moodie-app/moodie-api/data"
	"github.com/go-chi/chi"
	"github.com/pressly/lg"
	"github.com/robfig/cron"
)

type Handler struct {
	DB          *data.Database
	Environment string

	jobs []FuncJob
	wg   sync.WaitGroup
	cron *cron.Cron
}

type FuncJob struct {
	name string
	spec string
	fn   func()
}

func (f FuncJob) Run() {
	f.fn()
}

func New(db *data.Database) *Handler {
	var wg sync.WaitGroup
	return &Handler{
		DB:   db,
		wg:   wg,
		cron: cron.New(),
	}
}

func (h *Handler) Wait() {
	// wait for jobs to complete
	h.wg.Wait()
}

func (h *Handler) Start() {
	duration := time.Second
	if h.Environment != "production" {
		duration = 1 * time.Hour
	}

	h.jobs = []FuncJob{
		{
			name: "job_schedule_deals",
			spec: fmt.Sprintf("@every %s", duration),
			fn:   h.ScheduleDeals,
		},
		{
			name: "job_sync_deals",
			spec: "@midnight",
			fn:   h.SyncDeals,
		},
	}

	for _, s := range h.jobs {
		h.cron.AddJob(s.spec, s)
	}
	h.cron.Start()

	for _, e := range h.cron.Entries() {
		f := e.Job.(FuncJob)
		lg.Infof("job: %s scheduled next run: %s", f.name, e.Next)
	}
}

func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`⏰`))
	})

	return r
}
