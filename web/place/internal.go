package place

import (
	"net/http"
	"strconv"

	"bitbucket.org/moodie-app/moodie-api/data"
	"github.com/go-chi/render"
	db "upper.io/db.v3"
)

type internalUpdateRequest struct {
	ID          string           `json:"id"`
	Status      data.PlaceStatus `json:"status"`
	Gender      *data.Gender     `json:"gender"`
	StyleFemale *data.PlaceStyle `json:"styleFemale"`
	StyleMale   *data.PlaceStyle `json:"styleMale"`
	Pricing     string           `json:"pricing"`
}

func (*internalUpdateRequest) Bind(r *http.Request) error {
	return nil
}

func UpdateInternal(w http.ResponseWriter, r *http.Request) {
	var payload internalUpdateRequest
	if err := render.Bind(r, &payload); err != nil {
		render.Respond(w, r, err)
		return
	}

	placeID, err := strconv.ParseInt(payload.ID, 10, 64)
	if err != nil {
		render.Respond(w, r, err)
		return
	}

	place, err := data.DB.Place.FindByID(placeID)
	if err != nil {
		render.Respond(w, r, err)
		return
	}

	placeMeta, err := data.DB.PlaceMeta.FindByPlaceID(place.ID)
	if err != nil {
		if err != db.ErrNoMoreRows {
			render.Respond(w, r, err)
			return
		}
		placeMeta = &data.PlaceMeta{
			PlaceID: place.ID,
		}
	}

	switch data.PlaceStatus(payload.Status) {
	case data.PlaceStatusReviewing,
		data.PlaceStatusSelectPlan,
		data.PlaceStatusInActive:
		place.Status = payload.Status
	default:
		// ignore other status
	}

	if payload.Gender != nil &&
		(*payload.Gender == data.GenderMale ||
			*payload.Gender == data.GenderFemale) {
		placeMeta.Gender = payload.Gender
	}
	placeMeta.StyleMale = payload.StyleMale
	placeMeta.StyleFemale = payload.StyleFemale
	placeMeta.Pricing = payload.Pricing

	if err := data.DB.Place.Save(place); err != nil {
		render.Respond(w, r, err)
		return
	}
	if err := data.DB.PlaceMeta.Save(placeMeta); err != nil {
		render.Respond(w, r, err)
		return
	}

	render.Respond(w, r, "ok")
}
