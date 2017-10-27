package connect

type Config struct {
	AppId         string `toml:"app_id"`
	AppSecret     string `toml:"app_secret"`
	OAuthCallback string `toml:"oauth_callback"`
	WebhookURL    string `toml:"webhook_url,omitempty"`
}

type SlackConfig struct {
	Webhooks map[string]Config `toml:"webhooks"`
}

type Configs struct {
	Facebook Config      `toml:"facebook"`
	Shopify  Config      `toml:"shopify"`
	Stripe   Config      `toml:"stripe"`
	Slack    SlackConfig `toml:"slack"`
}

// Configure loads the connect configs from config file
func Configure(confs Configs) {
	SetupFacebook(confs.Facebook)
	SetupShopify(confs.Shopify)
	SetupStripe(confs.Stripe)
	SetupSlack(confs.Slack)
}
