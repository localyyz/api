package natsq

import (
	"errors"

	"github.com/nats-io/go-nats-streaming"
)

// Noop implements both subscriber and publisher
type NoopClient struct{}

func (*NoopClient) Ping() error {
	return errors.New("natsq: NoopClient")
}

func (*NoopClient) Close() error {
	return errors.New("natsq: NoopClient")
}

func (*NoopClient) Subscribe(subject string, cb stan.MsgHandler, opts ...stan.SubscriptionOption) (stan.Subscription, error) {
	return nil, nil
}

func (*NoopClient) QueueSubscribe(subject, qgroup string, cb stan.MsgHandler, opts ...stan.SubscriptionOption) (stan.Subscription, error) {
	return nil, nil
}

func (*NoopClient) Publish(subject string, data []byte) error {
	return nil
}

func (*NoopClient) PublishAsync(subject string, data []byte, ah stan.AckHandler) (string, error) {
	return "", nil
}
