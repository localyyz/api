package web

import (
	"io/ioutil"
	"net/http"
	"os"

	"bitbucket.org/moodie-app/moodie-api/lib/ws"
	"github.com/pkg/errors"
)

func Manifest(w http.ResponseWriter, r *http.Request) {
	payload := `<?xml version="1.0" encoding="UTF-8"?>
		<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
		<plist version="1.0">
		<dict>
			<key>items</key>
			<array>
				<dict>
					<key>assets</key>
					<array>
						<dict>
							<key>kind</key>
							<string>software-package</string>
							<key>url</key>
							<string>https://api.localyyz.com/Moodie.ipa</string>
						</dict>
						<dict>
							<key>kind</key>
							<string>display-image</string>
							<key>url</key>
							<string>http://localyyz.com/57.png</string>
						</dict>
						<dict>
							<key>kind</key>
							<string>full-size-image</string>
							<key>url</key>
							<string>https://localyyz.com/512.png</string>
						</dict>
					</array>
					<key>metadata</key>
					<dict>
						<key>bundle-identifier</key>
						<string>org.moodie.toronto</string>
						<key>bundle-version</key>
						<string>1.0.0</string>
						<key>kind</key>
						<string>software</string>
						<key>title</key>
						<string>Moodie</string>
					</dict>
				</dict>
			</array>
		</dict>
		</plist>`
	w.Header().Set("Content-Disposition", "attachement; filename=manifest.plist")
	w.Write([]byte(payload))
}

func MoodieApp(w http.ResponseWriter, r *http.Request) {
	f, err := os.Open("/etc/Moodie.ipa")
	if err != nil {
		ws.Respond(w, 404, errors.Wrap(err, "app download error"))
		return
	}
	defer f.Close()
	w.Header().Set("Content-Disposition", "attachement; filename=Moodie.ipa")
	binary, err := ioutil.ReadAll(f)
	if err != nil {
		ws.Respond(w, 404, errors.Wrap(err, "app download error"))
		return
	}
	w.Write(binary)
}
