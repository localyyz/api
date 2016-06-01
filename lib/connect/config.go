package connect

type Config struct {
	AppId     string `toml:"app_id"`
	AppSecret string `toml:"app_secret"`
}

type Configs struct {
	Facebook *Config `toml:"facebook"`
}

// Configure loads the connect configs from config file
func Configure(confs Configs) {
	if confs.Facebook != nil {
		SetupFB(confs.Facebook)
	}
}
