package pusher

import (
	"os"

	"github.com/pkg/errors"
	apns "github.com/sideshow/apns2"
	"github.com/sideshow/apns2/certificate"
)

type Client struct {
	*apns.Client

	Topic string
}

var (
	PusherApiError = errors.New("pusher api failed")
	client         *Client
)

func Setup(pemFile, topic, env string) error {
	if pemFile == "" {
		pemFile = os.Getenv("PEM")
	} // try environment var

	cert, err := certificate.FromPemFile(pemFile, "")
	if err != nil {
		return err
	}

	// setup app bundle topic name
	client = &Client{apns.NewClient(cert), topic}
	if env == "production" {
		client.Production()
	}

	return nil
}

func Push(deviceToken string, payload []byte) error {
	if client == nil {
		return nil
	}

	notification := &apns.Notification{}
	notification.DeviceToken = deviceToken
	notification.Topic = client.Topic
	notification.Payload = payload

	res, err := client.Push(notification)
	if err != nil {
		return errors.Wrap(err, "pusher response error")
	}

	if !res.Sent() {
		return errors.Wrapf(PusherApiError, "status: %d with reason: %s", res.StatusCode, res.Reason)
	}

	return nil
}
