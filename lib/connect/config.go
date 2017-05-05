package connect

type Config struct {
	AppId         string `toml:"app_id"`
	AppSecret     string `toml:"app_secret"`
	OAuthCallback string `toml:"oauth_callback"`
	WebhookURL    string `toml:"webhook_url,omitempty"`
}

type Configs struct {
	Facebook Config `toml:"facebook"`
	Shopify  Config `toml:"shopify"`
}

// Configure loads the connect configs from config file
func Configure(confs Configs) {
	SetupFacebook(confs.Facebook)
	SetupShopify(confs.Shopify)
}
