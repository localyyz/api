package place

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/lib/slack"
	"bitbucket.org/moodie-app/moodie-api/web/api"
	"github.com/go-chi/render"
	db "upper.io/db.v3"
)

const (
	ApprovalActionName    = "approval"
	ApprovalActionApprove = "approve"
	ApprovalActionReject  = "reject"

	LocationActionName = "location_list"
)

var (
	errInvalidCallback = errors.New("invalid callback")
)

type slackApprovalRequest struct {
	*slack.Attachment
	User struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"user"`
	ResponseURL string `json:"response_url,omitempty"`
}

func HandleApproval(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	formData := r.PostForm.Get("payload")
	var payload slackApprovalRequest
	if err := json.Unmarshal([]byte(formData), &payload); err != nil {
		render.Respond(w, r, api.ErrInvalidRequest(err))
		return
	}

	if len(payload.CallbackID) == 0 {
		render.Respond(w, r, api.ErrInvalidRequest(errInvalidCallback))
		return
	}

	// parse the callback id
	placeID, err := strconv.Atoi(strings.TrimPrefix(payload.CallbackID, "placeid"))
	if err != nil {
		render.Respond(w, r, api.ErrInvalidRequest(errInvalidCallback))
		return
	}

	// find the place
	place, err := data.DB.Place.FindOne(db.Cond{"id": placeID, "status": data.PlaceStatusWaitApproval})
	if err != nil {
		render.Respond(w, r, err)
		return
	}

	// send a response back to slack
	var actionResponse struct {
		Text string `json:"text"`
	}

	// approve or reject
	for _, a := range payload.Actions {
		if a.Name == ApprovalActionName {
			if a.Value == ApprovalActionApprove {
				place.Status = data.PlaceStatusActive
				place.ApprovedAt = data.GetTimeUTCPointer()
			} else if a.Value == ApprovalActionReject {
				place.Status = data.PlaceStatusInActive
			}
			actionResponse.Text = fmt.Sprintf("%s is now %s! (by: %s)", place.Name, place.Status, payload.User.Name)
		} else if a.Name == LocationActionName {
			for _, o := range a.SelectedOptions {
				localeID, _ := strconv.Atoi(strings.TrimPrefix(o.Value, "localeid"))
				place.LocaleID = int64(localeID)
				actionResponse.Text = fmt.Sprintf("place location updated!")
			}
		}
	}
	err = data.DB.Place.Save(place)
	if err != nil {
		actionResponse.Text = err.Error()
	}
	body, _ := json.Marshal(actionResponse)
	http.Post(payload.ResponseURL, "application/json", bytes.NewReader(body))
}
