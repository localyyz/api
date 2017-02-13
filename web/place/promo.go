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

	var promos []*data.Promo
	err := data.DB.Promo.Find(
		db.And(
			db.Cond{"place_id": place.ID},
			db.Raw("start_at <= NOW() AT TIME ZONE 'UTC'"),
			db.Raw("end_at > NOW() AT TIME ZONE 'UTC'"),
		),
	).All(&promos)
	if err != nil {
		ws.Respond(w, http.StatusInternalServerError, err)
		return
	}

	res := make([]*presenter.Promo, len(promos))
	for i, p := range promos {
		res[i] = presenter.NewPromo(ctx, p).WithClaim()
		res[i].Place = presenter.NewPlace(ctx, place)
	}

	ws.Respond(w, http.StatusOK, res)
}
