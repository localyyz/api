package natsq

import (
	"errors"
	"log"

	"github.com/nats-io/go-nats-streaming"
)

type Client struct {
	stan.Conn

	// is client still actively connected to the server
	disconnected bool

	Service string
}

var (
	DefaultClient         QClient = &NoopClient{}
	ErrClientDisconnected         = errors.New("disconnected")
)

func NewClient(clusterID, clientID, serverURL string) (*Client, error) {
	sc, err := ConnectNATS(clusterID, clientID, serverURL)
	if err != nil {
		return nil, err
	}
	c := &Client{
		Conn:    sc,
		Service: clientID,
	}
	DefaultClient = c
	return c, nil
}

func ConnectNATS(clusterID, clientID, serverURL string) (stan.Conn, error) {
	return stan.Connect(clusterID, clientID,
		stan.NatsURL(serverURL),
		stan.SetConnectionLostHandler(func(_ stan.Conn, reason error) {
			log.Fatalf("Connection lost, reason: %v", reason)
		}),
	)
}

func (c *Client) Close() error {
	c.disconnected = true
	return c.Conn.Close()
}

func (c *Client) Ping() error {
	if c.disconnected {
		return ErrClientDisconnected
	}
	return nil
}
