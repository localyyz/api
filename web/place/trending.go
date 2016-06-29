package place

import (
	"net/http"
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

	n := time.Now().AddDate(0, 0, -2) // 2 days rolling
	cond := db.Cond{"created_at >=": n}

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
