package deals

import (
	"net/http"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/data/presenter"
	"github.com/go-chi/render"
	db "upper.io/db.v3"
)

const DealStatusCtxKey = "deal.status"

func ListDeal(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	status := ctx.Value(DealStatusCtxKey).(data.DealStatus)

	var dealCond interface{}
	if user, ok := ctx.Value("session.user").(*data.User); ok {
		dealCond = db.Or(
			db.Cond{"status": status},
			db.Cond{"status": status, "user_id": user.ID},
		)
	} else {
		dealCond = db.Cond{"status": status}
	}
	var orderBy string
	switch status {
	case data.DealStatusActive:
		// active: order by ending the "SOONEST" first
		orderBy = "end_at ASC"
	case data.DealStatusInactive:
		// inactive: order by ended "LAST" first
		orderBy = "end_at DESC"
	case data.DealStatusQueued:
		// queued: order by starting "SOONEST" first
		orderBy = "start_at ASC"
	}

	var deals []*data.Deal
	err := data.DB.Deal.Find(dealCond).OrderBy(orderBy).All(&deals)
	if err != nil {
		render.Respond(w, r, err)
		return
	}

	presented := presenter.NewDealList(ctx, deals)
	if err := render.RenderList(w, r, presented); err != nil {
		render.Respond(w, r, err)
	}
}

func GetDeal(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	deal := ctx.Value("deal").(*data.Deal)
	presented := presenter.NewDeal(ctx, deal)
	if err := render.Render(w, r, presented); err != nil {
		render.Respond(w, r, err)
	}
}
