package deals

import (
	"net/http"

	"bitbucket.org/moodie-app/moodie-api/web/api"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/data/presenter"
	"github.com/go-chi/render"
	"upper.io/db.v3"
	"upper.io/db.v3/lib/sqlbuilder"
)

const DealStatusCtxKey = "deal.status"

func ListDeal(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	status := ctx.Value(DealStatusCtxKey).(data.DealStatus)

	var dealCond interface{}

	//if place, _ := ctx.Value("place").(data.Place); place.ID != 4164 {
	//	return
	//}

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

func ListUpcomingDeal(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	cursor := ctx.Value("cursor").(*api.Page)

	orderBy := "start_at ASC"

	dealCond := db.Cond{
		"status":   data.DealStatusQueued,
		"featured": false,
		"start_at": db.Lte(db.Raw("NOW()::date + 7")),
	}

	var upcomingDeals []*data.Deal
	query := data.DB.Select("d.*").
		From("deals d").
		Where(dealCond).
		OrderBy(orderBy)

	paginate := cursor.UpdateQueryBuilder(query)

	if err := paginate.All(&upcomingDeals); err != nil {
		render.Respond(w, r, err)
	}
	cursor.Update(upcomingDeals)

	presented := presenter.NewDealList(ctx, upcomingDeals)
	if err := render.RenderList(w, r, presented); err != nil {
		render.Respond(w, r, err)
	}
}

func ListOngoingDeal(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	cursor := ctx.Value("cursor").(*api.Page)

	orderBy := "start_at ASC"

	dealCond := db.Cond{
		"status":   data.DealStatusActive,
		"timed":    false,
		"featured": false,
	}

	var ongoingDeals []*data.Deal

	query := data.DB.Select("d.*").
		From("deals d").
		Where(dealCond).
		OrderBy(orderBy)

	paginate := cursor.UpdateQueryBuilder(query)

	if err := paginate.All(&ongoingDeals); err != nil {
		render.Respond(w, r, err)
	}
	cursor.Update(ongoingDeals)

	presented := presenter.NewDealList(ctx, ongoingDeals)
	if err := render.RenderList(w, r, presented); err != nil {
		render.Respond(w, r, err)
	}
}

func ListTimedDeal(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	cursor := ctx.Value("cursor").(*api.Page)

	orderBy := "end_at ASC"

	dealCond := db.Cond{
		"status":   data.DealStatusActive,
		"timed":    true,
		"featured": false,
	}

	var timedDeals []*data.Deal

	query := data.DB.Select("d.*").
		From("deals d").
		Where(dealCond).
		OrderBy(orderBy)

	paginate := cursor.UpdateQueryBuilder(query)

	if err := paginate.All(&timedDeals); err != nil {
		render.Respond(w, r, err)
	}
	cursor.Update(timedDeals)

	presented := presenter.NewDealList(ctx, timedDeals)
	if err := render.RenderList(w, r, presented); err != nil {
		render.Respond(w, r, err)
	}
}

func ListFeaturedDeal(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	cursor := ctx.Value("cursor").(*api.Page)

	orderBy := "value ASC"

	dealCond := db.Cond{
		"status":   data.DealStatusActive,
		"featured": true,
	}

	var featuredDeals []*data.Deal

	query := data.DB.Select("d.*").
		From("deals d").
		Where(dealCond).
		OrderBy(orderBy)

	paginate := cursor.UpdateQueryBuilder(query)

	if err := paginate.All(&featuredDeals); err != nil {
		render.Respond(w, r, err)
	}
	cursor.Update(featuredDeals)

	presented := presenter.NewDealList(ctx, featuredDeals)

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

func ListProducts(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	cursor := ctx.Value("cursor").(*api.Page)
	filterSort := ctx.Value("filter.sort").(*api.FilterSort)
	deal := ctx.Value("deal").(*data.Deal)

	var query sqlbuilder.Selector
	if deal.ProductListType == data.ProductListTypeAssociated {
		cond := db.Cond{
			"d.deal_id": deal.ID,
			"p.status":  data.ProductStatusApproved,
		}
		query = data.DB.Select("p.*").
			From("deal_products d").
			LeftJoin("products p").
			On("d.product_id = p.id").
			Where(cond).
			OrderBy("d.deal_id")
	} else if deal.ProductListType == data.ProductListTypeBXGY {
		var productIds []int64
		productIds = deal.BXGYPrerequisite.PrerequisiteProductIds
		productIds = append(productIds, deal.BXGYPrerequisite.EntitledProductIds...)

		query = data.DB.Select("p.*").
			From("products p").
			Where(
				db.Cond{
					"p.id":     productIds,
					"p.status": data.ProductStatusApproved,
				},
			).
			GroupBy("p.id").
			OrderBy("p.created_at DESC")
	} else {
		query = data.DB.Select("p.*").
			From("products p").
			Where(
				db.Cond{
					"p.place_id": deal.MerchantID,
					"p.status":   data.ProductStatusApproved,
				}).
			OrderBy("p.id DESC")
	}
	query = filterSort.UpdateQueryBuilder(query)

	var products []*data.Product
	paginate := cursor.UpdateQueryBuilder(query)
	if err := paginate.All(&products); err != nil {
		render.Respond(w, r, err)
	}
	cursor.Update(products)

	render.RenderList(w, r, presenter.NewProductList(ctx, products))

}
