package ping

import (
	"net/http"

	"bitbucket.org/moodie-app/moodie-api/data"
	"github.com/go-chi/render"
	"github.com/pressly/lg"
)

type DeviceData struct {
	InstallReferer    string `json:"installReferer"`
	BuildNumber       string `json:"buildNumber"`
	Brand             string `json:"brand"`
	SystemName        string `json:"systemName"`
	DeviceID          string `json:"deviceID"`
	UniqueID          string `json:"uniqueID"`
	CodePushVersion   string `json:"codePushVersion"`
	OneSignalPlayerID string `json:"playerId"`
}

func (d *DeviceData) Bind(r *http.Request) error {
	return nil
}

func LogDeviceData(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logData := &DeviceData{}

	if err := render.Bind(r, logData); err != nil {
		lg.Info("Error: Failed to log device data")
		return
	}

	if logData.UniqueID != "" {
		lg.SetEntryField(ctx, "device_id", logData.DeviceID)
		lg.SetEntryField(ctx, "unique_id", logData.UniqueID)
		lg.SetEntryField(ctx, "system_name", logData.SystemName)
		lg.SetEntryField(ctx, "brand", logData.Brand)
		lg.SetEntryField(ctx, "install_referer", logData.InstallReferer)
		lg.SetEntryField(ctx, "build_number", logData.BuildNumber)
		lg.SetEntryField(ctx, "code_push_version", logData.CodePushVersion)
	}

	if logData.OneSignalPlayerID != "" {
		// this saves the onesignal player id
		user, ok := ctx.Value("session.user").(*data.User)
		if ok && user.Etc.OSPlayerID != logData.OneSignalPlayerID {
			user.Etc.OSPlayerID = logData.OneSignalPlayerID
			data.DB.User.Save(user)
		}
	}
}
