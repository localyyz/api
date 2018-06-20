package ping

import (
	"encoding/json"
	"github.com/go-chi/chi"
	"github.com/pressly/lg"
	"net/http"
)

func Routes() chi.Router {
	r := chi.NewRouter()

	r.Post("/", LogDeviceData)

	return r
}

type DeviceData struct {
	InstallReferer string
	BuildNumber    string
	Brand          string
	SystemName     string
	DeviceID       string
}

func LogDeviceData(w http.ResponseWriter, r *http.Request) {
	var logData DeviceData
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&logData)
	if err != nil {
		lg.Info("Error: Failed to log device data")
	} else {
		lg.Infof("Device ID: %s | System Name: %s | Brand: %s | Install Referer: %s | Build Number: %s",
			logData.DeviceID,
			logData.SystemName,
			logData.Brand,
			logData.InstallReferer,
			logData.BuildNumber)
	}
}
