package scheduler

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/web/api"
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
	name           string
	spec           string
	fn             func()
	runImmediately bool
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
			name:           "job_schedule_deals",
			spec:           fmt.Sprintf("@every %s", duration),
			fn:             h.ScheduleDeals,
			runImmediately: true,
		},
		{
			name:           "job_get_merchant_deals",
			spec:           fmt.Sprintf("@every %s", 1*time.Hour),
			fn:             h.SyncDiscountCodes,
			runImmediately: true,
		},
		{
			name:           "abandoned_cart",
			spec:           "@every 4h",
			fn:             h.AbandonCartHandler,
			runImmediately: true,
		},
		{
			name:           "favourite_product",
			spec:           "@every 4h",
			fn:             h.FavouriteProductHandler,
			runImmediately: true,
		},
	}

	for _, s := range h.jobs {
		h.cron.AddJob(s.spec, s)
		if s.runImmediately {
			s.fn()
		}
	}
	h.cron.Start()

	for _, e := range h.cron.Entries() {
		f := e.Job.(FuncJob)
		lg.Infof("job: %s scheduled next run: %s", f.name, e.Schedule.Next(time.Now()))
	}
}

func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Use(NewStructuredLogger())

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`⏰`))
	})

	return r
}

func NewStructuredLogger() func(next http.Handler) http.Handler {
	return api.RequestLogger(lg.DefaultLogger)
}
