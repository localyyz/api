package reporter

import (
	"fmt"
	"net/http"
	"strconv"

	"bitbucket.org/moodie-app/moodie-api/data/presenter"
	"bitbucket.org/moodie-app/moodie-api/data/stash"
	"github.com/go-chi/render"
)

func (h *Handler) HandleProductViewed(product *presenter.ProductEvent) {
	stash.IncrProductViews(product.ID, product.ViewerID)
	h.trend.Incr(fmt.Sprintf("%d", product.ID))
}

func (h *Handler) HandleProductPurchased(product *presenter.ProductEvent) {
	stash.IncrProductPurchases(product.ID)
	h.trend.IncrBy(fmt.Sprintf("%d", product.ID), 2)
}

func (h *Handler) GetTrending(w http.ResponseWriter, r *http.Request) {
	result := &presenter.ProductTrend{
		// always initialize so it's not null
		IDs: []int64{},
	}
	scores, _ := h.trend.TopScores(50)
	for k, _ := range scores {
		ID, _ := strconv.ParseInt(k, 10, 64)
		result.IDs = append(result.IDs, ID)
	}
	render.Respond(w, r, result)
}
