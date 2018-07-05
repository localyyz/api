package natsq

import (
	nats "github.com/nats-io/go-nats"
	"github.com/nats-io/go-nats-streaming"
)

type QClient interface {
	QSubscriber
	QPublisher
	QPinger
	QCloser
}

type QSubscriber interface {
	Subscribe(subject string, cb stan.MsgHandler, opts ...stan.SubscriptionOption) (stan.Subscription, error)
	QueueSubscribe(subject, qgroup string, cb stan.MsgHandler, opts ...stan.SubscriptionOption) (stan.Subscription, error)
}

type QPublisher interface {
	Publish(subject string, data []byte) error
	PublishAsync(subject string, data []byte, ah stan.AckHandler) (string, error)
}

type QPinger interface {
	Ping() error
}

type QCloser interface {
	Close() error
}

var (
	defaultEncoderType = nats.EncoderForType(nats.JSON_ENCODER)
)
