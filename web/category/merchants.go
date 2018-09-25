package category

import (
	"errors"
	"net/http"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/data/presenter"
	"bitbucket.org/moodie-app/moodie-api/web/api"
	"github.com/go-chi/render"
	"github.com/pressly/lg"
	db "upper.io/db.v3"
)

type categoryMerchantRequest struct {
	Pricing []string `json:"pricing"`
	Gender  []string `json:"gender"`
	Style   []string `json:"style"`
}

func (s *categoryMerchantRequest) Bind(r *http.Request) error {
	var err error
	if len(s.Gender) == 0 {
		err = errors.New("missing gender")
		return api.ErrInvalidRequest(err)
	} else if len(s.Pricing) == 0 {
		err = errors.New("missing pricing")
	} else if len(s.Style) == 0 {
		// TODO
	}

	if err != nil {
		return api.ErrInvalidRequest(err)
	}

	return nil
}

func ListMerchants(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	cursor := ctx.Value("cursor").(*api.Page)

	var catRequest categoryMerchantRequest
	if err := render.Bind(r, &catRequest); err != nil {
		render.Respond(w, r, err)
		return
	}

	styleCol := "style_female"
	if catRequest.Gender[0] == "man" {
		styleCol = "style_male"
	}
	iter := data.DB.Select("place_id").
		From("place_meta").
		Where(
			db.And(
				db.Or(
					db.Cond{"gender": catRequest.Gender},
					db.Cond{"gender": db.IsNull()},
				),
				db.Cond{
					"pricing": catRequest.Pricing,
					styleCol:  catRequest.Style,
				},
			),
		).
		OrderBy("-place_id").
		Iterator()

	var placeIDs []int64
	for iter.Next() {
		var ID int64
		if err := iter.Scan(&ID); err != nil {
			lg.Warn(err)
			continue
		}
		placeIDs = append(placeIDs, ID)
	}
	if err := iter.Err(); err != nil {
		render.Respond(w, r, err)
		return
	}

	query := data.DB.Place.Find(
		db.Cond{
			"id":     placeIDs,
			"status": data.PlaceStatusActive,
		},
	).OrderBy("-id")
	query = cursor.UpdateQueryUpper(query)

	var places []*data.Place
	if err := query.All(&places); err != nil {
		render.Respond(w, r, err)
		return
	}
	cursor.Update(places)

	presented := presenter.NewPlaceList(ctx, places)
	render.RenderList(w, r, presented)
}
