package connect

type Config struct {
	AppName       string `toml:"app_name"`
	AppId         string `toml:"app_id"`
	AppSecret     string `toml:"app_secret"`
	OAuthCallback string `toml:"oauth_callback"`
	WebhookURL    string `toml:"webhook_url,omitempty"`
}

type SlackConfig struct {
	Webhooks map[string]Config `toml:"webhooks"`
}

type NatsConfig struct {
	ServerURL   string                 `toml:"server_url"`
	ClusterID   string                 `toml:"cluster_id"`
	AppName     string                 `toml:"app_name"`
	Durable     bool                   `toml:"durable"`
	Publishers  map[string]NatsSubject `toml:"publishers"`
	Subscribers map[string]NatsSubject `toml:"subscribers"`
}

type Configs struct {
	Facebook Config      `toml:"facebook"`
	Shopify  Config      `toml:"shopify"`
	Stripe   Config      `toml:"stripe"`
	Slack    SlackConfig `toml:"slack"`
	Nats     NatsConfig  `toml:"nats"`
}

// Configure loads the connect configs from config file
func Configure(confs Configs) {
	SetupFacebook(confs.Facebook)
	SetupShopify(confs.Shopify)
	SetupStripe(confs.Stripe)
	SetupSlack(confs.Slack)
	SetupNatsStream(confs.Nats)
}
