package product

import (
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/render"

	db "upper.io/db.v3"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/data/presenter"
	"bitbucket.org/moodie-app/moodie-api/web/api"
)

func ListFeedProduct(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := ctx.Value("session.user").(*data.User)
	cursor := ctx.Value("cursor").(*api.Page)
	filterSort := ctx.Value("filter.sort").(*api.FilterSort)

	favs, err := data.DB.FavouritePlace.FindByUserID(user.ID)
	if err != nil {
		render.Respond(w, r, err)
		return
	}

	feedCond := db.Cond{
		"p.status": data.ProductStatusApproved,
	}
	if len(favs) > 0 {
		placeIDs := make([]int64, len(favs))
		for i, f := range favs {
			placeIDs[i] = f.PlaceID
		}
		feedCond["p.place_id"] = placeIDs
	} else {
		// use session user's preferences
		if user.Preference == nil {
			render.Respond(w, r, []struct{}{})
			return
		}
		placeIDs, err := data.DB.PlaceMeta.GetPlacesFromPreference(user.Preference)
		if err != nil {
			render.Respond(w, r, err)
			return
		}
		feedCond["p.place_id"] = placeIDs
	}

	// TODO: respect user sorting preference
	query := data.DB.
		Select(db.Raw("p.*")).
		From("products p").
		Where(feedCond).
		OrderBy("-p.id")
	query = filterSort.UpdateQueryBuilder(query)

	var products []*data.Product
	paginate := cursor.UpdateQueryBuilder(query)
	if err := paginate.All(&products); err != nil {
		render.Respond(w, r, err)
		return
	}
	cursor.Update(products)

	presented := presenter.NewProductList(ctx, products)
	if err := render.RenderList(w, r, presented); err != nil {
		render.Respond(w, r, err)
	}
}

func ListRandomProduct(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	cursor := ctx.Value("cursor").(*api.Page)
	filterSort := ctx.Value("filter.sort").(*api.FilterSort)

	hardCond := db.Raw(`p.tsv @@ (to_tsquery($$puma$$) ||
				to_tsquery('simple', $$puma:*$$) ||
				to_tsquery($$puma:*$$) ||
				to_tsquery('simple', $$puma$$) ||
				
				to_tsquery($$nike$$) ||
				to_tsquery('simple', $$nike:*$$) ||
				to_tsquery($$nike:*$$) ||
				to_tsquery('simple', $$nike$$) ||

				to_tsquery($$yeezy$$) ||
				to_tsquery('simple', $$yeezy:*$$) ||
				to_tsquery($$yeezy:*$$) ||
				to_tsquery('simple', $$yeezy$$) ||
				
				to_tsquery($$supreme$$) ||
				to_tsquery('simple', $$supreme:*$$) ||
				to_tsquery($$supreme:*$$) ||
				to_tsquery('simple', $$supreme$$) ||

				to_tsquery($$adidas$$) ||
				to_tsquery('simple', $$adidas:*$$) ||
				to_tsquery($$adidas:*$$) ||
				to_tsquery('simple', $$adidas$$) ||

				to_tsquery($$moschino$$) ||
				to_tsquery($$untitled$$) ||
				
				to_tsquery($$jordans$$) ||
				to_tsquery('simple', $$jordans:*$$) ||
				to_tsquery($$jordans:*$$) ||
				to_tsquery('simple', $$jordans$$))
	`)
	cond := db.And(
		db.Cond{
			"p.status": data.ProductStatusApproved,
			"p.score":  db.Gte(4),
			db.Raw("p.category->>'type'"): []string{
				"apparel",
				"shoes",
				"sneakers",
			},
		},
		hardCond,
	)

	t := time.Now().Truncate(time.Hour).Unix()
	cursor.Limit = 20 // hard coded
	cursor.ItemTotal = 10000
	query := data.DB.Select("p.*").
		From("products p").
		Where(cond).
		OrderBy(db.Raw(fmt.Sprintf("%d %% id", t)))
	query = filterSort.UpdateQueryBuilder(query)

	var products []*data.Product
	if !filterSort.HasFilter() {
		paginate := cursor.UpdateQueryBuilder(query)
		if err := paginate.All(&products); err != nil {
			render.Respond(w, r, err)
			return
		}
		cursor.Update(products)
	}

	presented := presenter.NewProductList(ctx, products)
	if err := render.RenderList(w, r, presented); err != nil {
		render.Respond(w, r, err)
	}
}
