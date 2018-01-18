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
	"bitbucket.org/moodie-app/moodie-api/lib/connect"
	"bitbucket.org/moodie-app/moodie-api/lib/slack"
	"bitbucket.org/moodie-app/moodie-api/web/api"
	"github.com/go-chi/render"
	db "upper.io/db.v3"
)

const (
	ApprovalActionName    = "approval"
	ApprovalActionApprove = "approve"
	ApprovalActionReject  = "reject"

	GenderActionName = "pick_gender"
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
	place, err := data.DB.Place.FindOne(db.Cond{"id": placeID})
	if err != nil {
		render.Respond(w, r, err)
		return
	}

	// send a response back to slack
	var actionResponse struct {
		Text        string              `json:"text"`
		Attachments []*slack.Attachment `json:"attachments,omitempty"`
	}

	// check the actions
	for _, a := range payload.Actions {
		// approve or reject
		switch a.Name {
		case ApprovalActionName:
			if a.Value == ApprovalActionApprove {
				place.Status = data.PlaceStatusActive
				place.ApprovedAt = data.GetTimeUTCPointer()

				// register the webhooks
				connect.SH.RegisterWebhooks(r.Context(), place)
				// if approved. attach the next choice
				actionResponse.Attachments = []*slack.Attachment{
					{
						Text:       fmt.Sprintf("what kind of product does %s sell?", place.Name),
						Fallback:   "",
						CallbackID: fmt.Sprintf("placeid%d", place.ID),
						Color:      "0195ff",
						Actions: []*slack.AttachmentAction{
							{
								Name:  GenderActionName,
								Text:  "Male",
								Type:  "button",
								Value: "male",
								Style: "primary",
							},
							{
								Name:  GenderActionName,
								Text:  "Female",
								Type:  "button",
								Value: "female",
								Style: "primary",
							},
							{
								Name:  GenderActionName,
								Text:  "Unisex",
								Type:  "button",
								Value: "unisex",
								Style: "primary",
							},
						},
					},
				}
			} else if a.Value == ApprovalActionReject {
				place.Status = data.PlaceStatusInActive
			}
			// approved or rejected.
			actionResponse.Text = fmt.Sprintf("%s is now %s! (by: %s)", place.Name, place.Status, payload.User.Name)
		case GenderActionName:
			if a.Value == "male" {
				place.Gender = data.PlaceGender(data.ProductGenderMale)
			} else if a.Value == "female" {
				place.Gender = data.PlaceGender(data.ProductGenderFemale)
			} else {
				place.Gender = data.PlaceGender(data.ProductGenderUnisex)
			}
			// response text
			actionResponse.Text = fmt.Sprintf("%s gender set to %v (by: %s)", place.Name, place.Gender, payload.User.Name)
		}
	}
	err = data.DB.Place.Save(place)
	if err != nil {
		actionResponse.Text = err.Error()
	}
	body, _ := json.Marshal(actionResponse)
	http.Post(payload.ResponseURL, "application/json", bytes.NewReader(body))
}
