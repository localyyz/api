package ping

import (
	"github.com/go-chi/chi"
	"github.com/go-chi/render"
	"github.com/pressly/lg"
	"net/http"
)

func Routes() chi.Router {
	r := chi.NewRouter()

	r.Post("/", LogDeviceData)

	return r
}

type DeviceData struct {
	InstallReferer string `json:"installreferer"`
	BuildNumber    string `json:"buildnumber"`
	Brand          string `json:"brand"`
	SystemName     string `json:"systemname"`
	DeviceID       string `json:"deviceid"`
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

	lg.SetEntryField(ctx, "device_id", logData.DeviceID)
	lg.SetEntryField(ctx, "system_name", logData.SystemName)
	lg.SetEntryField(ctx, "brand", logData.Brand)
	lg.SetEntryField(ctx, "install_referer", logData.InstallReferer)
	lg.SetEntryField(ctx, "build_number", logData.BuildNumber)

}
