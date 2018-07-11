package reporter

import (
	"net/http"

	"bitbucket.org/moodie-app/moodie-api/lib/connect"
	"bitbucket.org/moodie-app/moodie-api/lib/events"
	"bitbucket.org/moodie-app/moodie-api/lib/forgett"
	"github.com/go-chi/chi"
)

type Handler struct {
	nats *connect.Nats

	trend *forgett.Distribution
}

func New(nats *connect.Nats) *Handler {
	// TODO: make this configurable
	trend, _ := forgett.NewDistribution(
		"product:trend",
		forgett.DefaultOptions.Lifetime,
		forgett.DefaultOptions.Norm,
	)

	return &Handler{
		nats:  nats,
		trend: trend,
	}
}

func (h *Handler) Subscribe(config connect.NatsConfig) {
	h.nats.Subscribe(events.EvProductViewed, h.HandleProductViewed)
	h.nats.Subscribe(events.EvProductPurchased, h.HandleProductPurchased)
}

func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`📺`))
	})

	r.Get("/trend", h.GetTrending)

	return r
}
