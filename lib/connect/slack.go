package connect

import (
	"bitbucket.org/moodie-app/moodie-api/lib/slack"
)

type Slack struct {
	// message event -> slack channel/webhook mapping
	channels map[string]*slack.Webhook
}

var (
	SL *Slack
)

func SetupSlack(confs SlackConfig) *Slack {
	SL = &Slack{
		channels: map[string]*slack.Webhook{},
	}
	for t, conf := range confs.Webhooks {
		SL.channels[t] = slack.NewWebhook(conf.WebhookURL)
	}
	return SL
}

// TODO: proper constant events
func (s *Slack) Notify(event, msg string) {
	chn, found := s.channels[event]
	if !found {
		return
	}

	payload := &slack.WebhookPostPayload{
		Text: msg,
	}

	chn.PostMessage(payload)
}
