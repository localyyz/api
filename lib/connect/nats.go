package connect

import (
	"fmt"
	"os"

	"bitbucket.org/moodie-app/moodie-api/lib/events"
	"bitbucket.org/moodie-app/moodie-api/lib/natsq"
	"github.com/gosimple/slug"
	"github.com/nats-io/go-nats-streaming"
	"github.com/pkg/errors"
	"github.com/pressly/lg"
)

type Nats struct {
	serverURL   string
	publishers  map[events.Event]*natsq.PubClient
	subscribers map[events.Event]*natsq.SubClient

	sc stan.Conn
}

type NatsSubject struct {
	Subject     string `toml:"subject"`
	DurableName string `toml:"durable_name"`
	GroupName   string `toml:"group_name"`
	Encoding    string `toml:"encoding"`
}

var (
	NATS                 *Nats
	ErrInvalidPubSubject = errors.New("invalid pub subject")
	ErrInvalidSubSubject = errors.New("invalid sub subject")
)

func SetupNatsStream(config NatsConfig) *Nats {

	NATS = &Nats{
		publishers:  map[events.Event]*natsq.PubClient{},
		subscribers: map[events.Event]*natsq.SubClient{},
	}

	// setup the default client first
	hostName, _ := os.Hostname()
	clientID := fmt.Sprintf("%s %s", config.AppName, hostName)

	sc, err := natsq.NewClient(config.ClusterID, slug.Make(clientID), config.ServerURL)
	if err != nil {
		lg.Warnf("connect to nats %+v", err)
		return nil
	}
	NATS.sc = sc
	for _, pubs := range config.Publishers {
		client, err := natsq.NewPublisher(pubs.Subject, pubs.Encoding)
		if err != nil {
			lg.Fatalf("connect to publisher %+v", err)
		}
		NATS.publishers[events.EventForType(pubs.Subject)] = client
	}
	for _, subs := range config.Subscribers {
		client, err := natsq.NewSubscriber(subs.Subject, subs.DurableName, subs.GroupName, subs.Encoding)
		if err != nil {
			lg.Fatalf("connect to subscriber %+v", err)
		}
		NATS.subscribers[events.EventForType(subs.Subject)] = client
	}
	return NATS
}

func (n *Nats) Emit(event events.Event, data interface{}) (string, error) {
	c, ok := n.publishers[event]
	if !ok {
		return "", ErrInvalidPubSubject
	}
	// returns GUID and/or error
	return c.PublishAsync(data)
}

func (n *Nats) Subscribe(event events.Event, cb natsq.Handler) (stan.Subscription, error) {
	c, ok := n.subscribers[event]
	if !ok {
		return nil, ErrInvalidSubSubject
	}
	return c.Subscribe(cb)
}

func (n *Nats) Unsubscribe(event events.Event) error {
	c, ok := n.subscribers[event]
	if !ok {
		return ErrInvalidSubSubject
	}
	return c.Unsubscribe()
}

func (n *Nats) UnsubscribeAll() {
	for _, c := range n.subscribers {
		c.Unsubscribe()
	}
}

func (n *Nats) Close() error {
	if n.sc != nil {
		return n.sc.Close()
	}
	return nil
}
