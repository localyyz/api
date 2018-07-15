package connect

import (
	"bitbucket.org/moodie-app/moodie-api/lib/onesignal"
)

var (
	ON *onesignal.Client
)

func SetupOneSignal(conf Config) *onesignal.Client {
	ON = onesignal.NewClient(nil, conf.AppSecret)
	return ON
}
