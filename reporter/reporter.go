package reporter

import (
	"net/http"

	"bitbucket.org/moodie-app/moodie-api/lib/connect"
	"bitbucket.org/moodie-app/moodie-api/lib/events"
	"github.com/go-chi/chi"
)

type Handler struct {
	nats *connect.Nats
}

func New(nats *connect.Nats) *Handler {
	return &Handler{
		nats: nats,
	}
}

func (h *Handler) Subscribe(config connect.NatsConfig) {
	h.nats.Subscribe(events.EvProductViewed, HandleProductViewed)
	h.nats.Subscribe(events.EvProductPurchased, HandleProductPurchased)
}

func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`📺`))
	})

	return r
}
