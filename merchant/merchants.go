package merchant

import (
	"fmt"
	"net/http"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/lib/connect"
	"bitbucket.org/moodie-app/moodie-api/lib/slack"
	"github.com/go-chi/render"
)

func AcceptTOS(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	place := ctx.Value("place").(*data.Place)

	// must be in "Wait Agreement" status to accept
	if place.Status != data.PlaceStatusWaitAgreement {
		return
	}

	// proceed to next status
	place.Status = data.PlaceStatusWaitApproval
	place.TOSAgreedAt = data.GetTimeUTCPointer()
	place.TOSIP = r.RemoteAddr
	if err := data.DB.Place.Save(place); err != nil {
		// error has occured. respond
		render.Status(r, http.StatusInternalServerError)
		render.Respond(w, r, err)
		return
	}

	// notify slack with button for approving the merchant
	sl := ctx.Value("slack.client").(*connect.Slack)
	sl.Notify(
		"store",
		fmt.Sprintf("<%s|%s> (id: %v) just accepted the TOS!", place.Website, place.Name, place.ID),
		&slack.Attachment{
			Title:      "start review process:",
			TitleLink:  fmt.Sprintf("https://merchant.localyyz.com/approval/%d", place.ID),
			Fallback:   "You are unable to approve / reject the store.",
			CallbackID: fmt.Sprintf("placeid%d", place.ID),
			Color:      "0195ff",
		},
	)

	render.Respond(w, r, "success")
	return
}
