package place

import (
	"net/http"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/data/presenter"
	"bitbucket.org/moodie-app/moodie-api/lib/ws"
	db "upper.io/db.v2"
)

func ListPromo(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	place := ctx.Value("place").(*data.Place)

	promos, err := data.DB.Promo.FindAll(db.Cond{"place_id": place.ID})
	if err != nil {
		ws.Respond(w, http.StatusInternalServerError, err)
		return
	}

	res := make([]*presenter.Promo, len(promos))
	for i, p := range promos {
		res[i] = presenter.NewPromo(ctx, p).WithClaim()
		res[i].Place = place
	}

	ws.Respond(w, http.StatusOK, promos)
}
