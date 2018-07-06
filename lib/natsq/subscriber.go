package natsq

import (
	"errors"
	"reflect"

	nats "github.com/nats-io/go-nats"
	"github.com/nats-io/go-nats-streaming"
)

type SubClient struct {
	// Durable name is used when (re-)connecting to nats streaming server and
	// automatically start from where last was left off.
	// nats streaming will track the last acknowledged message for that clientID + durable name,
	// so that only messages since the last acknowledged message will be delivered to the client.
	// more documentation: https://github.com/nats-io/go-nats-streaming
	DurableName string

	// All subscriptions with the same queue name (regardless of the connection they originate from)
	// will form a queue group. Each message will be delivered to only one subscriber per queue group,
	// using queuing semantics. You can have as many queue groups as you wish.
	QueueName string

	// Subject to subscribe to
	Subject string

	// underlying subscription
	sc  stan.Subscription
	enc nats.Encoder
}

var (
	ErrSubscriberDisconnected = errors.New("subscriber disconnected")
)

// pre-configured nats streaming server subscriber:
func NewSubscriber(subject, durableName, qgroup, encoding string) (*SubClient, error) {
	encoder := defaultEncoderType
	if enc := nats.EncoderForType(encoding); enc != nil {
		encoder = enc
	}
	return &SubClient{
		enc:         encoder,
		Subject:     subject,
		DurableName: durableName,
		QueueName:   qgroup,
	}, nil
}

// Handler is a specific callback used for Subscribe. It is generalized to
// an interface{}, but we will discover its format and arguments at runtime
// and perform the correct callback, including de-marshaling JSON strings
// back into the appropriate struct based on the signature of the Handler.
//
// Handlers are expected to have one of four signatures.
//
//	type person struct {
//		Name string `json:"name,omitempty"`
//		Age  uint   `json:"age,omitempty"`
//	}
//
//	handler := func(m *Msg)
//	handler := func(p *person)
//	handler := func(subject string, o *obj)
//	handler := func(subject, reply string, o *obj)
//
// These forms allow a callback to request a raw Msg ptr, where the processing
// of the message from the wire is untouched. Process a JSON representation
// and demarshal it into the given struct, e.g. person.
// There are also variants where the callback wants either the subject, or the
// subject and the reply subject.
type Handler interface{}

// Dissect the cb Handler's signature
func argInfo(cb Handler) (reflect.Type, int) {
	cbType := reflect.TypeOf(cb)
	if cbType.Kind() != reflect.Func {
		panic("nats: Handler needs to be a func")
	}
	numArgs := cbType.NumIn()
	if numArgs == 0 {
		return nil, numArgs
	}
	return cbType.In(numArgs - 1), numArgs
}

var emptyMsgType = reflect.TypeOf(&stan.Msg{})

// subscribe mimics nats EncodedConn subscribe. it wraps a stan.MstHandler with a decoder
// take a look at go-nats/enc.go for the full subscribe function. We've omitted
// a lot of type and error checking because we're assuming it's safe
func (c *SubClient) decodedCb(cb Handler) (stan.MsgHandler, error) {
	if cb == nil {
		return nil, errors.New("natsq: Handler required for subscription")
	}

	argType, numArgs := argInfo(cb)
	if argType == nil {
		return nil, errors.New("natsq: Handler requires at least one argument")
	}
	cbValue := reflect.ValueOf(cb)

	return func(m *stan.Msg) {
		var oV []reflect.Value
		var oPtr reflect.Value
		if argType.Kind() != reflect.Ptr {
			oPtr = reflect.New(argType)
		} else {
			oPtr = reflect.New(argType.Elem())
		}
		if err := c.enc.Decode(m.Subject, m.Data, oPtr.Interface()); err != nil {
			return
		}
		if argType.Kind() != reflect.Ptr {
			oPtr = reflect.Indirect(oPtr)
		}

		// Callback Arity
		switch numArgs {
		case 1:
			oV = []reflect.Value{oPtr}
		case 2:
			subV := reflect.ValueOf(m.Subject)
			oV = []reflect.Value{subV, oPtr}
		case 3:
			subV := reflect.ValueOf(m.Subject)
			replyV := reflect.ValueOf(m.Reply)
			oV = []reflect.Value{subV, replyV, oPtr}
		}

		cbValue.Call(oV)
	}, nil
}

func (c *SubClient) Subscribe(cb Handler) (stan.Subscription, error) {
	stanCb, err := c.decodedCb(cb)
	if err != nil {
		return nil, err
	}
	c.sc, err = DefaultClient.Subscribe(c.Subject, stanCb, stan.DurableName(c.DurableName))
	return c.sc, err
}

func (c *SubClient) QueueSubscribe(cb Handler) (stan.Subscription, error) {
	stanCb, err := c.decodedCb(cb)
	if err != nil {
		return nil, err
	}
	c.sc, err = DefaultClient.QueueSubscribe(c.Subject, c.QueueName, stanCb, stan.DurableName(c.DurableName))
	return c.sc, err
}

func (c *SubClient) Unsubscribe() error {
	return c.sc.Unsubscribe()
}

func (c *SubClient) Ping() error {
	if err := DefaultClient.Ping(); err != nil {
		return err
	}
	if !c.sc.IsValid() {
		return ErrSubscriberDisconnected
	}
	return nil
}
