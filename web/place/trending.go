package place

import (
	"net/http"
	"strings"
	"time"

	"upper.io/db"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/lib/ws"

	"golang.org/x/net/context"
)

type postScore struct {
	PlaceID int64 `db:"place_id"`
	Scores  int32 `db:"scores"`
}

func ListTrendingPlaces(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	cursor := ws.NewPage(r)

	// mood type
	// NOTE: this is not REST, but let's just pretend it is
	placeCond := db.Cond{}
	if pType := strings.TrimSpace(r.URL.Query().Get("placeType")); pType != "" {
		// try to parse a placetype out of it
		var p data.PlaceType
		if err := p.UnmarshalText([]byte(pType)); err != nil {
			ws.Respond(w, http.StatusBadRequest, err)
			return
		}
		placeCond["place_type"] = p
	}
	if lId := strings.TrimSpace(r.URL.Query().Get("localeId")); lId != "" {
		placeCond["locale_id"] = lId
	}

	var places []*data.Place
	if err := data.DB.Place.Find(placeCond).All(&places); err != nil {
		ws.Respond(w, http.StatusInternalServerError, err)
		return
	}

	// get list of place ids
	var placeIDs []int64
	for _, p := range places {
		placeIDs = append(placeIDs, p.ID)
	}

	n := time.Now().AddDate(0, 0, -2) // 2 days rolling
	cond := db.Cond{
		"created_at >=": n,
		"place_id":      placeIDs,
	}

	// order list of places by post score in the last 48h
	var scores []postScore
	err := data.DB.Post.Find(cond).
		Select("place_id", "SUM(score) AS scores").
		Group("place_id").
		Sort("-scores").
		All(&scores)
	if err != nil {
		ws.Respond(w, http.StatusInternalServerError, err)
		return
	}

	// find the venues and respond
	var resp []*data.PlaceWithPost

	for _, s := range scores {
		p, err := data.DB.Place.FindByID(s.PlaceID)
		if err != nil {
			continue
		}

		var posts []*data.Post
		cond := db.Cond{"place_id": p.ID, "created_at >=": n}
		err = data.DB.Post.Find(cond).Sort("-score").Limit(5).All(&posts)
		if err != nil {
			continue
		}
		resp = append(resp, &data.PlaceWithPost{Place: p, Posts: posts})
	}

	ws.Respond(w, http.StatusOK, resp, cursor.Update(resp))
}
