package category

import (
	"context"
	"net/http"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/data/presenter"
	"bitbucket.org/moodie-app/moodie-api/web/api"
	"github.com/go-chi/render"
	db "upper.io/db.v3"
)

func listMerchants(ctx context.Context, catIDs []int64) ([]*data.Place, error) {
	cursor := ctx.Value("cursor").(*api.Page)

	iter := data.DB.
		Select(db.Raw("p.place_id")).
		From("products p").
		Where(db.Cond{
			"category_id": catIDs,
			"status":      data.ProductStatusApproved,
			"score":       db.Gte(3),
		}).
		OrderBy(db.Raw("count(1) desc")).
		GroupBy("p.place_id").
		Iterator()
	defer iter.Close()

	var placeIDs []int64
	for iter.Next() {
		var pID int64
		if err := iter.Scan(&pID); err != nil {
			break
		}
		placeIDs = append(placeIDs, pID)
	}
	if err := iter.Err(); err != nil {
		return nil, err
	}

	query := data.DB.Place.Find(
		db.Cond{
			"id":     placeIDs,
			"status": data.PlaceStatusActive,
		},
	).OrderBy(data.MaintainOrder("id", placeIDs))
	query = cursor.UpdateQueryUpper(query)

	var places []*data.Place
	if err := query.All(&places); err != nil {
		return nil, err
	}
	cursor.Update(places)

	return places, nil
}

type categoryMerchantRequest struct {
	CategoryIDs []int64 `json:"categories"`
}

func (*categoryMerchantRequest) Bind(r *http.Request) error {
	return nil
}

func ListMerchants(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var catIDs []int64
	if category, ok := ctx.Value("category").(*data.Category); ok {
		catIDs = []int64{category.ID}
		// for the category
		cats, _ := data.DB.Category.FindDescendants(category.ID)
		for _, c := range cats {
			catIDs = append(catIDs, c.ID)
		}
	} else {
		var catRequest categoryMerchantRequest
		if err := render.Bind(r, &catRequest); err != nil {
			render.Respond(w, r, err)
			return
		}
		for _, catID := range catRequest.CategoryIDs {
			cats, _ := data.DB.Category.FindDescendants(catID)
			catIDs = append(catIDs, catID)
			for _, c := range cats {
				catIDs = append(catIDs, c.ID)
			}
		}
	}

	places, err := listMerchants(ctx, catIDs)
	if err != nil {
		render.Respond(w, r, err)
		return
	}

	presented := presenter.NewPlaceList(ctx, places)
	render.RenderList(w, r, presented)
}
