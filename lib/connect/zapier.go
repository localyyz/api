package connect

import (
	"bitbucket.org/moodie-app/moodie-api/lib/zapier"
	"github.com/pkg/errors"
)

var (
	ErrZapierChan = errors.New("unknown zapier channel")
)

type Zapier struct {
	// message event -> zapier webhook mapping
	channels map[string]*zapier.Webhook
}

var (
	ZP *Zapier
)

func SetupZapier(confs ZapierConfig) *Zapier {
	ZP = &Zapier{
		channels: map[string]*zapier.Webhook{},
	}
	for t, conf := range confs.Webhooks {
		ZP.channels[t] = zapier.NewWebhook(conf.WebhookURL)
	}
	return ZP
}

// TODO: proper constant events
func (s *Zapier) Post(chn string, v interface{}) error {
	if chn, found := s.channels[chn]; found {
		return chn.Post(v)
	}
	return errors.Wrapf(ErrZapierChan, "channel: %s", chn)
}

func (s *Zapier) Put(chn string, v interface{}) error {
	if chn, found := s.channels[chn]; found {
		return chn.Put(v)
	}
	return errors.Wrapf(ErrZapierChan, "channel: %s", chn)
}
