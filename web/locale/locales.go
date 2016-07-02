package locale

import (
	"net/http"
	"strconv"

	"github.com/pressly/chi"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/lib/ws"
	"bitbucket.org/moodie-app/moodie-api/web/utils"

	"golang.org/x/net/context"
)

func LocaleCtx(next chi.Handler) chi.Handler {
	handler := func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		localeID, err := strconv.ParseInt(chi.URLParam(ctx, "localeID"), 10, 64)
		if err != nil {
			ws.Respond(w, http.StatusBadRequest, utils.ErrBadID)
			return
		}

		locale, err := data.DB.Locale.FindByID(localeID)
		if err != nil {
			ws.Respond(w, http.StatusInternalServerError, err)
			return
		}
		ctx = context.WithValue(ctx, "locale", locale)
		next.ServeHTTPC(ctx, w, r)
	}
	return chi.HandlerFunc(handler)
}

func ListLocales(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	cursor := ws.NewPage(r)
	q := data.DB.Locale.Find().Sort("-id")
	q = cursor.UpdateQueryUpper(q)

	var resp []*data.Locale
	if err := q.All(&resp); err != nil {
		ws.Respond(w, http.StatusInternalServerError, err)
		return
	}

	ws.Respond(w, http.StatusOK, resp, cursor.Update(resp))
}
