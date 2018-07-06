package natsq

import (
	"log"

	nats "github.com/nats-io/go-nats"
)

type PubClient struct {
	Subject string
	enc     nats.Encoder
}

// default setup:
// - publisher connects with a default json encoder (mimics nats json encoder)
// - default async noop ack handler
func NewPublisher(subject, encoding string) (*PubClient, error) {
	encoder := defaultEncoderType
	if enc := nats.EncoderForType(encoding); enc != nil {
		encoder = enc
	}
	return &PubClient{
		Subject: subject,
		enc:     encoder,
	}, nil
}

func (c *PubClient) Publish(data interface{}) error {
	b, err := c.enc.Encode(c.Subject, data)
	if err != nil {
		return err
	}
	return DefaultClient.Publish(c.Subject, b)
}

func ackHandler(ackedNuid string, err error) {
	if err != nil {
		log.Printf("Warning: error publishing msg id %s: %v\n", ackedNuid, err.Error())
	} else {
		log.Printf("Received ack for msg id %s\n", ackedNuid)
	}
}

func (c *PubClient) PublishAsync(data interface{}) (string, error) {
	b, err := c.enc.Encode(c.Subject, data)
	if err != nil {
		return "", err
	}
	return DefaultClient.PublishAsync(c.Subject, b, ackHandler)
}

func (c *PubClient) Ping() error {
	return DefaultClient.Ping()
}
