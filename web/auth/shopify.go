package auth

import (
	"fmt"
	"net/http"
	"strconv"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/lib/connect"
	"bitbucket.org/moodie-app/moodie-api/lib/token"
	"bitbucket.org/moodie-app/moodie-api/lib/ws"
)

// callback from initiating AuthCodeURL
func ShopifyOAuthCb(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	// TODO: check HMAC signature

	//code := q.Get("code")
	state := q.Get("state")
	shop := q.Get("shop")

	token, err := token.Decode(state)
	if err != nil {
		ws.Respond(w, http.StatusBadRequest, err)
		return
	}

	pId, ok := token.Claims["place_id"].(string)
	if !ok {
		ws.Respond(w, http.StatusBadRequest, connect.ErrInvalidState)
		return
	}

	placeId, _ := strconv.ParseInt(pId, 10, 64)
	place, err := data.DB.Place.FindByID(placeId)

	if err != nil {
		ws.Respond(w, http.StatusInternalServerError, err)
		return
	}

	if fmt.Sprintf("%s.myshopify.com", place.ShopifyID) != shop {
		ws.Respond(w, http.StatusBadRequest, "")
		return
	}

	tok, err := connect.SH.Exchange(place, r)
	if err != nil {
		ws.Respond(w, http.StatusBadRequest, err)
		return
	}

	// save authorization
	cred := &data.ShopifyCred{
		PlaceID:     place.ID,
		AccessToken: tok.AccessToken,
	}
	if err := data.DB.ShopifyCred.Save(cred); err != nil {
		ws.Respond(w, http.StatusInternalServerError, err)
		return
	}

	ws.Respond(w, http.StatusCreated, "shopify connected.")
}
