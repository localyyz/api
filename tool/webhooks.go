package tool

import (
	"context"
	"net/http"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/lib/connect"
	db "upper.io/db.v3"
)

func AddWebhooks(w http.ResponseWriter, r *http.Request) {

	places, _ := data.DB.Place.FindAll(db.Cond{"status": data.PlaceStatusActive})

	for _, p := range places {
		ctx := context.Background()
		connect.SH.RegisterWebhooks(ctx, p)
	}

}
