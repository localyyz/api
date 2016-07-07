package place

import (
	"net/http"
	"strconv"

	"golang.org/x/net/context"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/lib/ws"
	"bitbucket.org/moodie-app/moodie-api/web/utils"
	"github.com/pressly/chi"
)

func PlaceCtx(next chi.Handler) chi.Handler {
	handler := func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		placeID, err := strconv.ParseInt(chi.URLParam(ctx, "placeID"), 10, 64)
		if err != nil {
			ws.Respond(w, http.StatusBadRequest, utils.ErrBadID)
			return
		}

		place, err := data.DB.Place.FindByID(placeID)
		if err != nil {
			ws.Respond(w, http.StatusInternalServerError, err)
			return
		}
		ctx = context.WithValue(ctx, "place", place)
		next.ServeHTTPC(ctx, w, r)
	}
	return chi.HandlerFunc(handler)
}

func GetPlace(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	place := ctx.Value("place").(*data.Place)

	locale, err := data.DB.Locale.FindByID(place.LocaleID)
	if err != nil {
		ws.Respond(w, http.StatusInternalServerError, err)
		return
	}

	resp := &data.PlacePresenter{
		Place:  place,
		Locale: locale,
	}
	ws.Respond(w, http.StatusOK, resp)
}
